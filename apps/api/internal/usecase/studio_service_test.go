package usecase

import (
	"testing"
	"time"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

func TestGenerateFramesCreatesTimeline(t *testing.T) {
	service := NewStudioService()

	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:         "project-1",
		Prompt:            "turn around",
		FrameCount:        3,
		StartImageDataURL: "data:image/png;base64,start",
		EndImageDataURL:   "data:image/png;base64,end",
	})

	job = waitForJob(t, service, job.ID)
	if job.Status != domain.JobStatusSucceeded {
		t.Fatalf("expected succeeded job, got %s", job.Status)
	}

	frames, err := service.ListFrames("project-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}
	if len(frames) != 5 {
		t.Fatalf("expected 5 frames including keyframes, got %d", len(frames))
	}
	if frames[0].Kind != domain.FrameKindKey || frames[4].Kind != domain.FrameKindKey {
		t.Fatalf("expected first and last frames to be keyframes")
	}
}

func TestUpdateFrameMarksFrameEdited(t *testing.T) {
	service := NewStudioService()
	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:  "project-1",
		FrameCount: 2,
	})
	waitForJob(t, service, job.ID)
	frames, err := service.ListFrames("project-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}
	frame := frames[1]

	updated, err := service.UpdateFrame(domain.UpdateFrameRequest{
		ProjectID:    "project-1",
		FrameID:      frame.ID,
		ImageDataURL: "data:image/png;base64,edited",
		Note:         "manual edit",
	})
	if err != nil {
		t.Fatalf("update frame failed: %v", err)
	}
	if updated.Kind != domain.FrameKindEdited {
		t.Fatalf("expected edited frame, got %s", updated.Kind)
	}
}

func waitForJob(t *testing.T, service *StudioService, jobID string) domain.Job {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		job, ok, err := service.GetJob(jobID)
		if err != nil {
			t.Fatalf("get job failed: %v", err)
		}
		if !ok {
			t.Fatalf("job not found: %s", jobID)
		}
		if job.Status == domain.JobStatusSucceeded || job.Status == domain.JobStatusFailed {
			return job
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for job %s", jobID)
	return domain.Job{}
}
