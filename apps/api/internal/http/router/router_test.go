package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/haseg/anifusion-canvas/apps/api/internal/config"
	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
	"github.com/haseg/anifusion-canvas/apps/api/internal/http/handler"
	"github.com/haseg/anifusion-canvas/apps/api/internal/infrastructure/dependency"
	"github.com/haseg/anifusion-canvas/apps/api/internal/usecase"
	"github.com/labstack/echo/v4"
)

func TestStudioWorkflowRoutes(t *testing.T) {
	e := newTestEcho()
	projectID := "http-project"

	createProjectRes := performJSON(t, e, http.MethodPost, "/projects", map[string]any{
		"id":   projectID,
		"name": "HTTP project",
	})
	if createProjectRes.Code != http.StatusCreated {
		t.Fatalf("expected create project status %d, got %d: %s", http.StatusCreated, createProjectRes.Code, createProjectRes.Body.String())
	}

	var projectPayload domain.ProjectResponse
	decodeJSON(t, createProjectRes, &projectPayload)
	if projectPayload.Project.ID != projectID || projectPayload.Project.Name != "HTTP project" {
		t.Fatalf("unexpected created project: %#v", projectPayload.Project)
	}

	updateProjectRes := performJSON(t, e, http.MethodPut, "/projects/"+projectID, map[string]any{
		"name": "Renamed HTTP project",
	})
	if updateProjectRes.Code != http.StatusOK {
		t.Fatalf("expected update project status %d, got %d: %s", http.StatusOK, updateProjectRes.Code, updateProjectRes.Body.String())
	}

	getProjectRes := performJSON(t, e, http.MethodGet, "/projects/"+projectID, nil)
	if getProjectRes.Code != http.StatusOK {
		t.Fatalf("expected get project status %d, got %d: %s", http.StatusOK, getProjectRes.Code, getProjectRes.Body.String())
	}

	generateBody := map[string]any{
		"projectId":         projectID,
		"prompt":            "clean character turn",
		"frameCount":        2,
		"startImageDataUrl": "data:image/png;base64,start",
		"endImageDataUrl":   "data:image/png;base64,end",
	}
	generateRes := performJSON(t, e, http.MethodPost, "/inference/generate", generateBody)
	if generateRes.Code != http.StatusAccepted {
		t.Fatalf("expected generate status %d, got %d: %s", http.StatusAccepted, generateRes.Code, generateRes.Body.String())
	}

	var generatePayload struct {
		Job domain.Job `json:"job"`
	}
	decodeJSON(t, generateRes, &generatePayload)
	job := waitForHTTPJob(t, e, generatePayload.Job.ID)
	if job.Status != domain.JobStatusSucceeded {
		t.Fatalf("expected succeeded generation job, got %s: %s", job.Status, job.Error)
	}

	framesRes := performJSON(t, e, http.MethodGet, "/projects/"+projectID+"/frames", nil)
	if framesRes.Code != http.StatusOK {
		t.Fatalf("expected frames status %d, got %d: %s", http.StatusOK, framesRes.Code, framesRes.Body.String())
	}

	var framesPayload struct {
		Frames []domain.Frame `json:"frames"`
	}
	decodeJSON(t, framesRes, &framesPayload)
	if len(framesPayload.Frames) != 4 {
		t.Fatalf("expected 4 frames including keyframes, got %d", len(framesPayload.Frames))
	}
	if framesPayload.Frames[0].Kind != domain.FrameKindKey || framesPayload.Frames[3].Kind != domain.FrameKindKey {
		t.Fatalf("expected first and last frames to be keyframes")
	}

	target := framesPayload.Frames[1]
	metadataRes := performJSON(t, e, http.MethodPut, "/projects/"+projectID+"/frames/"+target.ID+"/metadata", map[string]any{
		"kind": "edited",
		"note": "metadata cleanup",
	})
	if metadataRes.Code != http.StatusOK {
		t.Fatalf("expected metadata status %d, got %d: %s", http.StatusOK, metadataRes.Code, metadataRes.Body.String())
	}

	reorderRes := performJSON(t, e, http.MethodPut, "/projects/"+projectID+"/frames/reorder", map[string]any{
		"frameIds": []string{
			framesPayload.Frames[1].ID,
			framesPayload.Frames[0].ID,
			framesPayload.Frames[2].ID,
			framesPayload.Frames[3].ID,
		},
	})
	if reorderRes.Code != http.StatusOK {
		t.Fatalf("expected reorder status %d, got %d: %s", http.StatusOK, reorderRes.Code, reorderRes.Body.String())
	}
	var reorderPayload domain.ReorderFramesResult
	decodeJSON(t, reorderRes, &reorderPayload)
	if reorderPayload.Frames[0].ID != target.ID || reorderPayload.Frames[0].Index != 0 {
		t.Fatalf("unexpected reorder result: %#v", reorderPayload.Frames)
	}

	updateRes := performJSON(t, e, http.MethodPut, "/projects/"+projectID+"/frames/"+target.ID, map[string]any{
		"imageDataUrl": "data:image/png;base64,edited",
		"note":         "cleanup pass",
	})
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d: %s", http.StatusOK, updateRes.Code, updateRes.Body.String())
	}

	var updatePayload domain.UpdateFrameResult
	decodeJSON(t, updateRes, &updatePayload)
	if updatePayload.Frame.Kind != domain.FrameKindEdited {
		t.Fatalf("expected edited frame, got %s", updatePayload.Frame.Kind)
	}
	if updatePayload.Frame.ImageURL != "data:image/png;base64,edited" {
		t.Fatalf("expected updated image data URL, got %q", updatePayload.Frame.ImageURL)
	}

	deleteRes := performJSON(t, e, http.MethodDelete, "/projects/"+projectID+"/frames/"+target.ID, nil)
	if deleteRes.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d: %s", http.StatusNoContent, deleteRes.Code, deleteRes.Body.String())
	}
}

func TestStudioUpdateRouteReturnsNotFoundForMissingFrame(t *testing.T) {
	e := newTestEcho()

	res := performJSON(t, e, http.MethodPut, "/projects/missing-project/frames/missing-frame", map[string]any{
		"imageDataUrl": "data:image/png;base64,edited",
	})
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, res.Code, res.Body.String())
	}
	assertErrorResponse(t, res, "Not Found", "frame not found")

	metadataRes := performJSON(t, e, http.MethodPut, "/projects/missing-project/frames/missing-frame/metadata", map[string]any{
		"note": "missing",
	})
	if metadataRes.Code != http.StatusNotFound {
		t.Fatalf("expected metadata status %d, got %d: %s", http.StatusNotFound, metadataRes.Code, metadataRes.Body.String())
	}
	assertErrorResponse(t, metadataRes, "Not Found", "frame not found")

	deleteRes := performJSON(t, e, http.MethodDelete, "/projects/missing-project/frames/missing-frame", nil)
	if deleteRes.Code != http.StatusNotFound {
		t.Fatalf("expected delete status %d, got %d: %s", http.StatusNotFound, deleteRes.Code, deleteRes.Body.String())
	}
	assertErrorResponse(t, deleteRes, "Not Found", "frame not found")
}

func TestStudioUpdateRouteReturnsServerErrorForStorageFailure(t *testing.T) {
	service := usecase.NewStudioServiceWithStoreAndObjects(usecase.NewMemoryStudioStore(), failingObjectStore{})
	e := newTestEchoWithService(service)
	projectID := "storage-failure-project"

	generateRes := performJSON(t, e, http.MethodPost, "/inference/generate", map[string]any{
		"projectId":         projectID,
		"prompt":            "clean character turn",
		"frameCount":        2,
		"startImageDataUrl": "data:image/png;base64,start",
		"endImageDataUrl":   "data:image/png;base64,end",
	})
	if generateRes.Code != http.StatusAccepted {
		t.Fatalf("expected generate status %d, got %d: %s", http.StatusAccepted, generateRes.Code, generateRes.Body.String())
	}
	var generatePayload struct {
		Job domain.Job `json:"job"`
	}
	decodeJSON(t, generateRes, &generatePayload)
	waitForHTTPJob(t, e, generatePayload.Job.ID)

	framesRes := performJSON(t, e, http.MethodGet, "/projects/"+projectID+"/frames", nil)
	var framesPayload struct {
		Frames []domain.Frame `json:"frames"`
	}
	decodeJSON(t, framesRes, &framesPayload)

	res := performJSON(t, e, http.MethodPut, "/projects/"+projectID+"/frames/"+framesPayload.Frames[1].ID, map[string]any{
		"imageDataUrl": "data:image/png;base64,edited",
	})
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d: %s", http.StatusInternalServerError, res.Code, res.Body.String())
	}
	assertErrorResponse(t, res, "Internal Server Error", "store edited frame object: upload failed")
}

func TestStudioProjectRoutesReturnNotFound(t *testing.T) {
	e := newTestEcho()

	getRes := performJSON(t, e, http.MethodGet, "/projects/missing-project", nil)
	if getRes.Code != http.StatusNotFound {
		t.Fatalf("expected get project status %d, got %d: %s", http.StatusNotFound, getRes.Code, getRes.Body.String())
	}
	assertErrorResponse(t, getRes, "Not Found", "project not found")

	updateRes := performJSON(t, e, http.MethodPut, "/projects/missing-project", map[string]any{
		"name": "Missing project",
	})
	if updateRes.Code != http.StatusNotFound {
		t.Fatalf("expected update project status %d, got %d: %s", http.StatusNotFound, updateRes.Code, updateRes.Body.String())
	}
	assertErrorResponse(t, updateRes, "Not Found", "project not found")
}

func TestStudioRoutesRejectInvalidRequests(t *testing.T) {
	e := newTestEcho()

	tests := []struct {
		name   string
		method string
		path   string
		body   map[string]any
	}{
		{
			name:   "create project missing id",
			method: http.MethodPost,
			path:   "/projects",
			body: map[string]any{
				"name": "Demo project",
			},
		},
		{
			name:   "create project missing name",
			method: http.MethodPost,
			path:   "/projects",
			body: map[string]any{
				"id": "project-1",
			},
		},
		{
			name:   "update project missing name",
			method: http.MethodPut,
			path:   "/projects/project-1",
			body:   map[string]any{},
		},
		{
			name:   "generate missing project",
			method: http.MethodPost,
			path:   "/inference/generate",
			body: map[string]any{
				"prompt":            "turn around",
				"frameCount":        2,
				"startImageDataUrl": "data:image/png;base64,start",
				"endImageDataUrl":   "data:image/png;base64,end",
			},
		},
		{
			name:   "generate missing prompt",
			method: http.MethodPost,
			path:   "/inference/generate",
			body: map[string]any{
				"projectId":         "project-1",
				"frameCount":        2,
				"startImageDataUrl": "data:image/png;base64,start",
				"endImageDataUrl":   "data:image/png;base64,end",
			},
		},
		{
			name:   "generate invalid frame count",
			method: http.MethodPost,
			path:   "/inference/generate",
			body: map[string]any{
				"projectId":         "project-1",
				"prompt":            "turn around",
				"frameCount":        1,
				"startImageDataUrl": "data:image/png;base64,start",
				"endImageDataUrl":   "data:image/png;base64,end",
			},
		},
		{
			name:   "generate invalid start image",
			method: http.MethodPost,
			path:   "/inference/generate",
			body: map[string]any{
				"projectId":         "project-1",
				"prompt":            "turn around",
				"frameCount":        2,
				"startImageDataUrl": "https://example.test/start.png",
				"endImageDataUrl":   "data:image/png;base64,end",
			},
		},
		{
			name:   "inpaint missing mask",
			method: http.MethodPost,
			path:   "/inference/inpaint",
			body: map[string]any{
				"projectId": "project-1",
				"frameId":   "frame-1",
				"prompt":    "fix hand",
				"strength":  0.7,
			},
		},
		{
			name:   "inpaint invalid strength",
			method: http.MethodPost,
			path:   "/inference/inpaint",
			body: map[string]any{
				"projectId":   "project-1",
				"frameId":     "frame-1",
				"prompt":      "fix hand",
				"maskDataUrl": "data:image/png;base64,mask",
				"strength":    1.2,
			},
		},
		{
			name:   "update missing image",
			method: http.MethodPut,
			path:   "/projects/project-1/frames/frame-1",
			body:   map[string]any{},
		},
		{
			name:   "metadata update missing fields",
			method: http.MethodPut,
			path:   "/projects/project-1/frames/frame-1/metadata",
			body:   map[string]any{},
		},
		{
			name:   "metadata update invalid kind",
			method: http.MethodPut,
			path:   "/projects/project-1/frames/frame-1/metadata",
			body: map[string]any{
				"kind": "unknown",
			},
		},
		{
			name:   "reorder missing ids",
			method: http.MethodPut,
			path:   "/projects/project-1/frames/reorder",
			body:   map[string]any{},
		},
		{
			name:   "update invalid image",
			method: http.MethodPut,
			path:   "/projects/project-1/frames/frame-1",
			body: map[string]any{
				"imageDataUrl": "https://example.test/edited.png",
			},
		},
		{
			name:   "export invalid fps",
			method: http.MethodPost,
			path:   "/export/video",
			body: map[string]any{
				"projectId": "project-1",
				"fps":       0,
			},
		},
		{
			name:   "export too large fps",
			method: http.MethodPost,
			path:   "/export/video",
			body: map[string]any{
				"projectId": "project-1",
				"fps":       120,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := performJSON(t, e, tt.method, tt.path, tt.body)
			if res.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, res.Code, res.Body.String())
			}
			assertErrorCode(t, res, "Bad Request")
		})
	}
}

func newTestEcho() *echo.Echo {
	return newTestEchoWithService(usecase.NewStudioService())
}

func newTestEchoWithService(service *usecase.StudioService) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = handler.JSONErrorHandler
	studioHandler := handler.NewStudioHandler(service)
	healthHandler := handler.NewHealthHandler(dependency.NewChecker(config.Config{}))
	Register(e, studioHandler, healthHandler)
	return e
}

type failingObjectStore struct{}

func (failingObjectStore) PutDataURL(_ context.Context, key string, dataURL string) (domain.StorageObject, error) {
	if strings.Contains(dataURL, "edited") {
		return domain.StorageObject{}, fmt.Errorf("upload failed")
	}
	return domain.StorageObject{
		Key:         key,
		URL:         "https://assets.example.test/" + key,
		ContentType: "image/png",
		Size:        int64(len(dataURL)),
	}, nil
}

func performJSON(t *testing.T, e *echo.Echo, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			t.Fatalf("encode request body failed: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &payload)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	return res
}

func waitForHTTPJob(t *testing.T, e *echo.Echo, jobID string) domain.Job {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		res := performJSON(t, e, http.MethodGet, "/jobs/"+jobID, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected job status response %d, got %d: %s", http.StatusOK, res.Code, res.Body.String())
		}

		var payload struct {
			Job domain.Job `json:"job"`
		}
		decodeJSON(t, res, &payload)
		if payload.Job.Status == domain.JobStatusSucceeded || payload.Job.Status == domain.JobStatusFailed {
			return payload.Job
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for job %s", jobID)
	return domain.Job{}
}

func decodeJSON(t *testing.T, res *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		t.Fatalf("decode response failed: %v; body: %s", err, res.Body.String())
	}
}

func assertErrorCode(t *testing.T, res *httptest.ResponseRecorder, code string) {
	t.Helper()

	var payload handler.ErrorBody
	decodeJSON(t, res, &payload)
	if payload.Error.Code != code {
		t.Fatalf("expected error code %q, got %q", code, payload.Error.Code)
	}
	if payload.Error.Message == "" {
		t.Fatalf("expected error message to be set")
	}
}

func assertErrorResponse(t *testing.T, res *httptest.ResponseRecorder, code string, message string) {
	t.Helper()

	var payload handler.ErrorBody
	decodeJSON(t, res, &payload)
	if payload.Error.Code != code {
		t.Fatalf("expected error code %q, got %q", code, payload.Error.Code)
	}
	if payload.Error.Message != message {
		t.Fatalf("expected error message %q, got %q", message, payload.Error.Message)
	}
}
