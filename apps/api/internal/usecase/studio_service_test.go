package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
	"github.com/haseg/anifusion-canvas/apps/api/internal/infrastructure/replicate"
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

func TestGenerateFramesStoresImagesWhenConfigured(t *testing.T) {
	objectStore := &fakeObjectStore{}
	service := NewStudioServiceWithStoreAndObjects(NewMemoryStudioStore(), objectStore)

	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:         "project-1",
		FrameCount:        2,
		StartImageDataURL: "data:image/png;base64,start",
		EndImageDataURL:   "data:image/png;base64,end",
	})
	job = waitForJob(t, service, job.ID)
	if job.Status != domain.JobStatusSucceeded {
		t.Fatalf("expected succeeded job, got %s: %s", job.Status, job.Error)
	}

	frames, err := service.ListFrames("project-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}
	if len(frames) != 4 {
		t.Fatalf("expected 4 frames, got %d", len(frames))
	}
	for _, frame := range frames {
		if !strings.HasPrefix(frame.ImageURL, "https://assets.example.test/projects/project-1/") {
			t.Fatalf("expected R2-backed frame URL, got %q", frame.ImageURL)
		}
	}
	if !objectStore.hasKeyPrefix("projects/project-1/inputs/") {
		t.Fatalf("expected keyframe input objects, got %#v", objectStore.calls)
	}
	if !objectStore.hasKeyPrefix("projects/project-1/frames/") {
		t.Fatalf("expected generated frame objects, got %#v", objectStore.calls)
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

func TestUpdateFrameReturnsObjectStoreErrors(t *testing.T) {
	objectStore := &fakeObjectStore{}
	service := NewStudioServiceWithStoreAndObjects(NewMemoryStudioStore(), objectStore)
	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:  "project-1",
		FrameCount: 2,
	})
	waitForJob(t, service, job.ID)
	objectStore.err = fmt.Errorf("upload failed")
	frames, err := service.ListFrames("project-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}

	_, err = service.UpdateFrame(context.Background(), domain.UpdateFrameRequest{
		ProjectID:    "project-1",
		FrameID:      frames[1].ID,
		ImageDataURL: "data:image/png;base64,edited",
	})
	if err == nil {
		t.Fatalf("expected update frame to fail")
	}
	if err.Error() != "store edited frame object: upload failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInpaintFrameStoresMaskAndResultWhenConfigured(t *testing.T) {
	objectStore := &fakeObjectStore{}
	service := NewStudioServiceWithStoreAndObjects(NewMemoryStudioStore(), objectStore)
	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:         "project-1",
		FrameCount:        2,
		StartImageDataURL: "data:image/png;base64,start",
		EndImageDataURL:   "data:image/png;base64,end",
	})
	waitForJob(t, service, job.ID)
	frames, err := service.ListFrames("project-1")
	if err != nil {
		t.Fatalf("list frames failed: %v", err)
	}

	inpaintJob := service.InpaintFrame(domain.InpaintFrameRequest{
		ProjectID:   "project-1",
		FrameID:     frames[1].ID,
		Prompt:      "repair hand",
		MaskDataURL: validMaskDataURL(t),
		Strength:    0.65,
	})
	inpaintJob = waitForJob(t, service, inpaintJob.ID)
	if inpaintJob.Status != domain.JobStatusSucceeded {
		t.Fatalf("expected succeeded inpainting job, got %s: %s", inpaintJob.Status, inpaintJob.Error)
	}

	result, ok := inpaintJob.Result.(domain.InpaintFrameResult)
	if !ok {
		t.Fatalf("expected typed inpainting result, got %T", inpaintJob.Result)
	}
	if result.Frame.Kind != domain.FrameKindInpainted {
		t.Fatalf("expected inpainted frame, got %s", result.Frame.Kind)
	}
	if !strings.HasPrefix(result.Frame.ImageURL, "https://assets.example.test/projects/project-1/frames/") {
		t.Fatalf("expected R2-backed inpaint result URL, got %q", result.Frame.ImageURL)
	}
	if !objectStore.hasKeyPrefix("projects/project-1/masks/") {
		t.Fatalf("expected inpainting mask object, got %#v", objectStore.calls)
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
	err     error
	calls   []fakeObjectPut
}

func (s *fakeObjectStore) PutDataURL(_ context.Context, key string, dataURL string) (domain.StorageObject, error) {
	s.key = key
	s.dataURL = dataURL
	s.calls = append(s.calls, fakeObjectPut{key: key, dataURL: dataURL})
	if s.err != nil {
		return domain.StorageObject{}, s.err
	}
	return domain.StorageObject{
		Key:         key,
		URL:         "https://assets.example.test/" + key,
		ContentType: "image/png",
		Size:        int64(len(dataURL)),
	}, nil
}

func (s *fakeObjectStore) PutBytes(_ context.Context, key string, contentType string, data []byte) (domain.StorageObject, error) {
	s.calls = append(s.calls, fakeObjectPut{key: key, dataURL: contentType})
	if s.err != nil {
		return domain.StorageObject{}, s.err
	}
	return domain.StorageObject{
		Key:         key,
		URL:         "https://assets.example.test/" + key,
		ContentType: contentType,
		Size:        int64(len(data)),
	}, nil
}

func (s *fakeObjectStore) hasKeyPrefix(prefix string) bool {
	for _, call := range s.calls {
		if strings.HasPrefix(call.key, prefix) {
			return true
		}
	}
	return false
}

type fakeObjectPut struct {
	key     string
	dataURL string
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

func validMaskDataURL(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode mask PNG: %v", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func generateTestMP4Bytes(t *testing.T, frameCount int) []byte {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}

	tmpDir, err := os.MkdirTemp("", "usecase-test-mp4-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	var framePaths []string
	for i := 1; i <= frameCount; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("frame_%04d.png", i))
		img := image.NewRGBA(image.Rect(0, 0, 640, 360))
		for y := 0; y < 360; y++ {
			for x := 0; x < 640; x++ {
				img.SetRGBA(x, y, color.RGBA{R: uint8(i * 40), G: 150, B: 200, A: 255})
			}
		}
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create frame PNG: %v", err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode frame PNG: %v", err)
		}
		f.Close()
		framePaths = append(framePaths, path)
	}

	listPath := filepath.Join(tmpDir, "input.txt")
	var content string
	for _, p := range framePaths {
		content += "file '" + p + "'\n"
		content += "duration 0.25\n"
	}
	if err := os.WriteFile(listPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input list: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "output.mp4")
	cmd := exec.Command("ffmpeg",
		"-f", "concat", "-safe", "0", "-i", listPath,
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-r", "4",
		"-y", outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg encode MP4: %v\n%s", err, string(out))
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output MP4: %v", err)
	}
	return data
}

type mockReplicateClient struct {
	createOutput        any
	createErr           error
	pollResult          *replicate.Prediction
	pollErr             error
	downloadData        []byte
	downloadErr         error
	createCalledVersion string
}

func (m *mockReplicateClient) CreatePrediction(_ context.Context, version string, input map[string]any) (*replicate.Prediction, error) {
	m.createCalledVersion = version
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &replicate.Prediction{
		ID:     "pred-001",
		Status: replicate.StatusStarting,
	}, nil
}

func (m *mockReplicateClient) GetPrediction(_ context.Context, id string) (*replicate.Prediction, error) {
	if m.pollResult != nil {
		return m.pollResult, nil
	}
	return &replicate.Prediction{
		ID:     id,
		Status: replicate.StatusSucceeded,
		Output: m.createOutput,
	}, nil
}

func (m *mockReplicateClient) WaitForPrediction(ctx context.Context, id string, pollInterval time.Duration) (*replicate.Prediction, error) {
	if m.pollErr != nil {
		return nil, m.pollErr
	}
	if m.pollResult != nil {
		return m.pollResult, nil
	}
	return &replicate.Prediction{
		ID:     id,
		Status: replicate.StatusSucceeded,
		Output: m.createOutput,
	}, nil
}

func (m *mockReplicateClient) DownloadOutput(_ context.Context, url string) ([]byte, error) {
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	if m.downloadData != nil {
		return m.downloadData, nil
	}
	return []byte("fake-result"), nil
}

func TestGenerateFramesWithReplicateCreatesFrames(t *testing.T) {
	objectStore := &fakeObjectStore{}
	replicateClient := &mockReplicateClient{
		createOutput: "https://replicate.delivery/test.mp4",
		downloadData: generateTestMP4Bytes(t, 2),
	}

	service := NewStudioServiceWithDependencies(
		NewMemoryStudioStore(),
		objectStore,
		replicateClient,
		"fofr/tooncrafter",
		"lucataco/sdxl-inpainting",
	)

	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:         "proj-rep-gen",
		Prompt:            "turn around",
		FrameCount:        2,
		StartImageDataURL: "data:image/png;base64,start",
		EndImageDataURL:   "data:image/png;base64,end",
	})
	job = waitForJob(t, service, job.ID)

	if job.Status != domain.JobStatusSucceeded {
		t.Fatalf("expected succeeded job, got %s: %s", job.Status, job.Error)
	}

	result, ok := job.Result.(domain.GenerateFramesResult)
	if !ok {
		t.Fatalf("expected typed generation result, got %T", job.Result)
	}
	if result.RawVideoURL == "" {
		t.Fatalf("expected raw video URL in result")
	}
	if len(result.Frames) != 4 {
		t.Fatalf("expected 4 frames, got %d", len(result.Frames))
	}
	if result.Frames[1].Kind != domain.FrameKindGenerated {
		t.Fatalf("expected generated frame, got %s", result.Frames[1].Kind)
	}
}

func TestInpaintFrameWithReplicateUpdatesFrame(t *testing.T) {
	objectStore := &fakeObjectStore{}
	genReplicate := &mockReplicateClient{
		createOutput: "https://replicate.delivery/test.mp4",
		downloadData: generateTestMP4Bytes(t, 2),
	}
	inpaintReplicate := &mockReplicateClient{
		createOutput: []any{"https://replicate.delivery/inpainted.png"},
	}

	// Generation needs a Replicate client that returns video output
	genService := NewStudioServiceWithDependencies(
		NewMemoryStudioStore(),
		objectStore,
		genReplicate,
		"fofr/tooncrafter",
		"",
	)
	genJob := genService.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:         "proj-rep-inp",
		Prompt:            "walk",
		FrameCount:        2,
		StartImageDataURL: "data:image/png;base64,start",
		EndImageDataURL:   "data:image/png;base64,end",
	})
	waitForJob(t, genService, genJob.ID)

	// Inpainting uses its own service with a separate Replicate client (list output)
	service := NewStudioServiceWithDependencies(
		genService.store,
		objectStore,
		inpaintReplicate,
		"",
		"lucataco/sdxl-inpainting",
	)
	frames, _ := service.ListFrames("proj-rep-inp")

	inpaintJob := service.InpaintFrame(domain.InpaintFrameRequest{
		ProjectID:   "proj-rep-inp",
		FrameID:     frames[1].ID,
		Prompt:      "fix the hand",
		MaskDataURL: validMaskDataURL(t),
		Strength:    0.8,
	})
	inpaintJob = waitForJob(t, service, inpaintJob.ID)

	if inpaintJob.Status != domain.JobStatusSucceeded {
		t.Fatalf("expected succeeded inpainting job, got %s: %s", inpaintJob.Status, inpaintJob.Error)
	}

	result, ok := inpaintJob.Result.(domain.InpaintFrameResult)
	if !ok {
		t.Fatalf("expected typed inpainting result, got %T", inpaintJob.Result)
	}
	if result.Frame.Kind != domain.FrameKindInpainted {
		t.Fatalf("expected inpainted frame, got %s", result.Frame.Kind)
	}
	if result.Frame.Note != "fix the hand" {
		t.Fatalf("expected prompt in note, got %q", result.Frame.Note)
	}
}

func TestGenerateFramesReplicateErrorMarksJobFailed(t *testing.T) {
	replicateClient := &mockReplicateClient{
		createErr: fmt.Errorf("API rate limit exceeded"),
	}

	service := NewStudioServiceWithDependencies(
		NewMemoryStudioStore(),
		nil,
		replicateClient,
		"fofr/tooncrafter",
		"",
	)

	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:         "proj-1",
		Prompt:            "turn",
		FrameCount:        2,
		StartImageDataURL: "data:image/png;base64,start",
		EndImageDataURL:   "data:image/png;base64,end",
	})
	job = waitForJob(t, service, job.ID)

	if job.Status != domain.JobStatusFailed {
		t.Fatalf("expected failed job, got %s", job.Status)
	}
}

func TestInpaintFrameReplicatePredictionFailed(t *testing.T) {
	inpaintReplicate := &mockReplicateClient{
		pollResult: &replicate.Prediction{
			ID:     "pred-001",
			Status: replicate.StatusFailed,
			Error:  "CUDA out of memory",
		},
	}

	// Generate frames in demo mode (no Replicate client)
	demoService := NewStudioService()
	genJob := demoService.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:  "proj-inp-fail",
		Prompt:     "walk",
		FrameCount: 2,
	})
	waitForJob(t, demoService, genJob.ID)
	frames, _ := demoService.ListFrames("proj-inp-fail")

	// Inpainting with the failing Replicate client
	service := NewStudioServiceWithDependencies(
		demoService.store,
		nil,
		inpaintReplicate,
		"",
		"lucataco/sdxl-inpainting",
	)

	inpaintJob := service.InpaintFrame(domain.InpaintFrameRequest{
		ProjectID:   "proj-inp-fail",
		FrameID:     frames[1].ID,
		Prompt:      "fix",
		MaskDataURL: validMaskDataURL(t),
		Strength:    0.5,
	})
	inpaintJob = waitForJob(t, service, inpaintJob.ID)

	if inpaintJob.Status != domain.JobStatusFailed {
		t.Fatalf("expected failed inpainting job, got %s", inpaintJob.Status)
	}
}

func TestStudioServiceFallsBackToDemoWithoutReplicateClient(t *testing.T) {
	service := NewStudioServiceWithDependencies(
		NewMemoryStudioStore(),
		nil,
		nil,
		"fofr/tooncrafter",
		"lucataco/sdxl-inpainting",
	)

	job := service.GenerateFrames(domain.GenerateFramesRequest{
		ProjectID:         "proj-1",
		Prompt:            "turn",
		FrameCount:        2,
		StartImageDataURL: "data:image/png;base64,start",
		EndImageDataURL:   "data:image/png;base64,end",
	})
	job = waitForJob(t, service, job.ID)

	if job.Status != domain.JobStatusSucceeded {
		t.Fatalf("expected succeeded demo job, got %s: %s", job.Status, job.Error)
	}
	result, _ := job.Result.(domain.GenerateFramesResult)
	if result.RawVideoURL != "" {
		t.Fatalf("expected no raw video URL in demo mode")
	}
}
