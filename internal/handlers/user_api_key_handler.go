package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
)

type UserAPIKeyHandler struct {
	service        *services.UserAPIKeyService
	channelService *services.ChannelService
}

func NewUserAPIKeyHandler(service *services.UserAPIKeyService, channelService *services.ChannelService) *UserAPIKeyHandler {
	return &UserAPIKeyHandler{service: service, channelService: channelService}
}

func (h *UserAPIKeyHandler) CreateAPIKey(c *fiber.Ctx) error {
	var req struct {
		UserID     uint     `json:"user_id"`
		Name       string   `json:"name"`
		UsageLimit int64    `json:"usage_limit"`
		ExpiresAt  *string  `json:"expires_at"`
		ChannelIDs []uint   `json:"channel_ids"`
		ModelIDs   []uint   `json:"model_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid expires_at format, use RFC3339"})
		}
		expiresAt = &t
	}

	apiKey, err := h.service.CreateAPIKey(
		req.UserID,
		req.Name,
		req.UsageLimit,
		expiresAt,
		req.ChannelIDs,
		req.ModelIDs,
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(sanitizeUserAPIKey(apiKey))
}

func (h *UserAPIKeyHandler) GetAllAPIKeys(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	apiKeys, total, err := h.service.GetAllAPIKeys(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	response := make([]fiber.Map, 0, len(apiKeys))
	for i := range apiKeys {
		response = append(response, sanitizeUserAPIKey(&apiKeys[i]))
	}

	return c.JSON(fiber.Map{
		"api_keys": response,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *UserAPIKeyHandler) GetAPIKey(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	apiKey, err := h.service.GetAPIKey(id)
	if err != nil {
		if err == repositories.ErrUserAPIKeyNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Ownership check: non-admin users can only view their own keys
	currentUserID, _ := c.Locals("user_id").(uint)
	userRole, _ := c.Locals("user_role").(string)
	if userRole != "admin" && userRole != "superadmin" && apiKey.UserID != currentUserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	return c.JSON(sanitizeUserAPIKey(apiKey))
}

func (h *UserAPIKeyHandler) GetUserAPIKeys(c *fiber.Ctx) error {
	userID, err := parseID(c.Params("userID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	// Ownership check: non-admin users can only view their own keys
	currentUserID, _ := c.Locals("user_id").(uint)
	userRole, _ := c.Locals("user_role").(string)
	if userRole != "admin" && userRole != "superadmin" && userID != currentUserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	apiKeys, err := h.service.GetUserAPIKeys(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	response := make([]fiber.Map, 0, len(apiKeys))
	for i := range apiKeys {
		response = append(response, sanitizeUserAPIKey(&apiKeys[i]))
	}

	return c.JSON(fiber.Map{
		"api_keys": response,
		"total":    len(apiKeys),
	})
}

func (h *UserAPIKeyHandler) UpdateAPIKey(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	var req struct {
		Name       *string `json:"name"`
		IsActive   *bool   `json:"is_active"`
		UsageLimit *int64  `json:"usage_limit"`
		ExpiresAt  *string `json:"expires_at"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.UsageLimit != nil {
		updates["usage_limit"] = *req.UsageLimit
	}
	if req.ExpiresAt != nil {
		if *req.ExpiresAt == "" {
			updates["expires_at"] = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid expires_at format"})
			}
			updates["expires_at"] = &t
		}
	}

	apiKey, err := h.service.UpdateAPIKey(id, updates)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(sanitizeUserAPIKey(apiKey))
}

func (h *UserAPIKeyHandler) DeleteAPIKey(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	if err := h.service.DeleteAPIKey(id); err != nil {
		if err == repositories.ErrUserAPIKeyNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "api key deleted successfully"})
}

func (h *UserAPIKeyHandler) AddChannel(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	var req struct {
		ChannelID uint `json:"channel_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.service.AddChannel(id, req.ChannelID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "channel added successfully"})
}

func (h *UserAPIKeyHandler) RemoveChannel(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	channelID, err := parseID(c.Params("channelID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	if err := h.service.RemoveChannel(id, channelID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "channel removed successfully"})
}

func (h *UserAPIKeyHandler) AddModel(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	var req struct {
		ModelID uint `json:"model_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.service.AddModel(id, req.ModelID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "model added successfully"})
}

func (h *UserAPIKeyHandler) RemoveModel(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	modelID, err := parseID(c.Params("modelID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid model id"})
	}

	if err := h.service.RemoveModel(id, modelID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "model removed successfully"})
}

func (h *UserAPIKeyHandler) CreateSelfAPIKey(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req struct {
		Name       string  `json:"name"`
		ChannelIDs []uint  `json:"channel_ids"`
		ModelIDs   []uint  `json:"model_ids"`
		TokenLimit int64   `json:"token_limit"`
		ExpiresAt  *string `json:"expires_at"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid expires_at format, use RFC3339"})
		}
		expiresAt = &t
	}

	apiKey, err := h.service.CreateSelfServiceAPIKey(
		userID,
		req.Name,
		req.ChannelIDs,
		req.ModelIDs,
		req.TokenLimit,
		expiresAt,
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(sanitizeUserAPIKey(apiKey))
}

func (h *UserAPIKeyHandler) DeleteSelfAPIKey(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	if err := h.service.DeleteSelfAPIKey(id, userID); err != nil {
		if err == repositories.ErrUserAPIKeyNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "api key deleted successfully"})
}

func (h *UserAPIKeyHandler) GetMyChannels(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	channels, err := h.channelService.GetUserChannelsWithModels(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Sanitize: do NOT include API keys for regular users
	response := make([]fiber.Map, 0, len(channels))
	for _, ch := range channels {
		modelsList := make([]fiber.Map, 0, len(ch.Models))
		for _, m := range ch.Models {
			modelsList = append(modelsList, fiber.Map{
				"id":           m.ID,
				"channel_id":   m.ChannelID,
				"name":         m.Name,
				"display_name": m.DisplayName,
				"is_active":    m.IsActive,
				"token_price":  m.TokenPrice,
			})
		}
		response = append(response, fiber.Map{
			"id":          ch.ID,
			"name":        ch.Name,
			"type":        ch.Type,
			"endpoint":    ch.Endpoint,
			"is_active":   ch.IsActive,
			"description": ch.Description,
			"budget":      ch.Budget,
			"budget_used": ch.BudgetUsed,
			"models":      modelsList,
		})
	}

	return c.JSON(fiber.Map{
		"channels": response,
		"total":    len(response),
	})
}

func parseID(param string) (uint, error) {
	value, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func sanitizeUserAPIKey(apiKey *models.UserAPIKey) fiber.Map {
	channels := make([]fiber.Map, 0, len(apiKey.Channels))
	for i := range apiKey.Channels {
		channels = append(channels, fiber.Map{
			"id":   apiKey.Channels[i].ID,
			"name": apiKey.Channels[i].Name,
			"type": apiKey.Channels[i].Type,
		})
	}

	models := make([]fiber.Map, 0, len(apiKey.Models))
	for i := range apiKey.Models {
		models = append(models, fiber.Map{
			"id":           apiKey.Models[i].Model.ID,
			"name":         apiKey.Models[i].Model.Name,
			"display_name": apiKey.Models[i].Model.DisplayName,
		})
	}

	result := fiber.Map{
		"id":           apiKey.ID,
		"user_id":      apiKey.UserID,
		"name":         apiKey.Name,
		"key":          apiKey.Key,
		"is_active":    apiKey.IsActive,
		"usage_limit":  apiKey.UsageLimit,
		"usage_count":  apiKey.UsageCount,
		"expires_at":   apiKey.ExpiresAt,
		"last_used_at": apiKey.LastUsedAt,
		"created_at":   apiKey.CreatedAt,
		"updated_at":   apiKey.UpdatedAt,
		"channels":     channels,
		"models":       models,
	}

	if apiKey.User.ID != 0 {
		result["user"] = fiber.Map{
			"id":    apiKey.User.ID,
			"email": apiKey.User.Email,
			"name":  apiKey.User.Name,
		}
	}

	return result
}
