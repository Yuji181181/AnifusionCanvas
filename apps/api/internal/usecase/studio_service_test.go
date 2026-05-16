package usecase

import (
	"context"
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
	if job.Type != domain.JobTypeGeneration {
		t.Fatalf("expected generation job type, got %s", job.Type)
	}
	if _, ok := job.Result.(domain.GenerateFramesResult); !ok {
		t.Fatalf("expected typed generation result, got %T", job.Result)
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

func TestProjectLifecycle(t *testing.T) {
	service := NewStudioService()

	project, err := service.CreateProject(domain.CreateProjectRequest{
		ID:   "project-1",
		Name: "Demo project",
	})
	if err != nil {
		t.Fatalf("create project failed: %v", err)
	}
	if project.Name != "Demo project" {
		t.Fatalf("expected project name to be set, got %q", project.Name)
	}

	project, ok, err := service.GetProject("project-1")
	if err != nil {
		t.Fatalf("get project failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected project to exist")
	}
	if project.CreatedAt == "" || project.UpdatedAt == "" {
		t.Fatalf("expected timestamps to be set")
	}

	project, ok, err = service.UpdateProject(domain.UpdateProjectRequest{
		ID:   "project-1",
		Name: "Renamed project",
	})
	if err != nil {
		t.Fatalf("update project failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected project update to find project")
	}
	if project.Name != "Renamed project" {
		t.Fatalf("expected project to be renamed, got %q", project.Name)
	}
}

func TestGenerateFramesDoesNotOverwriteProjectName(t *testing.T) {
	service := NewStudioService()
	_, err := service.CreateProject(domain.CreateProjectRequest{
		ID:   "project-1",
		Name: "Story cut 01",
	})
	if err != nil {
		t.Fatalf("create project failed: %v", err)
	}

	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:  "project-1",
		FrameCount: 2,
	})
	waitForJob(t, service, job.ID)

	project, ok, err := service.GetProject("project-1")
	if err != nil {
		t.Fatalf("get project failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected project to exist")
	}
	if project.Name != "Story cut 01" {
		t.Fatalf("expected project name to be preserved, got %q", project.Name)
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

	updated, err := service.UpdateFrame(context.Background(), domain.UpdateFrameRequest{
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

func TestUpdateFrameStoresEditedFrameObjectWhenConfigured(t *testing.T) {
	objectStore := &fakeObjectStore{}
	service := NewStudioServiceWithStoreAndObjects(NewMemoryStudioStore(), objectStore)
	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:  "project-1",
		FrameCount: 2,
	})
	waitForJob(t, service, job.ID)
	frames, err := service.ListFrames("project-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}

	updated, err := service.UpdateFrame(context.Background(), domain.UpdateFrameRequest{
		ProjectID:    "project-1",
		FrameID:      frames[1].ID,
		ImageDataURL: "data:image/png;base64,edited",
		Note:         "manual edit",
	})
	if err != nil {
		t.Fatalf("update frame failed: %v", err)
	}

	expectedKey := "projects/project-1/frames/" + frames[1].ID + ".png"
	if objectStore.key != expectedKey {
		t.Fatalf("expected object key %q, got %q", expectedKey, objectStore.key)
	}
	if updated.ImageURL != "https://assets.example.test/"+expectedKey {
		t.Fatalf("expected R2-backed image URL, got %q", updated.ImageURL)
	}
	if updated.ThumbnailURL != updated.ImageURL {
		t.Fatalf("expected thumbnail URL to match image URL")
	}
}

func TestFrameMetadataDeleteAndReorder(t *testing.T) {
	service := NewStudioService()
	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:  "project-1",
		FrameCount: 3,
	})
	waitForJob(t, service, job.ID)

	frames, err := service.ListFrames("project-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}
	note := "needs cleanup"
	kind := domain.FrameKindEdited
	updated, ok, err := service.UpdateFrameMetadata(domain.UpdateFrameMetadataRequest{
		ProjectID: "project-1",
		FrameID:   frames[1].ID,
		Kind:      &kind,
		Note:      &note,
	})
	if err != nil {
		t.Fatalf("update frame metadata failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected frame to be found")
	}
	if updated.Kind != domain.FrameKindEdited || updated.Note != note {
		t.Fatalf("unexpected metadata update: %#v", updated)
	}

	reordered, err := service.ReorderFrames(domain.ReorderFramesRequest{
		ProjectID: "project-1",
		FrameIDs:  []string{frames[2].ID, frames[1].ID, frames[0].ID, frames[3].ID, frames[4].ID},
	})
	if err != nil {
		t.Fatalf("reorder frames failed: %v", err)
	}
	if reordered[0].ID != frames[2].ID || reordered[0].Index != 0 {
		t.Fatalf("expected frame %s at index 0, got %#v", frames[2].ID, reordered[0])
	}

	deleted, err := service.DeleteFrame("project-1", frames[1].ID)
	if err != nil {
		t.Fatalf("delete frame failed: %v", err)
	}
	if !deleted {
		t.Fatalf("expected frame to be deleted")
	}
	remaining, err := service.ListFrames("project-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}
	if len(remaining) != 4 {
		t.Fatalf("expected 4 remaining frames, got %d", len(remaining))
	}
	for index, frame := range remaining {
		if frame.Index != index {
			t.Fatalf("expected compact frame index %d, got %d", index, frame.Index)
		}
	}
}

type fakeObjectStore struct {
	key     string
	dataURL string
}

func (s *fakeObjectStore) PutDataURL(_ context.Context, key string, dataURL string) (domain.StorageObject, error) {
	s.key = key
	s.dataURL = dataURL
	return domain.StorageObject{
		Key:         key,
		URL:         "https://assets.example.test/" + key,
		ContentType: "image/png",
		Size:        int64(len(dataURL)),
	}, nil
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
