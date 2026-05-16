package usecase

import (
	"fmt"
	"sync"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

type MemoryStudioStore struct {
	mu       sync.RWMutex
	projects map[string]domain.Project
	frames   map[string][]domain.Frame
	jobs     map[string]domain.Job
}

func NewMemoryStudioStore() *MemoryStudioStore {
	return &MemoryStudioStore{
		projects: make(map[string]domain.Project),
		frames:   make(map[string][]domain.Frame),
		jobs:     make(map[string]domain.Job),
	}
}

func (s *MemoryStudioStore) UpsertProject(project domain.Project) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.projects[project.ID]
	timestamp := now()
	if !exists {
		current = domain.Project{
			ID:        project.ID,
			CreatedAt: timestamp,
		}
	}
	current.Name = project.Name
	if current.Name == "" {
		current.Name = project.ID
	}
	current.UpdatedAt = timestamp
	if current.CreatedAt == "" {
		current.CreatedAt = timestamp
	}
	s.projects[current.ID] = current
	return current, nil
}

func (s *MemoryStudioStore) GetProject(projectID string) (domain.Project, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, ok := s.projects[projectID]
	return project, ok, nil
}

func (s *MemoryStudioStore) UpdateProject(project domain.Project) (domain.Project, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.projects[project.ID]
	if !ok {
		return domain.Project{}, false, nil
	}
	current.Name = project.Name
	current.UpdatedAt = now()
	s.projects[current.ID] = current
	return current, true, nil
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

func (s *MemoryStudioStore) UpdateFrameMetadata(input domain.UpdateFrameMetadataRequest) (domain.Frame, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frames := s.frames[input.ProjectID]
	for index, frame := range frames {
		if frame.ID == input.FrameID {
			if input.Kind != nil {
				frame.Kind = *input.Kind
			}
			if input.Note != nil {
				frame.Note = *input.Note
			}
			frame.UpdatedAt = now()
			frames[index] = frame
			s.frames[input.ProjectID] = frames
			return frame, true, nil
		}
	}

	return domain.Frame{}, false, nil
}

func (s *MemoryStudioStore) DeleteFrame(projectID string, frameID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frames := s.frames[projectID]
	next := frames[:0]
	deleted := false
	for _, frame := range frames {
		if frame.ID == frameID {
			deleted = true
			continue
		}
		frame.Index = len(next)
		frame.UpdatedAt = now()
		next = append(next, frame)
	}
	if !deleted {
		return false, nil
	}
	s.frames[projectID] = append([]domain.Frame(nil), next...)
	return true, nil
}

func (s *MemoryStudioStore) ReorderFrames(projectID string, frameIDs []string) ([]domain.Frame, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frames := s.frames[projectID]
	byID := make(map[string]domain.Frame, len(frames))
	for _, frame := range frames {
		byID[frame.ID] = frame
	}

	reordered := make([]domain.Frame, 0, len(frameIDs))
	for index, frameID := range frameIDs {
		frame, ok := byID[frameID]
		if !ok {
			return nil, fmt.Errorf("frame not found: %s", frameID)
		}
		frame.Index = index
		frame.UpdatedAt = now()
		reordered = append(reordered, frame)
		delete(byID, frameID)
	}
	if len(byID) > 0 {
		return nil, fmt.Errorf("frameIds must include every frame in the project")
	}

	s.frames[projectID] = reordered
	return append([]domain.Frame(nil), reordered...), nil
}

func (s *MemoryStudioStore) CreateJob(job domain.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
	return nil
}

func (s *MemoryStudioStore) UpdateJob(job domain.Job) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.jobs[job.ID]
	if !ok {
		return false, nil
	}
	if existing.Version != job.Version {
		return false, nil
	}
	job.Version++
	s.jobs[job.ID] = job
	return true, nil
}

func (s *MemoryStudioStore) GetJob(jobID string) (domain.Job, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[jobID]
	return job, ok, nil
}
