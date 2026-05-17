package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/haseg/anifusion-canvas/apps/api/internal/usecase"
	"github.com/labstack/echo/v4"
)

func TestStudioHandlerRejectsBlankProjectCreation(t *testing.T) {
	e := echo.New()
	h := NewStudioHandler(usecase.NewStudioService())

	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{"id":"","name":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateProject(c)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestStudioHandlerCreateProject(t *testing.T) {
	e := echo.New()
	h := NewStudioHandler(usecase.NewStudioService())

	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{"id":"p-1","name":"Test"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CreateProject(c); err != nil {
		t.Fatalf("create project failed: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestStudioHandlerRejectsBlankGeneration(t *testing.T) {
	e := echo.New()
	h := NewStudioHandler(usecase.NewStudioService())

	req := httptest.NewRequest(http.MethodPost, "/inference/generate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GenerateFrames(c)
	if err == nil {
		t.Fatalf("expected validation error for blank fields")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestStudioHandlerRejectsInvalidFrameCount(t *testing.T) {
	e := echo.New()
	h := NewStudioHandler(usecase.NewStudioService())

	req := httptest.NewRequest(http.MethodPost, "/inference/generate", strings.NewReader(`{
		"projectId": "p-1",
		"prompt": "test",
		"frameCount": 100,
		"startImageDataUrl": "data:image/png;base64,x",
		"endImageDataUrl": "data:image/png;base64,x"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GenerateFrames(c)
	if err == nil {
		t.Fatalf("expected validation error for frameCount=100")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestStudioHandlerRejectsNonDataURL(t *testing.T) {
	e := echo.New()
	h := NewStudioHandler(usecase.NewStudioService())

	req := httptest.NewRequest(http.MethodPost, "/inference/generate", strings.NewReader(`{
		"projectId": "p-1",
		"prompt": "test",
		"frameCount": 4,
		"startImageDataUrl": "https://example.com/image.png",
		"endImageDataUrl": "https://example.com/image.png"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GenerateFrames(c)
	if err == nil {
		t.Fatalf("expected validation error for non-data URLs")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestStudioHandlerRejectsOversizedDataURL(t *testing.T) {
	e := echo.New()
	h := NewStudioHandler(usecase.NewStudioService())

	req := httptest.NewRequest(http.MethodPut, "/projects/p-1/frames/f-1", strings.NewReader(`{
		"imageDataUrl": "`+"data:image/png;base64,"+strings.Repeat("a", 8*1024*1024)+`"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("projectId", "frameId")
	c.SetParamValues("p-1", "f-1")

	err := h.UpdateFrame(c)
	if err == nil {
		t.Fatalf("expected validation error for oversized data URL")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestStudioHandlerUpdateFrameNotFound(t *testing.T) {
	e := echo.New()
	h := NewStudioHandler(usecase.NewStudioService())

	req := httptest.NewRequest(http.MethodPut, "/projects/p-1/frames/f-missing", strings.NewReader(`{
		"projectId": "p-1",
		"frameId": "f-missing",
		"imageDataUrl": "data:image/png;base64,x",
		"note": "test"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("projectId", "frameId")
	c.SetParamValues("p-1", "f-missing")

	err := h.UpdateFrame(c)
	if err == nil {
		t.Fatalf("expected not found error")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
