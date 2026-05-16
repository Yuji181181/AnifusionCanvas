package usecase

import (
	"sync"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

type MemoryStudioStore struct {
	mu     sync.RWMutex
	frames map[string][]domain.Frame
	jobs   map[string]domain.Job
}

func NewMemoryStudioStore() *MemoryStudioStore {
	return &MemoryStudioStore{
		frames: make(map[string][]domain.Frame),
		jobs:   make(map[string]domain.Job),
	}
}

func (s *MemoryStudioStore) ListFrames(projectID string) ([]domain.Frame, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	frames := s.frames[projectID]
	if frames == nil {
		return []domain.Frame{}, nil
	}

	return append([]domain.Frame(nil), frames...), nil
}

func (s *MemoryStudioStore) ReplaceFrames(projectID string, frames []domain.Frame) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.frames[projectID] = append([]domain.Frame(nil), frames...)
	return nil
}

func (s *MemoryStudioStore) FindFrame(projectID string, frameID string) (domain.Frame, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, frame := range s.frames[projectID] {
		if frame.ID == frameID {
			return frame, true, nil
		}
	}

	return domain.Frame{}, false, nil
}

func (s *MemoryStudioStore) UpsertFrame(next domain.Frame) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	frames := s.frames[next.ProjectID]
	for index, frame := range frames {
		if frame.ID == next.ID {
			frames[index] = next
			s.frames[next.ProjectID] = frames
			return nil
		}
	}

	s.frames[next.ProjectID] = append(frames, next)
	return nil
}

func (s *MemoryStudioStore) CreateJob(job domain.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
	return nil
}

func (s *MemoryStudioStore) UpdateJob(job domain.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
	return nil
}

func (s *MemoryStudioStore) GetJob(jobID string) (domain.Job, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[jobID]
	return job, ok, nil
}
