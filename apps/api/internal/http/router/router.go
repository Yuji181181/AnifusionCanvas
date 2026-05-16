package router

import (
	"net/http"

	"github.com/haseg/anifusion-canvas/apps/api/internal/http/handler"
	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, studio *handler.StudioHandler, health *handler.HealthHandler) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/health/dependencies", health.Readiness)

	e.POST("/projects", studio.CreateProject)
	e.GET("/projects/:projectId", studio.GetProject)
	e.PUT("/projects/:projectId", studio.UpdateProject)
	e.PUT("/projects/:projectId/frames/reorder", studio.ReorderFrames)
	e.GET("/projects/:projectId/frames", studio.ListFrames)
	e.PUT("/projects/:projectId/frames/:frameId", studio.UpdateFrame)
	e.PUT("/projects/:projectId/frames/:frameId/metadata", studio.UpdateFrameMetadata)
	e.DELETE("/projects/:projectId/frames/:frameId", studio.DeleteFrame)
	e.POST("/inference/generate", studio.GenerateFrames)
	e.POST("/inference/inpaint", studio.InpaintFrame)
	e.POST("/export/video", studio.ExportVideo)
	e.GET("/jobs/:jobId", studio.GetJob)
}
