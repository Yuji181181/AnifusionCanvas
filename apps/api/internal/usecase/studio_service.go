package usecase

import (
	"encoding/base64"
	"fmt"
	"math"
	"time"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

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

type StudioService struct {
	store StudioStore
}

func NewStudioService() *StudioService {
	return NewStudioServiceWithStore(NewMemoryStudioStore())
}

func NewStudioServiceWithStore(store StudioStore) *StudioService {
	return &StudioService{store: store}
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

	job := s.createJob(input.ProjectID, "generation", "中割り生成を受け付けました")
	if job.Status == domain.JobStatusFailed {
		return job
	}

	go s.runJob(job.ID, func(update func(int, string)) (any, error) {
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

		if err := s.store.ReplaceFrames(input.ProjectID, frames); err != nil {
			return nil, err
		}

		return domain.GenerateFramesResult{Frames: frames}, nil
	})

	return job
}

func (s *StudioService) InpaintFrame(input domain.InpaintFrameRequest) domain.Job {
	job := s.createJob(input.ProjectID, "inpainting", "Inpaintingを受け付けました")
	if job.Status == domain.JobStatusFailed {
		return job
	}

	go s.runJob(job.ID, func(update func(int, string)) (any, error) {
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
		frame.Kind = domain.FrameKindInpainted
		frame.Note = input.Prompt
		frame.ImageURL = demoImage("INPAINT", 146)
		frame.ThumbnailURL = frame.ImageURL
		frame.UpdatedAt = now()
		if err := s.store.UpsertFrame(frame); err != nil {
			return nil, err
		}

		return domain.InpaintFrameResult{Frame: frame}, nil
	})

	return job
}

func (s *StudioService) UpdateFrame(input domain.UpdateFrameRequest) (domain.Frame, error) {
	frame, ok, err := s.store.FindFrame(input.ProjectID, input.FrameID)
	if err != nil {
		return domain.Frame{}, err
	}
	if !ok {
		return domain.Frame{}, fmt.Errorf("frame not found")
	}

	frame.Kind = domain.FrameKindEdited
	frame.ImageURL = input.ImageDataURL
	frame.ThumbnailURL = input.ImageDataURL
	frame.Note = input.Note
	frame.UpdatedAt = now()
	if err := s.store.UpsertFrame(frame); err != nil {
		return domain.Frame{}, err
	}

	return frame, nil
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
	job := s.createJob(input.ProjectID, "export", "動画書き出しを受け付けました")
	if job.Status == domain.JobStatusFailed {
		return job
	}

	go s.runJob(job.ID, func(update func(int, string)) (any, error) {
		update(24, "フレーム列を確認しています")
		time.Sleep(250 * time.Millisecond)
		update(66, "FFmpegエンコードを模擬実行しています")
		time.Sleep(550 * time.Millisecond)
		update(92, "書き出し結果を登録しています")
		time.Sleep(250 * time.Millisecond)

		return domain.ExportVideoResult{VideoURL: "data:text/plain;base64,RGVtbyBleHBvcnQgaXMgc3VjY2Vzc2Z1bC4="}, nil
	})

	return job
}

func (s *StudioService) GetJob(jobID string) (domain.Job, bool, error) {
	return s.store.GetJob(jobID)
}

func (s *StudioService) createJob(projectID string, jobType string, message string) domain.Job {
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
