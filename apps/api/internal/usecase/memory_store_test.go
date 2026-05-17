package usecase

import (
	"testing"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

func TestMemoryStoreUpsertAndRetrieveProject(t *testing.T) {
	store := NewMemoryStudioStore()

	project, err := store.UpsertProject(domain.Project{ID: "p-1", Name: "Test"})
	if err != nil {
		t.Fatalf("upsert project failed: %v", err)
	}
	if project.ID != "p-1" || project.Name != "Test" {
		t.Fatalf("unexpected project: %#v", project)
	}
	if project.CreatedAt == "" || project.UpdatedAt == "" {
		t.Fatalf("expected timestamps to be set")
	}

	got, ok, err := store.GetProject("p-1")
	if err != nil {
		t.Fatalf("get project failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected project to exist")
	}
	if got.Name != "Test" {
		t.Fatalf("expected name 'Test', got %q", got.Name)
	}
}

func TestMemoryStoreGetProjectNotFound(t *testing.T) {
	store := NewMemoryStudioStore()

	_, ok, err := store.GetProject("missing")
	if err != nil {
		t.Fatalf("get project failed: %v", err)
	}
	if ok {
		t.Fatalf("expected project not found")
	}
}

func TestMemoryStoreUpdateProject(t *testing.T) {
	store := NewMemoryStudioStore()
	store.UpsertProject(domain.Project{ID: "p-1", Name: "Old"})

	updated, ok, err := store.UpdateProject(domain.Project{ID: "p-1", Name: "New"})
	if err != nil {
		t.Fatalf("update project failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected project to be found")
	}
	if updated.Name != "New" {
		t.Fatalf("expected name 'New', got %q", updated.Name)
	}
}

func TestMemoryStoreUpdateProjectNotFound(t *testing.T) {
	store := NewMemoryStudioStore()

	_, ok, err := store.UpdateProject(domain.Project{ID: "missing", Name: "X"})
	if err != nil {
		t.Fatalf("update project failed: %v", err)
	}
	if ok {
		t.Fatalf("expected project not found")
	}
}

func TestMemoryStoreFrameLifecycle(t *testing.T) {
	store := NewMemoryStudioStore()

	frames, err := store.ListFrames("p-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}
	if len(frames) != 0 {
		t.Fatalf("expected empty frame list, got %d", len(frames))
	}

	frame := domain.Frame{
		ID:        "f-1",
		ProjectID: "p-1",
		Index:     0,
		ImageURL:  "data:image/png;base64,test",
		Kind:      domain.FrameKindKey,
	}
	if err := store.ReplaceFrames("p-1", []domain.Frame{frame}); err != nil {
		t.Fatalf("replace frames failed: %v", err)
	}

	frames, err = store.ListFrames("p-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}
	if frames[0].ImageURL != frame.ImageURL {
		t.Fatalf("unexpected image URL: %q", frames[0].ImageURL)
	}
}

func TestMemoryStoreReplaceFramesClearsPrevious(t *testing.T) {
	store := NewMemoryStudioStore()
	store.ReplaceFrames("p-1", []domain.Frame{{ID: "f-1", ProjectID: "p-1", Index: 0, Kind: domain.FrameKindKey}})
	store.ReplaceFrames("p-1", []domain.Frame{{ID: "f-2", ProjectID: "p-1", Index: 0, Kind: domain.FrameKindGenerated}})

	frames, _ := store.ListFrames("p-1")
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame after replace, got %d", len(frames))
	}
	if frames[0].ID != "f-2" {
		t.Fatalf("expected f-2, got %s", frames[0].ID)
	}
}

func TestMemoryStoreFindFrame(t *testing.T) {
	store := NewMemoryStudioStore()
	store.ReplaceFrames("p-1", []domain.Frame{
		{ID: "f-1", ProjectID: "p-1", Index: 0, Kind: domain.FrameKindKey},
		{ID: "f-2", ProjectID: "p-1", Index: 1, Kind: domain.FrameKindGenerated},
	})

	frame, ok, err := store.FindFrame("p-1", "f-2")
	if err != nil {
		t.Fatalf("find frame failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected frame to be found")
	}
	if frame.Kind != domain.FrameKindGenerated {
		t.Fatalf("unexpected kind: %s", frame.Kind)
	}

	_, ok, err = store.FindFrame("p-1", "missing")
	if err != nil {
		t.Fatalf("find frame failed: %v", err)
	}
	if ok {
		t.Fatalf("expected frame not found")
	}
}

func TestMemoryStoreDeleteFrame(t *testing.T) {
	store := NewMemoryStudioStore()
	store.ReplaceFrames("p-1", []domain.Frame{
		{ID: "f-1", ProjectID: "p-1", Index: 0, Kind: domain.FrameKindKey},
		{ID: "f-2", ProjectID: "p-1", Index: 1, Kind: domain.FrameKindGenerated},
		{ID: "f-3", ProjectID: "p-1", Index: 2, Kind: domain.FrameKindKey},
	})

	deleted, err := store.DeleteFrame("p-1", "f-2")
	if err != nil {
		t.Fatalf("delete frame failed: %v", err)
	}
	if !deleted {
		t.Fatalf("expected frame to be deleted")
	}

	frames, _ := store.ListFrames("p-1")
	if len(frames) != 2 {
		t.Fatalf("expected 2 frames after delete, got %d", len(frames))
	}

	deleted, err = store.DeleteFrame("p-1", "missing")
	if err != nil {
		t.Fatalf("delete frame failed: %v", err)
	}
	if deleted {
		t.Fatalf("expected missing frame not to be deleted")
	}
}

func TestMemoryStoreReorderFrames(t *testing.T) {
	store := NewMemoryStudioStore()
	store.ReplaceFrames("p-1", []domain.Frame{
		{ID: "f-1", ProjectID: "p-1", Index: 0, Kind: domain.FrameKindKey},
		{ID: "f-2", ProjectID: "p-1", Index: 1, Kind: domain.FrameKindGenerated},
		{ID: "f-3", ProjectID: "p-1", Index: 2, Kind: domain.FrameKindKey},
	})

	frames, err := store.ReorderFrames("p-1", []string{"f-3", "f-1", "f-2"})
	if err != nil {
		t.Fatalf("reorder frames failed: %v", err)
	}
	if frames[0].ID != "f-3" || frames[0].Index != 0 {
		t.Fatalf("expected f-3 at index 0, got %s", frames[0].ID)
	}
	if frames[1].ID != "f-1" || frames[1].Index != 1 {
		t.Fatalf("expected f-1 at index 1, got %s", frames[1].ID)
	}
	if frames[2].ID != "f-2" || frames[2].Index != 2 {
		t.Fatalf("expected f-2 at index 2, got %s", frames[2].ID)
	}
}

func TestMemoryStoreReorderMismatchedFrames(t *testing.T) {
	store := NewMemoryStudioStore()
	store.ReplaceFrames("p-1", []domain.Frame{
		{ID: "f-1", ProjectID: "p-1", Index: 0, Kind: domain.FrameKindKey},
	})

	_, err := store.ReorderFrames("p-1", []string{"f-1", "f-2"})
	if err == nil {
		t.Fatalf("expected error for mismatched frame count")
	}
}

func TestMemoryStoreJobLifecycle(t *testing.T) {
	store := NewMemoryStudioStore()

	job := domain.Job{
		ID:      "j-1",
		Type:    domain.JobTypeGeneration,
		Status:  domain.JobStatusQueued,
		Message: "test",
		Version: 1,
	}
	if err := store.CreateJob(job); err != nil {
		t.Fatalf("create job failed: %v", err)
	}

	got, ok, err := store.GetJob("j-1")
	if err != nil {
		t.Fatalf("get job failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected job to exist")
	}
	if got.Status != domain.JobStatusQueued {
		t.Fatalf("expected queued, got %s", got.Status)
	}

	got.Status = domain.JobStatusRunning
	updated, err := store.UpdateJob(got)
	if err != nil {
		t.Fatalf("update job failed: %v", err)
	}
	if !updated {
		t.Fatalf("expected job update to succeed")
	}

	got, ok, _ = store.GetJob("j-1")
	if !ok || got.Status != domain.JobStatusRunning {
		t.Fatalf("expected running status, got %v", got)
	}
}

func TestMemoryStoreUpdateJobVersionConflict(t *testing.T) {
	store := NewMemoryStudioStore()
	store.CreateJob(domain.Job{ID: "j-1", Type: domain.JobTypeGeneration, Status: domain.JobStatusQueued, Version: 1})

	original, _, _ := store.GetJob("j-1")

	original.Status = domain.JobStatusRunning
	store.UpdateJob(original)

	original.Status = domain.JobStatusFailed
	updated, _ := store.UpdateJob(original)
	if updated {
		t.Fatalf("expected version conflict to prevent update")
	}
}

func TestMemoryStoreListActiveJobs(t *testing.T) {
	store := NewMemoryStudioStore()
	jobs := []domain.Job{
		{ID: "j-queued", ProjectID: "p-1", Type: domain.JobTypeGeneration, Status: domain.JobStatusQueued, Version: 1},
		{ID: "j-running", ProjectID: "p-1", Type: domain.JobTypeGeneration, Status: domain.JobStatusRunning, Version: 1},
		{ID: "j-done", ProjectID: "p-1", Type: domain.JobTypeGeneration, Status: domain.JobStatusSucceeded, Version: 1},
		{ID: "j-other-project", ProjectID: "p-2", Type: domain.JobTypeGeneration, Status: domain.JobStatusQueued, Version: 1},
		{ID: "j-other-type", ProjectID: "p-1", Type: domain.JobTypeExport, Status: domain.JobStatusQueued, Version: 1},
	}
	for _, job := range jobs {
		if err := store.CreateJob(job); err != nil {
			t.Fatalf("create job failed: %v", err)
		}
	}

	active, err := store.ListActiveJobs("p-1", domain.JobTypeGeneration)
	if err != nil {
		t.Fatalf("list active jobs failed: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("expected 2 active jobs, got %d", len(active))
	}
	for _, job := range active {
		if job.ProjectID != "p-1" || job.Type != domain.JobTypeGeneration {
			t.Fatalf("unexpected active job: %+v", job)
		}
		if job.Status != domain.JobStatusQueued && job.Status != domain.JobStatusRunning {
			t.Fatalf("expected only queued/running jobs, got %s", job.Status)
		}
	}
}

func TestMemoryStoreUpdateFrameMetadata(t *testing.T) {
	store := NewMemoryStudioStore()
	store.ReplaceFrames("p-1", []domain.Frame{
		{ID: "f-1", ProjectID: "p-1", Index: 0, Kind: domain.FrameKindKey},
	})

	note := "updated note"
	kind := domain.FrameKindEdited
	frame, ok, err := store.UpdateFrameMetadata(domain.UpdateFrameMetadataRequest{
		ProjectID: "p-1",
		FrameID:   "f-1",
		Note:      &note,
		Kind:      &kind,
	})
	if err != nil {
		t.Fatalf("update metadata failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected frame to be found")
	}
	if frame.Note != "updated note" || frame.Kind != domain.FrameKindEdited {
		t.Fatalf("unexpected metadata: note=%q kind=%q", frame.Note, frame.Kind)
	}
}

func TestMemoryStoreUpsertFrame(t *testing.T) {
	store := NewMemoryStudioStore()
	store.ReplaceFrames("p-1", []domain.Frame{
		{ID: "f-1", ProjectID: "p-1", Index: 0, Kind: domain.FrameKindKey},
	})

	newFrame := domain.Frame{
		ID:        "f-2",
		ProjectID: "p-1",
		Index:     1,
		Kind:      domain.FrameKindGenerated,
		ImageURL:  "https://example.com/test.png",
	}
	if err := store.UpsertFrame(newFrame); err != nil {
		t.Fatalf("upsert frame failed: %v", err)
	}

	frames, _ := store.ListFrames("p-1")
	if len(frames) != 2 {
		t.Fatalf("expected 2 frames after upsert, got %d", len(frames))
	}
}
