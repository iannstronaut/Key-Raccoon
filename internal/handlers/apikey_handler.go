package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
)

type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

// CreateAPIKey handles POST /api-keys
func (h *APIKeyHandler) CreateAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req struct {
		Name        string  `json:"name"`
		TokenLimit  int64   `json:"token_limit"`
		CreditLimit float64 `json:"credit_limit"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	apiKey, err := h.apiKeyService.CreateAPIKey(
		userID,
		strings.TrimSpace(req.Name),
		req.TokenLimit,
		req.CreditLimit,
	)
	if err != nil {
		return apiKeyStatusFromError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(sanitizeAPIKey(apiKey))
}

// GetUserAPIKeys handles GET /api-keys
func (h *APIKeyHandler) GetUserAPIKeys(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	apiKeys, err := h.apiKeyService.GetUserAPIKeys(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	responseKeys := make([]fiber.Map, 0, len(apiKeys))
	for i := range apiKeys {
		responseKeys = append(responseKeys, sanitizeAPIKey(&apiKeys[i]))
	}

	return c.JSON(fiber.Map{
		"api_keys": responseKeys,
		"total":    len(apiKeys),
	})
}

// GetAPIKey handles GET /api-keys/:id
func (h *APIKeyHandler) GetAPIKey(c *fiber.Ctx) error {
	keyID, err := parseAPIKeyID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid api key id",
		})
	}

	apiKey, err := h.apiKeyService.GetAPIKeyByID(keyID)
	if err != nil {
		return apiKeyStatusFromError(c, err)
	}

	return c.JSON(sanitizeAPIKey(apiKey))
}

// UpdateAPIKey handles PUT /api-keys/:id
func (h *APIKeyHandler) UpdateAPIKey(c *fiber.Ctx) error {
	keyID, err := parseAPIKeyID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid api key id",
		})
	}

	var req struct {
		Name        *string  `json:"name"`
		IsActive    *bool    `json:"is_active"`
		TokenLimit  *int64   `json:"token_limit"`
		CreditLimit *float64 `json:"credit_limit"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.TokenLimit != nil {
		updates["token_limit"] = *req.TokenLimit
	}
	if req.CreditLimit != nil {
		updates["credit_limit"] = *req.CreditLimit
	}

	apiKey, err := h.apiKeyService.UpdateAPIKey(keyID, updates)
	if err != nil {
		return apiKeyStatusFromError(c, err)
	}

	return c.JSON(sanitizeAPIKey(apiKey))
}

// GetAPIKeyUsage handles GET /api-keys/:id/usage
func (h *APIKeyHandler) GetAPIKeyUsage(c *fiber.Ctx) error {
	keyID, err := parseAPIKeyID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid api key id",
		})
	}

	usage, err := h.apiKeyService.GetRealtimeUsage(keyID)
	if err != nil {
		return apiKeyStatusFromError(c, err)
	}

	return c.JSON(usage)
}

// DeleteAPIKey handles DELETE /api-keys/:id
func (h *APIKeyHandler) DeleteAPIKey(c *fiber.Ctx) error {
	keyID, err := parseAPIKeyID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid api key id",
		})
	}

	if err := h.apiKeyService.DeleteAPIKey(keyID); err != nil {
		return apiKeyStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "api key deleted successfully",
	})
}

// BindChannel handles POST /api-keys/:id/channels/:channelID/bind
func (h *APIKeyHandler) BindChannel(c *fiber.Ctx) error {
	keyID, channelID, err := parseAPIKeyAndChannelID(c)
	if err != nil {
		return err
	}

	if err := h.apiKeyService.BindChannel(keyID, channelID); err != nil {
		return apiKeyStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "channel bound successfully",
	})
}

// UnbindChannel handles DELETE /api-keys/:id/channels/:channelID/bind
func (h *APIKeyHandler) UnbindChannel(c *fiber.Ctx) error {
	keyID, channelID, err := parseAPIKeyAndChannelID(c)
	if err != nil {
		return err
	}

	if err := h.apiKeyService.UnbindChannel(keyID, channelID); err != nil {
		return apiKeyStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "channel unbound successfully",
	})
}

func parseAPIKeyID(raw string) (uint, error) {
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func parseAPIKeyAndChannelID(c *fiber.Ctx) (uint, uint, error) {
	keyID, err := parseAPIKeyID(c.Params("id"))
	if err != nil {
		return 0, 0, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid api key id",
		})
	}

	channelValue, err := strconv.ParseUint(c.Params("channelID"), 10, 32)
	if err != nil {
		return 0, 0, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid channel id",
		})
	}

	return keyID, uint(channelValue), nil
}

func sanitizeAPIKey(apiKey *models.APIKey) fiber.Map {
	channels := make([]fiber.Map, 0, len(apiKey.Channels))
	for i := range apiKey.Channels {
		channels = append(channels, fiber.Map{
			"id":   apiKey.Channels[i].ID,
			"name": apiKey.Channels[i].Name,
			"type": apiKey.Channels[i].Type,
		})
	}

	return fiber.Map{
		"id":           apiKey.ID,
		"user_id":      apiKey.UserID,
		"key":          apiKey.Key,
		"name":         apiKey.Name,
		"token_limit":  apiKey.TokenLimit,
		"credit_limit": apiKey.CreditLimit,
		"token_used":   apiKey.TokenUsed,
		"credit_used":  apiKey.CreditUsed,
		"is_active":    apiKey.IsActive,
		"last_used":    apiKey.LastUsed,
		"created_at":   apiKey.CreatedAt,
		"updated_at":   apiKey.UpdatedAt,
		"channels":     channels,
	}
}

func apiKeyStatusFromError(c *fiber.Ctx, err error) error {
	status := fiber.StatusBadRequest
	if errors.Is(err, repositories.ErrAPIKeyNotFound) ||
		errors.Is(err, repositories.ErrUserNotFound) ||
		errors.Is(err, repositories.ErrChannelNotFound) {
		status = fiber.StatusNotFound
	}
	return c.Status(status).JSON(fiber.Map{"error": err.Error()})
}
