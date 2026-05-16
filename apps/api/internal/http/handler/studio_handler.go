package handler

import (
	"net/http"
	"strings"

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
	if isBlank(projectID) {
		return echo.NewHTTPError(http.StatusBadRequest, "projectId is required")
	}

	frames, err := h.service.ListFrames(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"frames": frames})
}

func (h *StudioHandler) GenerateFrames(c echo.Context) error {
	var input domain.GenerateFramesRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if isBlank(input.ProjectID) {
		return echo.NewHTTPError(http.StatusBadRequest, "projectId is required")
	}
	if isBlank(input.Prompt) {
		return echo.NewHTTPError(http.StatusBadRequest, "prompt is required")
	}

	return c.JSON(http.StatusAccepted, map[string]any{"job": h.service.GenerateFrames(input)})
}

func (h *StudioHandler) InpaintFrame(c echo.Context) error {
	var input domain.InpaintFrameRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if isBlank(input.ProjectID) {
		return echo.NewHTTPError(http.StatusBadRequest, "projectId is required")
	}
	if isBlank(input.FrameID) {
		return echo.NewHTTPError(http.StatusBadRequest, "frameId is required")
	}
	if isBlank(input.Prompt) {
		return echo.NewHTTPError(http.StatusBadRequest, "prompt is required")
	}
	if isBlank(input.MaskDataURL) {
		return echo.NewHTTPError(http.StatusBadRequest, "maskDataUrl is required")
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
	if isBlank(input.ProjectID) {
		return echo.NewHTTPError(http.StatusBadRequest, "projectId is required")
	}
	if isBlank(input.FrameID) {
		return echo.NewHTTPError(http.StatusBadRequest, "frameId is required")
	}
	if isBlank(input.ImageDataURL) {
		return echo.NewHTTPError(http.StatusBadRequest, "imageDataUrl is required")
	}

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
	if isBlank(input.ProjectID) {
		return echo.NewHTTPError(http.StatusBadRequest, "projectId is required")
	}
	if input.FPS <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "fps must be greater than 0")
	}

	return c.JSON(http.StatusAccepted, map[string]any{"job": h.service.ExportVideo(input)})
}

func (h *StudioHandler) GetJob(c echo.Context) error {
	jobID := c.Param("jobId")
	if isBlank(jobID) {
		return echo.NewHTTPError(http.StatusBadRequest, "jobId is required")
	}

	job, ok, err := h.service.GetJob(jobID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}

	return c.JSON(http.StatusOK, map[string]any{"job": job})
}

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}
