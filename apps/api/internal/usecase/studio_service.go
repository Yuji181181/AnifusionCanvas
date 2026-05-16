package usecase

import (
	"encoding/base64"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

type StudioService struct {
	mu     sync.RWMutex
	frames map[string][]domain.Frame
	jobs   map[string]domain.Job
}

func NewStudioService() *StudioService {
	return &StudioService{
		frames: make(map[string][]domain.Frame),
		jobs:   make(map[string]domain.Job),
	}
}

func (s *StudioService) ListFrames(projectID string) []domain.Frame {
	s.mu.RLock()
	defer s.mu.RUnlock()

	frames := s.frames[projectID]
	if frames == nil {
		return []domain.Frame{}
	}

	return append([]domain.Frame(nil), frames...)
}

func (s *StudioService) GenerateFrames(input domain.GenerateFramesRequest) domain.Job {
	if input.FrameCount < 2 {
		input.FrameCount = 2
	}
	if input.FrameCount > 12 {
		input.FrameCount = 12
	}

	job := s.createJob("generation", "中割り生成を受け付けました")
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

		s.mu.Lock()
		s.frames[input.ProjectID] = frames
		s.mu.Unlock()

		return domain.GenerateFramesResult{Frames: frames}, nil
	})

	return job
}

func (s *StudioService) InpaintFrame(input domain.InpaintFrameRequest) domain.Job {
	job := s.createJob("inpainting", "Inpaintingを受け付けました")
	go s.runJob(job.ID, func(update func(int, string)) (any, error) {
		update(28, "マスク領域を解析しています")
		time.Sleep(300 * time.Millisecond)
		update(72, "プロンプトに沿って部分修正しています")
		time.Sleep(500 * time.Millisecond)

		frame, ok := s.findFrame(input.ProjectID, input.FrameID)
		if !ok {
			return nil, fmt.Errorf("frame not found")
		}
		frame.Kind = domain.FrameKindInpainted
		frame.Note = input.Prompt
		frame.ImageURL = demoImage("INPAINT", 146)
		frame.ThumbnailURL = frame.ImageURL
		frame.UpdatedAt = now()
		s.replaceFrame(frame)

		return domain.InpaintFrameResult{Frame: frame}, nil
	})

	return job
}

func (s *StudioService) UpdateFrame(input domain.UpdateFrameRequest) (domain.Frame, error) {
	frame, ok := s.findFrame(input.ProjectID, input.FrameID)
	if !ok {
		return domain.Frame{}, fmt.Errorf("frame not found")
	}

	frame.Kind = domain.FrameKindEdited
	frame.ImageURL = input.ImageDataURL
	frame.ThumbnailURL = input.ImageDataURL
	frame.Note = input.Note
	frame.UpdatedAt = now()
	s.replaceFrame(frame)

	return frame, nil
}

func (s *StudioService) ExportVideo(input domain.ExportVideoRequest) domain.Job {
	job := s.createJob("export", "動画書き出しを受け付けました")
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

func (s *StudioService) GetJob(jobID string) (domain.Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[jobID]
	return job, ok
}

func (s *StudioService) createJob(jobType string, message string) domain.Job {
	timestamp := now()
	job := domain.Job{
		ID:        fmt.Sprintf("job-%d", time.Now().UnixNano()),
		Type:      jobType,
		Status:    domain.JobStatusQueued,
		Progress:  0,
		Message:   message,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}

	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	return job
}

func (s *StudioService) runJob(jobID string, run func(update func(int, string)) (any, error)) {
	update := func(progress int, message string) {
		s.mu.Lock()
		job := s.jobs[jobID]
		job.Status = domain.JobStatusRunning
		job.Progress = progress
		job.Message = message
		job.UpdatedAt = now()
		s.jobs[jobID] = job
		s.mu.Unlock()
	}

	result, err := run(update)

	s.mu.Lock()
	defer s.mu.Unlock()
	job := s.jobs[jobID]
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
	s.jobs[jobID] = job
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

func (s *StudioService) findFrame(projectID string, frameID string) (domain.Frame, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, frame := range s.frames[projectID] {
		if frame.ID == frameID {
			return frame, true
		}
	}

	return domain.Frame{}, false
}

func (s *StudioService) replaceFrame(next domain.Frame) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frames := s.frames[next.ProjectID]
	for index, frame := range frames {
		if frame.ID == next.ID {
			frames[index] = next
			break
		}
	}
	s.frames[next.ProjectID] = frames
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func demoImage(label string, hue int) string {
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="640" height="360" viewBox="0 0 640 360"><defs><linearGradient id="g" x1="0" x2="1" y1="0" y2="1"><stop stop-color="hsl(%d,70%%,58%%)"/><stop offset="1" stop-color="hsl(%d,67%%,42%%)"/></linearGradient></defs><rect width="640" height="360" fill="url(#g)"/><circle cx="320" cy="180" r="88" fill="rgba(255,255,255,0.9)"/><text x="320" y="192" text-anchor="middle" font-family="Arial,sans-serif" font-size="42" font-weight="700" fill="#18202f">%s</text></svg>`, hue, (hue+68)%360, label)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}
