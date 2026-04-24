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

type ChannelHandler struct {
	channelService *services.ChannelService
}

func NewChannelHandler(channelService *services.ChannelService) *ChannelHandler {
	return &ChannelHandler{channelService: channelService}
}

func (h *ChannelHandler) CreateChannel(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Endpoint    string `json:"endpoint"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	channel, err := h.channelService.CreateChannel(req.Name, req.Type, req.Endpoint, req.Description)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(sanitizeChannel(channel))
}

func (h *ChannelHandler) GetAllChannels(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	channels, total, err := h.channelService.GetAllChannels(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	responseChannels := make([]fiber.Map, 0, len(channels))
	for i := range channels {
		responseChannels = append(responseChannels, sanitizeChannel(&channels[i]))
	}

	return c.JSON(fiber.Map{
		"channels": responseChannels,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *ChannelHandler) GetChannel(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	channel, err := h.channelService.GetChannel(channelID)
	if err != nil {
		if errors.Is(err, repositories.ErrChannelNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(sanitizeChannel(channel))
}

func (h *ChannelHandler) UpdateChannel(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Endpoint    *string  `json:"endpoint"`
		IsActive    *bool    `json:"is_active"`
		Budget      *float64 `json:"budget"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.Endpoint != nil {
		updates["endpoint"] = strings.TrimSpace(*req.Endpoint)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Budget != nil {
		updates["budget"] = *req.Budget
	}

	channel, err := h.channelService.UpdateChannel(channelID, updates)
	if err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(sanitizeChannel(channel))
}

func (h *ChannelHandler) DeleteChannel(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	if err := h.channelService.DeleteChannel(channelID); err != nil {
		if errors.Is(err, repositories.ErrChannelNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "channel deleted successfully"})
}

func (h *ChannelHandler) AddAPIKey(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	var req struct {
		APIKey string `json:"api_key"`
	}

	// Try to parse body, but it's optional
	_ = c.BodyParser(&req)

	var apiKey *models.ChannelAPIKey
	if req.APIKey != "" {
		// Add with specific value
		apiKey, err = h.channelService.AddAPIKeyWithValue(channelID, req.APIKey)
	} else {
		// Auto-generate
		apiKey, err = h.channelService.AddAPIKey(channelID)
	}

	if err != nil {
		return channelStatusFromError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(apiKey)
}

func (h *ChannelHandler) GetChannelAPIKeys(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	apiKeys, err := h.channelService.GetChannelAPIKeys(channelID)
	if err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{
		"api_keys": apiKeys,
		"total":    len(apiKeys),
	})
}

func (h *ChannelHandler) DeleteAPIKey(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	keyID, err := strconv.ParseUint(c.Params("keyID"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	// Verify channel exists first
	if _, err := h.channelService.GetChannel(channelID); err != nil {
		return channelStatusFromError(c, err)
	}

	if err := h.channelService.RemoveAPIKey(uint(keyID)); err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{"message": "api key deleted successfully"})
}

func (h *ChannelHandler) RotateAPIKey(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	newKey, err := h.channelService.RotateAPIKey(channelID)
	if err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "api key rotated successfully",
		"new_key": newKey,
	})
}

func (h *ChannelHandler) AddModel(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	var req struct {
		Name         string  `json:"name"`
		DisplayName  string  `json:"display_name"`
		TokenPrice   float64 `json:"token_price"`
		SystemPrompt string  `json:"system_prompt"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	model, err := h.channelService.AddModel(channelID, req.Name, req.DisplayName, req.TokenPrice, req.SystemPrompt)
	if err != nil {
		return channelStatusFromError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(model)
}

func (h *ChannelHandler) GetChannelModels(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	modelsList, err := h.channelService.GetChannelModels(channelID)
	if err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{
		"models": modelsList,
		"total":  len(modelsList),
	})
}

func (h *ChannelHandler) UpdateModel(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	modelID, err := strconv.ParseUint(c.Params("modelID"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid model id"})
	}

	// Verify channel exists
	channel, err := h.channelService.GetChannel(channelID)
	if err != nil {
		return channelStatusFromError(c, err)
	}

	// Verify model belongs to this channel
	modelBelongs := false
	for _, m := range channel.Models {
		if m.ID == uint(modelID) {
			modelBelongs = true
			break
		}
	}
	if !modelBelongs {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "model not found in this channel"})
	}

	var req struct {
		DisplayName  *string  `json:"display_name"`
		TokenPrice   *float64 `json:"token_price"`
		SystemPrompt *string  `json:"system_prompt"`
		IsActive     *bool    `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	updates := make(map[string]any)
	if req.DisplayName != nil {
		updates["display_name"] = strings.TrimSpace(*req.DisplayName)
	}
	if req.TokenPrice != nil {
		updates["token_price"] = *req.TokenPrice
	}
	if req.SystemPrompt != nil {
		updates["system_prompt"] = strings.TrimSpace(*req.SystemPrompt)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no fields to update"})
	}

	model, err := h.channelService.UpdateModel(uint(modelID), updates)
	if err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{
		"id":            model.ID,
		"channel_id":    model.ChannelID,
		"name":          model.Name,
		"display_name":  model.DisplayName,
		"is_active":     model.IsActive,
		"token_price":   model.TokenPrice,
		"system_prompt": model.SystemPrompt,
		"created_at":    model.CreatedAt,
		"updated_at":    model.UpdatedAt,
	})
}

func (h *ChannelHandler) DeleteModel(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	modelID, err := strconv.ParseUint(c.Params("modelID"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid model id"})
	}

	// Verify channel exists first
	if _, err := h.channelService.GetChannel(channelID); err != nil {
		return channelStatusFromError(c, err)
	}

	if err := h.channelService.DeleteModel(uint(modelID)); err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{"message": "model deleted successfully"})
}

func (h *ChannelHandler) GetChannelUsers(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	users, err := h.channelService.GetChannelUsers(channelID)
	if err != nil {
		return channelStatusFromError(c, err)
	}

	response := make([]fiber.Map, 0, len(users))
	for _, u := range users {
		response = append(response, fiber.Map{
			"id":        u.ID,
			"email":     u.Email,
			"name":      u.Name,
			"role":      u.Role,
			"is_active": u.IsActive,
		})
	}

	return c.JSON(fiber.Map{
		"users": response,
		"total": len(response),
	})
}

func (h *ChannelHandler) BindUserToChannel(c *fiber.Ctx) error {
	channelID, userID, err := parseChannelAndUserID(c)
	if err != nil {
		return err
	}

	if err := h.channelService.BindUserToChannel(userID, channelID); err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{"message": "user bound to channel successfully"})
}

func (h *ChannelHandler) UnbindUserFromChannel(c *fiber.Ctx) error {
	channelID, userID, err := parseChannelAndUserID(c)
	if err != nil {
		return err
	}

	if err := h.channelService.UnbindUserFromChannel(userID, channelID); err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{"message": "user unbound from channel successfully"})
}

func parseChannelID(raw string) (uint, error) {
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func parseChannelAndUserID(c *fiber.Ctx) (uint, uint, error) {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return 0, 0, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	userValue, err := strconv.ParseUint(c.Params("userID"), 10, 32)
	if err != nil {
		return 0, 0, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	return channelID, uint(userValue), nil
}

func (h *ChannelHandler) ResetBudget(c *fiber.Ctx) error {
	channelID, err := parseChannelID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid channel id"})
	}

	if err := h.channelService.ResetBudgetUsed(channelID); err != nil {
		return channelStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{"message": "budget usage reset successfully"})
}

func sanitizeChannel(channel *models.Channel) fiber.Map {
	apiKeys := make([]fiber.Map, 0, len(channel.APIKeys))
	for i := range channel.APIKeys {
		apiKeys = append(apiKeys, fiber.Map{
			"id":         channel.APIKeys[i].ID,
			"channel_id": channel.APIKeys[i].ChannelID,
			"api_key":    channel.APIKeys[i].APIKey,
			"is_active":  channel.APIKeys[i].IsActive,
			"created_at": channel.APIKeys[i].CreatedAt,
			"updated_at": channel.APIKeys[i].UpdatedAt,
		})
	}

	modelsList := make([]fiber.Map, 0, len(channel.Models))
	for i := range channel.Models {
		modelsList = append(modelsList, fiber.Map{
			"id":            channel.Models[i].ID,
			"channel_id":    channel.Models[i].ChannelID,
			"name":          channel.Models[i].Name,
			"display_name":  channel.Models[i].DisplayName,
			"is_active":     channel.Models[i].IsActive,
			"token_price":   channel.Models[i].TokenPrice,
			"system_prompt": channel.Models[i].SystemPrompt,
			"created_at":    channel.Models[i].CreatedAt,
			"updated_at":    channel.Models[i].UpdatedAt,
		})
	}

	return fiber.Map{
		"id":          channel.ID,
		"name":        channel.Name,
		"type":        channel.Type,
		"endpoint":    channel.Endpoint,
		"is_active":   channel.IsActive,
		"description": channel.Description,
		"budget":      channel.Budget,
		"budget_used": channel.BudgetUsed,
		"created_at":  channel.CreatedAt,
		"updated_at":  channel.UpdatedAt,
		"api_keys":    apiKeys,
		"models":      modelsList,
	}
}

func channelStatusFromError(c *fiber.Ctx, err error) error {
	status := fiber.StatusBadRequest
	switch {
	case errors.Is(err, repositories.ErrChannelNotFound),
		errors.Is(err, repositories.ErrModelNotFound),
		errors.Is(err, repositories.ErrUserNotFound),
		errors.Is(err, repositories.ErrChannelAPIKeyNotFound):
		status = fiber.StatusNotFound
	}

	return c.Status(status).JSON(fiber.Map{"error": err.Error()})
}
