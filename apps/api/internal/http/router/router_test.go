package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
}

func TestStudioUpdateRouteReturnsNotFoundForMissingFrame(t *testing.T) {
	e := newTestEcho()

	res := performJSON(t, e, http.MethodPut, "/projects/missing-project/frames/missing-frame", map[string]any{
		"imageDataUrl": "data:image/png;base64,edited",
	})
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, res.Code, res.Body.String())
	}
}

func newTestEcho() *echo.Echo {
	e := echo.New()
	service := usecase.NewStudioService()
	studioHandler := handler.NewStudioHandler(service)
	healthHandler := handler.NewHealthHandler(dependency.NewChecker(config.Config{}))
	Register(e, studioHandler, healthHandler)
	return e
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
