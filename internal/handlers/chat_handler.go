package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
	"keyraccoon/internal/utils"
	"keyraccoon/pkg/logger"
)

// ChatHandler handles OpenAI-compatible API endpoints.
type ChatHandler struct {
	userAPIKeyService *services.UserAPIKeyService
	channelService    *services.ChannelService
	proxyService      *services.ProxyService
	logService        *services.LogService
	httpClient        *http.Client
	openaiBaseURL     string
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(
	userAPIKeyService *services.UserAPIKeyService,
	channelService *services.ChannelService,
	proxyService *services.ProxyService,
	logService *services.LogService,
) *ChatHandler {
	return &ChatHandler{
		userAPIKeyService: userAPIKeyService,
		channelService:    channelService,
		proxyService:      proxyService,
		logService:        logService,
		httpClient:        http.DefaultClient,
		openaiBaseURL:     "https://api.openai.com/v1",
	}
}

// ChatCompletion handles POST /api/v1/chat/completions.
func (h *ChatHandler) ChatCompletion(c *fiber.Ctx) error {
	startTime := time.Now()

	keyID := c.Locals("api_key_id").(uint)
	userID, _ := c.Locals("user_id").(uint)

	// Get user email from the API key data for log enrichment
	var userEmail string
	if userAPIKey, ok := c.Locals("user_api_key").(*models.UserAPIKey); ok && userAPIKey != nil {
		userEmail = userAPIKey.User.Email
	}

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
	hasToken, err := h.userAPIKeyService.HasTokenAvailable(keyID, inputTokens)
	if err != nil || !hasToken {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "token limit exceeded",
				"type":    "rate_limit_error",
			},
		})
	}

	// Find the correct channel and model for the requested model name
	var selectedChannel *models.Channel
	var selectedModel *models.Model
	requestedModel := strings.ToLower(req.Model)

	for i := range channels {
		channelModels, err := h.channelService.GetChannelModels(channels[i].ID)
		if err != nil {
			continue
		}
		for j := range channelModels {
			if strings.ToLower(channelModels[j].Name) == requestedModel && channelModels[j].IsActive {
				selectedChannel = &channels[i]
				selectedModel = &channelModels[j]
				break
			}
		}
		if selectedChannel != nil {
			break
		}
	}

	// Fallback: if no model match found, use first channel (backward compat)
	if selectedChannel == nil {
		selectedChannel = &channels[0]
		logger.Info(fmt.Sprintf("Model %s not found in channels, using first channel: %s", req.Model, selectedChannel.Name))
	}

	logger.Info(fmt.Sprintf("Using channel: %s (ID: %d, Endpoint: %s)", selectedChannel.Name, selectedChannel.ID, selectedChannel.Endpoint))

	// Check channel budget before proceeding
	hasBudget, err := h.channelService.CheckBudget(selectedChannel.ID)
	if err != nil || !hasBudget {
		h.logRequestAsync(requestLogData{
			KeyID: keyID, UserID: userID, UserEmail: userEmail,
			ChannelID: selectedChannel.ID, ChannelName: selectedChannel.Name,
			ModelName: req.Model, TokenPrice: modelTokenPrice(selectedModel),
			InputTokens: inputTokens, Status: "failed",
			ErrorMsg: "channel budget exceeded", StartTime: startTime, RequestIP: c.IP(),
		})
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "channel budget exceeded",
				"type":    "rate_limit_error",
			},
		})
	}

	// Record input usage and increment usage count
	_ = h.userAPIKeyService.RecordTokenUsage(keyID, inputTokens)
	_ = h.userAPIKeyService.IncrementUsage(keyID)

	// Get channel API keys
	apiKeys, err := h.channelService.GetChannelAPIKeys(selectedChannel.ID)
	if err != nil || len(apiKeys) == 0 {
		h.logRequestAsync(requestLogData{
			KeyID: keyID, UserID: userID, UserEmail: userEmail,
			ChannelID: selectedChannel.ID, ChannelName: selectedChannel.Name,
			ModelName: req.Model, TokenPrice: modelTokenPrice(selectedModel),
			InputTokens: inputTokens, Status: "failed",
			ErrorMsg: "channel api key not available", StartTime: startTime, RequestIP: c.IP(),
		})
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
	respBody, err := h.forwardRequest(selectedChannel, apiKeys[0].APIKey, req, proxy, "/chat/completions")
	if err != nil {
		latencyMs := time.Since(startTime).Milliseconds()
		logger.Error(fmt.Sprintf("forward to channel failed: %v", err))
		h.logRequestAsync(requestLogData{
			KeyID: keyID, UserID: userID, UserEmail: userEmail,
			ChannelID: selectedChannel.ID, ChannelName: selectedChannel.Name,
			ModelName: req.Model, TokenPrice: modelTokenPrice(selectedModel),
			InputTokens: inputTokens, LatencyMs: latencyMs, Status: "failed",
			ErrorMsg: err.Error(), StartTime: startTime, RequestIP: c.IP(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"message": err.Error(),
				"type":    "server_error",
			},
		})
	}

	// Count output tokens from response
	var outputTokens int64
	var responseData map[string]interface{}
	if err := json.Unmarshal(respBody, &responseData); err == nil {
		if usage, ok := responseData["usage"].(map[string]interface{}); ok {
			if ct, ok := usage["completion_tokens"].(float64); ok {
				outputTokens = int64(ct)
				_ = h.userAPIKeyService.RecordTokenUsage(keyID, outputTokens)
			}
			// Use upstream token counts if available
			if pt, ok := usage["prompt_tokens"].(float64); ok {
				inputTokens = int64(pt)
			}
		}
	}

	latencyMs := time.Since(startTime).Milliseconds()

	// Calculate cost based on model's token_price (per 1K tokens)
	var cost float64
	if selectedModel != nil && selectedModel.TokenPrice > 0 {
		cost = float64(inputTokens+outputTokens) / 1000.0 * selectedModel.TokenPrice
	}

	// Record channel budget usage atomically
	if cost > 0 {
		_ = h.channelService.RecordBudgetUsage(selectedChannel.ID, cost)
	}

	// Log the request asynchronously
	h.logRequestAsync(requestLogData{
		KeyID: keyID, UserID: userID, UserEmail: userEmail,
		ChannelID: selectedChannel.ID, ChannelName: selectedChannel.Name,
		ModelName: req.Model, TokenPrice: modelTokenPrice(selectedModel),
		InputTokens: inputTokens, OutputTokens: outputTokens,
		Cost: cost, LatencyMs: latencyMs, Status: "success",
		StartTime: startTime, RequestIP: c.IP(),
	})

	c.Set("Content-Type", "application/json")
	return c.Send(respBody)
}

// requestLogData holds all data needed to create a request log entry
type requestLogData struct {
	KeyID        uint
	UserID       uint
	UserEmail    string
	ChannelID    uint
	ChannelName  string
	ModelName    string
	TokenPrice   float64
	InputTokens  int64
	OutputTokens int64
	Cost         float64
	LatencyMs    int64
	Status       string
	ErrorMsg     string
	StartTime    time.Time
	RequestIP    string
}

// logRequestAsync logs a request asynchronously via the LogService
func (h *ChatHandler) logRequestAsync(data requestLogData) {
	if h.logService == nil {
		return
	}

	log := models.RequestLog{
		UserAPIKeyID: data.KeyID,
		UserID:       data.UserID,
		UserEmail:    data.UserEmail,
		ChannelID:    data.ChannelID,
		ChannelName:  data.ChannelName,
		ModelName:    data.ModelName,
		TokenPrice:   data.TokenPrice,
		InputTokens:  data.InputTokens,
		OutputTokens: data.OutputTokens,
		TotalTokens:  data.InputTokens + data.OutputTokens,
		Cost:         data.Cost,
		Status:       data.Status,
		ErrorMessage: data.ErrorMsg,
		LatencyMs:    data.LatencyMs,
		RequestIP:    data.RequestIP,
		CreatedAt:    data.StartTime,
	}

	go h.logService.LogRequest(log)
}

// modelTokenPrice safely extracts token price from a model pointer
func modelTokenPrice(m *models.Model) float64 {
	if m == nil {
		return 0
	}
	return m.TokenPrice
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

	// Use channel endpoint if available, otherwise use default openaiBaseURL
	baseURL := h.openaiBaseURL
	if channel.Endpoint != "" {
		baseURL = channel.Endpoint
	}

	targetURL := baseURL + endpoint
	httpReq, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(reqBody))
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
func (h *ChatHandler) SetOpenAIBaseURL(u string) {
	h.openaiBaseURL = u
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
