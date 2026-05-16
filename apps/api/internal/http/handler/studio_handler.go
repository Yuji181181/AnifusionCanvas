package handler

import (
	"net/http"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
	"github.com/haseg/anifusion-canvas/apps/api/internal/usecase"
	"github.com/labstack/echo/v4"
)

type StudioHandler struct {
	service *usecase.StudioService
}

func NewStudioHandler(service *usecase.StudioService) *StudioHandler {
	return &StudioHandler{service: service}
}

func (h *StudioHandler) ListFrames(c echo.Context) error {
	projectID := c.Param("projectId")
	return c.JSON(http.StatusOK, map[string]any{"frames": h.service.ListFrames(projectID)})
}

func (h *StudioHandler) GenerateFrames(c echo.Context) error {
	var input domain.GenerateFramesRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusAccepted, map[string]any{"job": h.service.GenerateFrames(input)})
}

func (h *StudioHandler) InpaintFrame(c echo.Context) error {
	var input domain.InpaintFrameRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusAccepted, map[string]any{"job": h.service.InpaintFrame(input)})
}

func (h *StudioHandler) UpdateFrame(c echo.Context) error {
	var input domain.UpdateFrameRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input.ProjectID = c.Param("projectId")
	input.FrameID = c.Param("frameId")
	frame, err := h.service.UpdateFrame(input)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, domain.UpdateFrameResult{Frame: frame})
}

func (h *StudioHandler) ExportVideo(c echo.Context) error {
	var input domain.ExportVideoRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusAccepted, map[string]any{"job": h.service.ExportVideo(input)})
}

func (h *StudioHandler) GetJob(c echo.Context) error {
	job, ok := h.service.GetJob(c.Param("jobId"))
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}

	return c.JSON(http.StatusOK, map[string]any{"job": job})
}
