package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
	"github.com/haseg/anifusion-canvas/apps/api/internal/infrastructure/media"
	"github.com/haseg/anifusion-canvas/apps/api/internal/infrastructure/replicate"
)

var ErrFrameNotFound = errors.New("frame not found")

type StudioStore interface {
	UpsertProject(project domain.Project) (domain.Project, error)
	GetProject(projectID string) (domain.Project, bool, error)
	UpdateProject(project domain.Project) (domain.Project, bool, error)
	ListFrames(projectID string) ([]domain.Frame, error)
	ReplaceFrames(projectID string, frames []domain.Frame) error
	FindFrame(projectID string, frameID string) (domain.Frame, bool, error)
	UpsertFrame(frame domain.Frame) error
	UpdateFrameMetadata(input domain.UpdateFrameMetadataRequest) (domain.Frame, bool, error)
	DeleteFrame(projectID string, frameID string) (bool, error)
	ReorderFrames(projectID string, frameIDs []string) ([]domain.Frame, error)
	CreateJob(job domain.Job) error
	UpdateJob(job domain.Job) error
	GetJob(jobID string) (domain.Job, bool, error)
}

type ObjectStore interface {
	PutDataURL(ctx context.Context, key string, dataURL string) (domain.StorageObject, error)
	PutBytes(ctx context.Context, key string, contentType string, data []byte) (domain.StorageObject, error)
}

// ReplicateClient abstracts the Replicate API for use by StudioService.
type ReplicateClient interface {
	CreatePrediction(ctx context.Context, version string, input map[string]any) (*replicate.Prediction, error)
	GetPrediction(ctx context.Context, id string) (*replicate.Prediction, error)
	WaitForPrediction(ctx context.Context, id string, pollInterval time.Duration) (*replicate.Prediction, error)
	DownloadOutput(ctx context.Context, url string) ([]byte, error)
}

type StudioService struct {
	store             StudioStore
	objectStore       ObjectStore
	replicateClient   ReplicateClient
	toonCrafterVer    string
	sdxlInpaintingVer string
}

func NewStudioService() *StudioService {
	return NewStudioServiceWithStore(NewMemoryStudioStore())
}

func NewStudioServiceWithStore(store StudioStore) *StudioService {
	return &StudioService{store: store}
}

func NewStudioServiceWithStoreAndObjects(store StudioStore, objectStore ObjectStore) *StudioService {
	return &StudioService{store: store, objectStore: objectStore}
}

func NewStudioServiceWithDependencies(store StudioStore, objectStore ObjectStore, replicateClient ReplicateClient, toonCrafterVer string, sdxlInpaintingVer string) *StudioService {
	return &StudioService{
		store:             store,
		objectStore:       objectStore,
		replicateClient:   replicateClient,
		toonCrafterVer:    toonCrafterVer,
		sdxlInpaintingVer: sdxlInpaintingVer,
	}
}

func (s *StudioService) CreateProject(input domain.CreateProjectRequest) (domain.Project, error) {
	return s.store.UpsertProject(domain.Project{
		ID:   input.ID,
		Name: input.Name,
	})
}

func (s *StudioService) GetProject(projectID string) (domain.Project, bool, error) {
	return s.store.GetProject(projectID)
}

func (s *StudioService) UpdateProject(input domain.UpdateProjectRequest) (domain.Project, bool, error) {
	return s.store.UpdateProject(domain.Project{
		ID:   input.ID,
		Name: input.Name,
	})
}

func (s *StudioService) ListFrames(projectID string) ([]domain.Frame, error) {
	return s.store.ListFrames(projectID)
}

func (s *StudioService) GenerateFrames(input domain.GenerateFramesRequest) domain.Job {
	if input.FrameCount < 2 {
		input.FrameCount = 2
	}
	if input.FrameCount > 12 {
		input.FrameCount = 12
	}

	job := s.createJob(input.ProjectID, domain.JobTypeGeneration, "中割り生成を受け付けました")
	if job.Status == domain.JobStatusFailed {
		return job
	}

	go s.runJob(job.ID, func(update func(int, string)) (any, error) {
		if s.replicateClient != nil {
			return s.generateFramesWithReplicate(job, input, update)
		}
		return s.generateFramesDemo(job, input, update)
	})

	return job
}

func (s *StudioService) generateFramesWithReplicate(job domain.Job, input domain.GenerateFramesRequest, update func(int, string)) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	update(10, "ToonCrafterで推論を開始しています")
	toonInput := replicate.BuildToonCrafterInput(input)
	prediction, err := s.replicateClient.CreatePrediction(ctx, s.toonCrafterVer, toonInput)
	if err != nil {
		return nil, fmt.Errorf("ToonCrafter推論の開始に失敗しました: %w", err)
	}

	update(25, "AIがフレームを生成しています")
	prediction, err = s.replicateClient.WaitForPrediction(ctx, prediction.ID, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("推論の待機に失敗しました: %w", err)
	}
	if prediction.Status != replicate.StatusSucceeded {
		return nil, fmt.Errorf("ToonCrafter推論に失敗しました: %s", prediction.Error)
	}

	videoURL := replicate.ParseToonCrafterOutput(prediction)
	if videoURL == "" {
		return nil, fmt.Errorf("ToonCrafterが動画出力を返しませんでした")
	}

	update(55, "生成結果をダウンロードしています")
	videoData, err := s.replicateClient.DownloadOutput(ctx, videoURL)
	if err != nil {
		return nil, fmt.Errorf("生成動画のダウンロードに失敗しました: %w", err)
	}

	var rawVideoURL string
	if s.objectStore != nil {
		update(70, "動画を保存しています")
		videoKey := fmt.Sprintf("projects/%s/generated/%s.mp4", job.ProjectID, job.ID)
		videoObj, err := s.objectStore.PutBytes(ctx, videoKey, "video/mp4", videoData)
		if err != nil {
			return nil, fmt.Errorf("生成動画の保存に失敗しました: %w", err)
		}
		rawVideoURL = videoObj.URL

		update(80, "フレームを分割しています")
		frames, err := s.splitMP4ToFrames(ctx, job, input, videoData)
		if err != nil {
			return nil, fmt.Errorf("フレーム分割に失敗しました: %w", err)
		}

		if err := s.store.ReplaceFrames(input.ProjectID, frames); err != nil {
			return nil, err
		}

		return domain.GenerateFramesResult{
			Frames:      frames,
			RawVideoURL: rawVideoURL,
		}, nil
	}

	update(90, "タイムラインへ登録しています")
	frames := make([]domain.Frame, 0, input.FrameCount+2)
	frames = append(frames, s.newFrame(input.ProjectID, 0, domain.FrameKindKey, input.StartImageDataURL, "start keyframe"))
	for i := 0; i < input.FrameCount; i++ {
		label := fmt.Sprintf("AI %02d", i+1)
		hue := int(math.Mod(float64(178+i*24), 360))
		frames = append(frames, s.newFrame(input.ProjectID, i+1, domain.FrameKindGenerated, demoImage(label, hue), input.Prompt))
	}
	frames = append(frames, s.newFrame(input.ProjectID, input.FrameCount+1, domain.FrameKindKey, input.EndImageDataURL, "end keyframe"))

	if err := s.store.ReplaceFrames(input.ProjectID, frames); err != nil {
		return nil, err
	}

	return domain.GenerateFramesResult{
		Frames:      frames,
		RawVideoURL: rawVideoURL,
	}, nil
}

func (s *StudioService) splitMP4ToFrames(ctx context.Context, job domain.Job, input domain.GenerateFramesRequest, videoData []byte) ([]domain.Frame, error) {
	tmpDir, err := os.MkdirTemp("", "anifusion-tooncrafter-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	mp4Path := filepath.Join(tmpDir, "output.mp4")
	if err := os.WriteFile(mp4Path, videoData, 0o644); err != nil {
		return nil, fmt.Errorf("write temp MP4: %w", err)
	}

	framesDir := filepath.Join(tmpDir, "frames")
	framePaths, err := media.SplitMP4ToFrames(ctx, mp4Path, framesDir)
	if err != nil {
		return nil, err
	}

	generatedFrames := make([]domain.Frame, input.FrameCount)
	actualCount := len(framePaths)
	for i := 0; i < input.FrameCount; i++ {
		var filePath string
		if i < actualCount {
			filePath = framePaths[i]
		} else {
			filePath = framePaths[actualCount-1]
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read frame file %d: %w", i+1, err)
		}

		frameID := fmt.Sprintf("frame-%d-%d-gen", i+1, time.Now().UnixNano())
		objectKey := fmt.Sprintf("projects/%s/frames/%s.png", job.ProjectID, frameID)
		object, err := s.objectStore.PutBytes(ctx, objectKey, "image/png", data)
		if err != nil {
			return nil, fmt.Errorf("store frame %d: %w", i+1, err)
		}

		generatedFrames[i] = domain.Frame{
			ID:           frameID,
			ProjectID:    job.ProjectID,
			Index:        i + 1,
			ImageURL:     object.URL,
			ThumbnailURL: object.URL,
			Kind:         domain.FrameKindGenerated,
			Note:         input.Prompt,
			UpdatedAt:    now(),
		}
	}

	frames := make([]domain.Frame, 0, input.FrameCount+2)
	frames = append(frames, s.newFrame(input.ProjectID, 0, domain.FrameKindKey, input.StartImageDataURL, "start keyframe"))

	startKeyObject, err := s.objectStore.PutDataURL(ctx, projectObjectKey(input.ProjectID, "inputs", frames[0].ID, input.StartImageDataURL), input.StartImageDataURL)
	if err == nil {
		frames[0].ImageURL = startKeyObject.URL
		frames[0].ThumbnailURL = startKeyObject.URL
	}

	frames = append(frames, generatedFrames...)
	frames = append(frames, s.newFrame(input.ProjectID, input.FrameCount+1, domain.FrameKindKey, input.EndImageDataURL, "end keyframe"))

	endKeyObject, err := s.objectStore.PutDataURL(ctx, projectObjectKey(input.ProjectID, "inputs", frames[len(frames)-1].ID, input.EndImageDataURL), input.EndImageDataURL)
	if err == nil {
		frames[len(frames)-1].ImageURL = endKeyObject.URL
		frames[len(frames)-1].ThumbnailURL = endKeyObject.URL
	}

	return frames, nil
}

func (s *StudioService) generateFramesDemo(job domain.Job, input domain.GenerateFramesRequest, update func(int, string)) (any, error) {
	update(18, "ToonCrafter入力を準備しています")
	time.Sleep(350 * time.Millisecond)
	update(52, "フレームを補間しています")
	time.Sleep(450 * time.Millisecond)
	update(82, "タイムラインへ登録しています")
	time.Sleep(300 * time.Millisecond)

	frames := make([]domain.Frame, 0, input.FrameCount+2)
	frames = append(frames, s.newFrame(input.ProjectID, 0, domain.FrameKindKey, input.StartImageDataURL, "start keyframe"))
	for i := 0; i < input.FrameCount; i++ {
		label := fmt.Sprintf("AI %02d", i+1)
		hue := int(math.Mod(float64(178+i*24), 360))
		frames = append(frames, s.newFrame(input.ProjectID, i+1, domain.FrameKindGenerated, demoImage(label, hue), input.Prompt))
	}
	frames = append(frames, s.newFrame(input.ProjectID, input.FrameCount+1, domain.FrameKindKey, input.EndImageDataURL, "end keyframe"))

	if s.objectStore != nil {
		for index := range frames {
			prefix := "frames"
			if frames[index].Kind == domain.FrameKindKey {
				prefix = "inputs"
			}
			if err := s.storeFrameImage(context.Background(), &frames[index], prefix); err != nil {
				return nil, err
			}
		}
	}

	if err := s.store.ReplaceFrames(input.ProjectID, frames); err != nil {
		return nil, err
	}

	return domain.GenerateFramesResult{Frames: frames}, nil
}

func (s *StudioService) InpaintFrame(input domain.InpaintFrameRequest) domain.Job {
	job := s.createJob(input.ProjectID, domain.JobTypeInpainting, "Inpaintingを受け付けました")
	if job.Status == domain.JobStatusFailed {
		return job
	}

	go s.runJob(job.ID, func(update func(int, string)) (any, error) {
		if s.replicateClient != nil {
			return s.inpaintFrameWithReplicate(job, input, update)
		}
		return s.inpaintFrameDemo(job, input, update)
	})

	return job
}

func (s *StudioService) inpaintFrameWithReplicate(job domain.Job, input domain.InpaintFrameRequest, update func(int, string)) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	frame, ok, err := s.store.FindFrame(input.ProjectID, input.FrameID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("frame not found")
	}

	update(20, "Inpainting推論を開始しています")
	inpaintInput, err := replicate.BuildSDXLInpaintingInput(input, frame.ImageURL)
	if err != nil {
		return nil, fmt.Errorf("Inpainting入力の構築に失敗しました: %w", err)
	}

	prediction, err := s.replicateClient.CreatePrediction(ctx, s.sdxlInpaintingVer, inpaintInput)
	if err != nil {
		return nil, fmt.Errorf("Inpainting推論の開始に失敗しました: %w", err)
	}

	update(45, "AIが部分修正を実行しています")
	prediction, err = s.replicateClient.WaitForPrediction(ctx, prediction.ID, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("推論の待機に失敗しました: %w", err)
	}
	if prediction.Status != replicate.StatusSucceeded {
		return nil, fmt.Errorf("Inpainting推論に失敗しました: %s", prediction.Error)
	}

	resultURL := replicate.ParseSDXLInpaintingOutput(prediction)
	if resultURL == "" {
		return nil, fmt.Errorf("Inpaintingが出力を返しませんでした")
	}

	update(70, "修正結果をダウンロードしています")
	resultData, err := s.replicateClient.DownloadOutput(ctx, resultURL)
	if err != nil {
		return nil, fmt.Errorf("修正結果のダウンロードに失敗しました: %w", err)
	}

	if s.objectStore != nil {
		if _, err := s.objectStore.PutDataURL(context.Background(), maskObjectKey(input.ProjectID, job.ID, input.MaskDataURL), input.MaskDataURL); err != nil {
			return nil, fmt.Errorf("store inpainting mask object: %w", err)
		}
	}

	frame.Kind = domain.FrameKindInpainted
	frame.Note = input.Prompt
	frame.UpdatedAt = now()

	if s.objectStore != nil {
		update(85, "修正結果を保存しています")
		frameKey := fmt.Sprintf("projects/%s/frames/%s.png", input.ProjectID, input.FrameID)
		object, err := s.objectStore.PutBytes(ctx, frameKey, "image/png", resultData)
		if err != nil {
			return nil, fmt.Errorf("修正結果の保存に失敗しました: %w", err)
		}
		frame.ImageURL = object.URL
		frame.ThumbnailURL = object.URL
	} else {
		frame.ImageURL = demoImage("INPAINT", 146)
		frame.ThumbnailURL = frame.ImageURL
	}

	if err := s.store.UpsertFrame(frame); err != nil {
		return nil, err
	}

	return domain.InpaintFrameResult{Frame: frame}, nil
}

func (s *StudioService) inpaintFrameDemo(job domain.Job, input domain.InpaintFrameRequest, update func(int, string)) (any, error) {
	update(28, "マスク領域を解析しています")
	time.Sleep(300 * time.Millisecond)
	update(72, "プロンプトに沿って部分修正しています")
	time.Sleep(500 * time.Millisecond)

	frame, ok, err := s.store.FindFrame(input.ProjectID, input.FrameID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("frame not found")
	}
	if s.objectStore != nil {
		if _, err := s.objectStore.PutDataURL(context.Background(), maskObjectKey(input.ProjectID, job.ID, input.MaskDataURL), input.MaskDataURL); err != nil {
			return nil, fmt.Errorf("store inpainting mask object: %w", err)
		}
	}
	frame.Kind = domain.FrameKindInpainted
	frame.Note = input.Prompt
	frame.ImageURL = demoImage("INPAINT", 146)
	frame.ThumbnailURL = frame.ImageURL
	frame.UpdatedAt = now()
	if s.objectStore != nil {
		if err := s.storeFrameImage(context.Background(), &frame, "frames"); err != nil {
			return nil, err
		}
	}
	if err := s.store.UpsertFrame(frame); err != nil {
		return nil, err
	}

	return domain.InpaintFrameResult{Frame: frame}, nil
}

func (s *StudioService) UpdateFrame(ctx context.Context, input domain.UpdateFrameRequest) (domain.Frame, error) {
	frame, ok, err := s.store.FindFrame(input.ProjectID, input.FrameID)
	if err != nil {
		return domain.Frame{}, err
	}
	if !ok {
		return domain.Frame{}, ErrFrameNotFound
	}

	frame.Kind = domain.FrameKindEdited
	imageURL := input.ImageDataURL
	if s.objectStore != nil {
		object, err := s.objectStore.PutDataURL(ctx, frameObjectKey(input.ProjectID, input.FrameID, input.ImageDataURL), input.ImageDataURL)
		if err != nil {
			return domain.Frame{}, fmt.Errorf("store edited frame object: %w", err)
		}
		imageURL = object.URL
	}
	frame.ImageURL = imageURL
	frame.ThumbnailURL = imageURL
	frame.Note = input.Note
	frame.UpdatedAt = now()
	if err := s.store.UpsertFrame(frame); err != nil {
		return domain.Frame{}, err
	}

	return frame, nil
}

func (s *StudioService) storeFrameImage(ctx context.Context, frame *domain.Frame, prefix string) error {
	object, err := s.objectStore.PutDataURL(ctx, projectObjectKey(frame.ProjectID, prefix, frame.ID, frame.ImageURL), frame.ImageURL)
	if err != nil {
		return fmt.Errorf("store %s object: %w", prefix, err)
	}
	frame.ImageURL = object.URL
	frame.ThumbnailURL = object.URL
	return nil
}

func frameObjectKey(projectID string, frameID string, dataURL string) string {
	return projectObjectKey(projectID, "frames", frameID, dataURL)
}

func maskObjectKey(projectID string, jobID string, dataURL string) string {
	return projectObjectKey(projectID, "masks", jobID, dataURL)
}

func projectObjectKey(projectID string, prefix string, objectID string, dataURL string) string {
	extension := ".bin"
	if strings.HasPrefix(dataURL, "data:image/png") {
		extension = ".png"
	}
	if strings.HasPrefix(dataURL, "data:image/jpeg") || strings.HasPrefix(dataURL, "data:image/jpg") {
		extension = ".jpg"
	}
	if strings.HasPrefix(dataURL, "data:image/webp") {
		extension = ".webp"
	}
	if strings.HasPrefix(dataURL, "data:image/svg+xml") {
		extension = ".svg"
	}
	return fmt.Sprintf("projects/%s/%s/%s%s", projectID, prefix, objectID, extension)
}

func (s *StudioService) UpdateFrameMetadata(input domain.UpdateFrameMetadataRequest) (domain.Frame, bool, error) {
	return s.store.UpdateFrameMetadata(input)
}

func (s *StudioService) DeleteFrame(projectID string, frameID string) (bool, error) {
	return s.store.DeleteFrame(projectID, frameID)
}

func (s *StudioService) ReorderFrames(input domain.ReorderFramesRequest) ([]domain.Frame, error) {
	return s.store.ReorderFrames(input.ProjectID, input.FrameIDs)
}

func (s *StudioService) ExportVideo(input domain.ExportVideoRequest) domain.Job {
	if input.FPS <= 0 {
		input.FPS = 12
	}

	job := s.createJob(input.ProjectID, domain.JobTypeExport, "動画書き出しを受け付けました")
	if job.Status == domain.JobStatusFailed {
		return job
	}

	go s.runJob(job.ID, func(update func(int, string)) (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		update(10, "フレーム列を取得しています")
		frames, err := s.store.ListFrames(input.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("フレーム一覧の取得に失敗しました: %w", err)
		}
		if len(frames) == 0 {
			return nil, fmt.Errorf("書き出すフレームがありません")
		}

		update(30, "フレーム画像をダウンロードしています")
		tmpDir, err := os.MkdirTemp("", "anifusion-export-*")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		var framePaths []string
		for i, frame := range frames {
			ext := ".png"
			path := filepath.Join(tmpDir, fmt.Sprintf("frame_%04d%s", i+1, ext))
			data, err := s.downloadFrameImage(ctx, frame.ImageURL)
			if err != nil {
				return nil, fmt.Errorf("frame %d のダウンロードに失敗しました: %w", i+1, err)
			}
			if err := os.WriteFile(path, data, 0o644); err != nil {
				return nil, fmt.Errorf("frame %d の保存に失敗しました: %w", i+1, err)
			}
			framePaths = append(framePaths, path)
		}

		update(60, "FFmpegでMP4にエンコードしています")
		outputPath := filepath.Join(tmpDir, "output.mp4")
		if err := media.EncodeFramesToMP4(ctx, framePaths, input.FPS, outputPath); err != nil {
			return nil, fmt.Errorf("MP4エンコードに失敗しました: %w", err)
		}

		videoData, err := os.ReadFile(outputPath)
		if err != nil {
			return nil, fmt.Errorf("出力ファイルの読み取りに失敗しました: %w", err)
		}

		update(85, "書き出し結果を保存しています")
		var videoURL string
		var artifact *domain.ExportArtifact
		if s.objectStore != nil {
			exportKey := fmt.Sprintf("projects/%s/exports/%s.mp4", input.ProjectID, job.ID)
			object, err := s.objectStore.PutBytes(ctx, exportKey, "video/mp4", videoData)
			if err != nil {
				return nil, fmt.Errorf("動画の保存に失敗しました: %w", err)
			}
			videoURL = object.URL
			artifact = &domain.ExportArtifact{
				Key:         object.Key,
				URL:         object.URL,
				ContentType: object.ContentType,
				Size:        object.Size,
				FrameCount:  len(frames),
				FPS:         input.FPS,
			}
		} else {
			videoURL = "data:video/mp4;base64," + base64.StdEncoding.EncodeToString(videoData)
		}

		return domain.ExportVideoResult{
			VideoURL: videoURL,
			Artifact: artifact,
		}, nil
	})

	return job
}

func (s *StudioService) downloadFrameImage(ctx context.Context, imageURL string) ([]byte, error) {
	if strings.HasPrefix(imageURL, "data:") {
		idx := strings.Index(imageURL, "base64,")
		if idx < 0 {
			idx = strings.Index(imageURL, ",")
			if idx < 0 {
				return nil, fmt.Errorf("invalid data URL")
			}
			return []byte(imageURL[idx+1:]), nil
		}
		return base64.StdEncoding.DecodeString(imageURL[idx+7:])
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (s *StudioService) GetJob(jobID string) (domain.Job, bool, error) {
	return s.store.GetJob(jobID)
}

func (s *StudioService) createJob(projectID string, jobType domain.JobType, message string) domain.Job {
	if projectID != "" {
		_, ok, err := s.store.GetProject(projectID)
		if err != nil {
			timestamp := now()
			return domain.Job{
				ID:        fmt.Sprintf("job-%d", time.Now().UnixNano()),
				ProjectID: projectID,
				Type:      jobType,
				Status:    domain.JobStatusFailed,
				Progress:  0,
				Message:   "プロジェクト作成に失敗しました",
				Error:     err.Error(),
				CreatedAt: timestamp,
				UpdatedAt: timestamp,
			}
		}
		if !ok {
			if _, err := s.store.UpsertProject(domain.Project{ID: projectID, Name: projectID}); err != nil {
				timestamp := now()
				return domain.Job{
					ID:        fmt.Sprintf("job-%d", time.Now().UnixNano()),
					ProjectID: projectID,
					Type:      jobType,
					Status:    domain.JobStatusFailed,
					Progress:  0,
					Message:   "プロジェクト作成に失敗しました",
					Error:     err.Error(),
					CreatedAt: timestamp,
					UpdatedAt: timestamp,
				}
			}
		}
	}

	timestamp := now()
	job := domain.Job{
		ID:        fmt.Sprintf("job-%d", time.Now().UnixNano()),
		ProjectID: projectID,
		Type:      jobType,
		Status:    domain.JobStatusQueued,
		Progress:  0,
		Message:   message,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}

	if err := s.store.CreateJob(job); err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = err.Error()
		job.Message = "ジョブ作成に失敗しました"
	}

	return job
}

func (s *StudioService) runJob(jobID string, run func(update func(int, string)) (any, error)) {
	update := func(progress int, message string) {
		job, ok, err := s.store.GetJob(jobID)
		if err != nil || !ok {
			return
		}
		job.Status = domain.JobStatusRunning
		job.Progress = progress
		job.Message = message
		job.UpdatedAt = now()
		_ = s.store.UpdateJob(job)
	}

	result, err := run(update)
	job, ok, getErr := s.store.GetJob(jobID)
	if getErr != nil || !ok {
		return
	}
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = err.Error()
		job.Message = "ジョブに失敗しました"
	} else {
		job.Status = domain.JobStatusSucceeded
		job.Progress = 100
		job.Message = "完了しました"
		job.Result = result
	}
	job.UpdatedAt = now()
	_ = s.store.UpdateJob(job)
}

func (s *StudioService) newFrame(projectID string, index int, kind domain.FrameKind, imageURL string, note string) domain.Frame {
	if imageURL == "" {
		imageURL = demoImage(fmt.Sprintf("%02d", index+1), 200+index*18)
	}

	return domain.Frame{
		ID:           fmt.Sprintf("frame-%d-%d", index, time.Now().UnixNano()),
		ProjectID:    projectID,
		Index:        index,
		ImageURL:     imageURL,
		ThumbnailURL: imageURL,
		Kind:         kind,
		Note:         note,
		UpdatedAt:    now(),
	}
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func demoImage(label string, hue int) string {
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="640" height="360" viewBox="0 0 640 360"><defs><linearGradient id="g" x1="0" x2="1" y1="0" y2="1"><stop stop-color="hsl(%d,70%%,58%%)"/><stop offset="1" stop-color="hsl(%d,67%%,42%%)"/></linearGradient></defs><rect width="640" height="360" fill="url(#g)"/><circle cx="320" cy="180" r="88" fill="rgba(255,255,255,0.9)"/><text x="320" y="192" text-anchor="middle" font-family="Arial,sans-serif" font-size="42" font-weight="700" fill="#18202f">%s</text></svg>`, hue, (hue+68)%360, label)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}
