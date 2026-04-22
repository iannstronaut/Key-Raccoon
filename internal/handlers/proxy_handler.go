package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/services"
)

type ProxyHandler struct {
	proxyService *services.ProxyService
}

func NewProxyHandler(proxyService *services.ProxyService) *ProxyHandler {
	return &ProxyHandler{proxyService: proxyService}
}

func (h *ProxyHandler) AddProxy(c *fiber.Ctx) error {
	var req struct {
		ProxyURL string `json:"proxy_url" validate:"required"`
		Type     string `json:"type" validate:"required"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	proxy, err := h.proxyService.AddProxy(req.ProxyURL, req.Type, req.Username, req.Password)
	if err != nil {
		return proxyStatusFromError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(proxy)
}

func (h *ProxyHandler) GetAllProxies(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	if limit > 100 {
		limit = 100
	}

	proxies, total, err := h.proxyService.GetAllProxies(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"proxies": proxies,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *ProxyHandler) GetProxy(c *fiber.Ctx) error {
	proxyID, err := parseProxyID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid proxy id",
		})
	}

	proxy, err := h.proxyService.GetProxy(proxyID)
	if err != nil {
		return proxyStatusFromError(c, err)
	}

	return c.JSON(proxy)
}

func (h *ProxyHandler) DeleteProxy(c *fiber.Ctx) error {
	proxyID, err := parseProxyID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid proxy id",
		})
	}

	if err := h.proxyService.DeleteProxy(proxyID); err != nil {
		return proxyStatusFromError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "proxy deleted successfully",
	})
}

func (h *ProxyHandler) TestProxy(c *fiber.Ctx) error {
	var req struct {
		ProxyURL string `json:"proxy_url" validate:"required"`
		Type     string `json:"type" validate:"required"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	isHealthy, err := h.proxyService.TestProxy(req.ProxyURL, req.Type, req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"proxy_url":  req.ProxyURL,
		"is_healthy": isHealthy,
		"status":     "ok",
	})
}

func (h *ProxyHandler) CheckProxyHealth(c *fiber.Ctx) error {
	proxyID, err := parseProxyID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid proxy id",
		})
	}

	if err := h.proxyService.CheckProxyHealth(proxyID); err != nil {
		return proxyStatusFromError(c, err)
	}

	proxy, _ := h.proxyService.GetProxy(proxyID)

	return c.JSON(fiber.Map{
		"proxy_id":   proxyID,
		"status":     proxy.Status,
		"last_check": proxy.LastCheck,
	})
}

func parseProxyID(raw string) (uint, error) {
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func proxyStatusFromError(c *fiber.Ctx, err error) error {
	status := fiber.StatusBadRequest
	if err.Error() == repositories.ErrProxyNotFound.Error() {
		status = fiber.StatusNotFound
	}
	return c.Status(status).JSON(fiber.Map{"error": err.Error()})
}
