package db

import (
	"testing"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

func TestDecodeJobResultUsesTypedResults(t *testing.T) {
	tests := []struct {
		name    string
		jobType domain.JobType
		payload string
		assert  func(t *testing.T, result any)
	}{
		{
			name:    "generation",
			jobType: domain.JobTypeGeneration,
			payload: `{"frames":[{"id":"frame-1","projectId":"project-1","index":0,"imageUrl":"data:image/png;base64,a","thumbnailUrl":"data:image/png;base64,a","kind":"generated","updatedAt":"2026-05-17T00:00:00Z"}]}`,
			assert: func(t *testing.T, result any) {
				t.Helper()
				typed, ok := result.(domain.GenerateFramesResult)
				if !ok {
					t.Fatalf("expected GenerateFramesResult, got %T", result)
				}
				if len(typed.Frames) != 1 || typed.Frames[0].Kind != domain.FrameKindGenerated {
					t.Fatalf("unexpected generation result: %#v", typed)
				}
			},
		},
		{
			name:    "inpainting",
			jobType: domain.JobTypeInpainting,
			payload: `{"frame":{"id":"frame-1","projectId":"project-1","index":0,"imageUrl":"data:image/png;base64,a","thumbnailUrl":"data:image/png;base64,a","kind":"inpainted","updatedAt":"2026-05-17T00:00:00Z"}}`,
			assert: func(t *testing.T, result any) {
				t.Helper()
				typed, ok := result.(domain.InpaintFrameResult)
				if !ok {
					t.Fatalf("expected InpaintFrameResult, got %T", result)
				}
				if typed.Frame.Kind != domain.FrameKindInpainted {
					t.Fatalf("unexpected inpainting result: %#v", typed)
				}
			},
		},
		{
			name:    "export",
			jobType: domain.JobTypeExport,
			payload: `{"videoUrl":"https://example.test/export.mp4"}`,
			assert: func(t *testing.T, result any) {
				t.Helper()
				typed, ok := result.(domain.ExportVideoResult)
				if !ok {
					t.Fatalf("expected ExportVideoResult, got %T", result)
				}
				if typed.VideoURL != "https://example.test/export.mp4" {
					t.Fatalf("unexpected export result: %#v", typed)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeJobResult(tt.jobType, []byte(tt.payload))
			if err != nil {
				t.Fatalf("decode job result failed: %v", err)
			}
			tt.assert(t, result)
		})
	}
}
