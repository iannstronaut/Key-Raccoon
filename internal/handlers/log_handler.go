package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/services"
)

type LogHandler struct {
	logService        *services.LogService
	userAPIKeyService *services.UserAPIKeyService
}

func NewLogHandler(logService *services.LogService, userAPIKeyService *services.UserAPIKeyService) *LogHandler {
	return &LogHandler{
		logService:        logService,
		userAPIKeyService: userAPIKeyService,
	}
}

// GetLogs returns all logs with filters (admin only)
func (h *LogHandler) GetLogs(c *fiber.Ctx) error {
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

	filters := h.parseFilters(c)

	logs, total, err := h.logService.GetLogs(limit, offset, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetStats returns aggregated usage statistics.
// Admin sees all stats (with optional filters). Regular users are scoped to their own data.
func (h *LogHandler) GetStats(c *fiber.Ctx) error {
	filters := h.parseFilters(c)

	// Non-admin users can only see their own stats
	userRole, _ := c.Locals("user_role").(string)
	if userRole != "admin" && userRole != "superadmin" {
		currentUserID, _ := c.Locals("user_id").(uint)
		filters.UserID = currentUserID
	}

	stats, err := h.logService.GetUsageStats(filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(stats)
}

// GetUserLogs returns logs for a specific user (admin or self)
func (h *LogHandler) GetUserLogs(c *fiber.Ctx) error {
	targetUserID, err := strconv.ParseUint(c.Params("userID"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	// Ownership check: user can only see their own logs unless admin
	currentUserID, _ := c.Locals("user_id").(uint)
	userRole, _ := c.Locals("user_role").(string)
	if userRole != "admin" && userRole != "superadmin" && currentUserID != uint(targetUserID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	logs, total, err := h.logService.GetLogsByUser(uint(targetUserID), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetAPIKeyLogs returns logs for a specific API key (admin or owner)
func (h *LogHandler) GetAPIKeyLogs(c *fiber.Ctx) error {
	keyID, err := strconv.ParseUint(c.Params("keyID"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid api key id"})
	}

	// Ownership check
	currentUserID, _ := c.Locals("user_id").(uint)
	userRole, _ := c.Locals("user_role").(string)
	if userRole != "admin" && userRole != "superadmin" {
		apiKey, err := h.userAPIKeyService.GetAPIKey(uint(keyID))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "api key not found"})
		}
		if apiKey.UserID != currentUserID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
		}
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	logs, total, err := h.logService.GetLogsByAPIKey(uint(keyID), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *LogHandler) parseFilters(c *fiber.Ctx) repositories.LogFilters {
	filters := repositories.LogFilters{
		Status: c.Query("status"),
		Model:  c.Query("model"),
	}

	if channelID, err := strconv.ParseUint(c.Query("channel_id"), 10, 32); err == nil {
		filters.ChannelID = uint(channelID)
	}
	if userID, err := strconv.ParseUint(c.Query("user_id"), 10, 32); err == nil {
		filters.UserID = uint(userID)
	}
	if apiKeyID, err := strconv.ParseUint(c.Query("api_key_id"), 10, 32); err == nil {
		filters.APIKeyID = uint(apiKeyID)
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filters.DateFrom = &t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filters.DateTo = &t
		}
	}

	return filters
}
