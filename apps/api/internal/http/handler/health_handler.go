package handler

import (
	"net/http"

	"github.com/haseg/anifusion-canvas/apps/api/internal/infrastructure/dependency"
	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	checker *dependency.Checker
}

func NewHealthHandler(checker *dependency.Checker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

func (h *HealthHandler) Readiness(c echo.Context) error {
	results := h.checker.CheckAll(c.Request().Context())
	status := "ok"
	for _, result := range results {
		if result.Status == dependency.StatusError {
			status = "degraded"
			break
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":  status,
		"results": results,
	})
}
