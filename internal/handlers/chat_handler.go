package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
	"keyraccoon/internal/utils"
	"keyraccoon/pkg/logger"
)

// ChatHandler handles OpenAI-compatible API endpoints.
type ChatHandler struct {
	apiKeyService  *services.APIKeyService
	channelService *services.ChannelService
	proxyService   *services.ProxyService
	httpClient     *http.Client
	openaiBaseURL  string
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(
	apiKeyService *services.APIKeyService,
	channelService *services.ChannelService,
	proxyService *services.ProxyService,
) *ChatHandler {
	return &ChatHandler{
		apiKeyService:  apiKeyService,
		channelService: channelService,
		proxyService:   proxyService,
		httpClient:     http.DefaultClient,
		openaiBaseURL:  "https://api.openai.com/v1",
	}
}

// ChatCompletion handles POST /api/v1/chat/completions.
func (h *ChatHandler) ChatCompletion(c *fiber.Ctx) error {
	keyID := c.Locals("api_key_id").(uint)
	channels, ok := c.Locals("api_key_channels").([]models.Channel)
	if !ok || len(channels) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "no channel assigned to this api key",
				"type":    "invalid_request_error",
			},
		})
	}

	var req struct {
		Model       string        `json:"model"`
		Messages    []interface{} `json:"messages"`
		Temperature float32       `json:"temperature"`
		MaxTokens   int           `json:"max_tokens"`
		Stream      bool          `json:"stream"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "invalid request body",
				"type":    "invalid_request_error",
			},
		})
	}

	if req.Model == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "model is required",
				"type":    "invalid_request_error",
			},
		})
	}

	if len(req.Messages) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "messages is required",
				"type":    "invalid_request_error",
			},
		})
	}

	// Count input tokens
	msgBytes, _ := json.Marshal(req.Messages)
	inputTokens := utils.CountTokens(string(msgBytes))

	// Check token availability
	hasToken, err := h.apiKeyService.HasTokenAvailable(keyID, inputTokens)
	if err != nil || !hasToken {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "token limit exceeded",
				"type":    "rate_limit_error",
			},
		})
	}

	// Record input usage
	_ = h.apiKeyService.RecordTokenUsage(keyID, inputTokens)

	// Select first channel (simple round-robin or first-available)
	channel := &channels[0]

	// Get channel API keys
	apiKeys, err := h.channelService.GetChannelAPIKeys(channel.ID)
	if err != nil || len(apiKeys) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "channel api key not available",
				"type":    "server_error",
			},
		})
	}

	// Get healthy proxy (optional)
	proxy, _ := h.proxyService.GetHealthyProxy()

	// Forward request
	respBody, err := h.forwardRequest(channel, apiKeys[0].APIKey, req, proxy, "/chat/completions")
	if err != nil {
		logger.Error(fmt.Sprintf("forward to channel failed: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"message": err.Error(),
				"type":    "server_error",
			},
		})
	}

	// Count output tokens from response
	var responseData map[string]interface{}
	if err := json.Unmarshal(respBody, &responseData); err == nil {
		if usage, ok := responseData["usage"].(map[string]interface{}); ok {
			if outputTokens, ok := usage["completion_tokens"].(float64); ok {
				_ = h.apiKeyService.RecordTokenUsage(keyID, int64(outputTokens))
			}
		}
	}

	c.Set("Content-Type", "application/json")
	return c.Send(respBody)
}

// Embeddings handles POST /api/v1/embeddings.
func (h *ChatHandler) Embeddings(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"object": "list",
		"data":   []interface{}{},
		"model":  "",
		"usage": fiber.Map{
			"prompt_tokens": 0,
			"total_tokens":  0,
		},
	})
}

// ListModels handles GET /api/v1/models.
func (h *ChatHandler) ListModels(c *fiber.Ctx) error {
	channels, ok := c.Locals("api_key_channels").([]models.Channel)
	if !ok {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"object": "list",
			"data":   []interface{}{},
		})
	}

	var modelList []fiber.Map
	for _, channel := range channels {
		channelModels, err := h.channelService.GetChannelModels(channel.ID)
		if err != nil {
			continue
		}
		for _, m := range channelModels {
			modelList = append(modelList, fiber.Map{
				"id":         m.Name,
				"object":     "model",
				"owned_by":   channel.Name,
				"permission": []interface{}{},
			})
		}
	}

	return c.JSON(fiber.Map{
		"object": "list",
		"data":   modelList,
	})
}

// forwardRequest forwards the request to the upstream API.
func (h *ChatHandler) forwardRequest(
	channel *models.Channel,
	apiKey string,
	req interface{},
	proxy *models.Proxy,
	endpoint string,
) ([]byte, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := h.openaiBaseURL + endpoint
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := h.buildHTTPClient(proxy)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream error: status %d, body %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// SetOpenAIBaseURL sets the base URL for OpenAI API requests (useful for testing).
func (h *ChatHandler) SetOpenAIBaseURL(url string) {
	h.openaiBaseURL = url
}

func (h *ChatHandler) buildHTTPClient(proxy *models.Proxy) *http.Client {
	if proxy == nil || proxy.URL == "" {
		return h.httpClient
	}

	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		return h.httpClient
	}

	if proxy.Username != "" && proxy.Password != "" {
		proxyURL.User = url.UserPassword(proxy.Username, proxy.Password)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 5 * time.Second,
		}).DialContext,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}
