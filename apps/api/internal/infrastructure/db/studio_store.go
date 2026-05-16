package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

type StudioStore struct {
	db *sql.DB
}

func NewStudioStore(databaseURL string) (*StudioStore, error) {
	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &StudioStore{db: db}, nil
}

func (s *StudioStore) ListFrames(projectID string) ([]domain.Frame, error) {
	rows, err := s.db.Query(`
SELECT
  id,
  project_id,
  frame_index,
  image_url,
  thumbnail_url,
  kind,
  COALESCE(note, ''),
  DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM studio_frames
WHERE project_id = ?
ORDER BY frame_index ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	frames := make([]domain.Frame, 0)
	for rows.Next() {
		var frame domain.Frame
		var kind string
		if err := rows.Scan(
			&frame.ID,
			&frame.ProjectID,
			&frame.Index,
			&frame.ImageURL,
			&frame.ThumbnailURL,
			&kind,
			&frame.Note,
			&frame.UpdatedAt,
		); err != nil {
			return nil, err
		}
		frame.Kind = domain.FrameKind(kind)
		frames = append(frames, frame)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return frames, nil
}

func (s *StudioStore) UpsertProject(project domain.Project) (domain.Project, error) {
	if err := upsertStudioProject(s.db, project.ID, project.Name); err != nil {
		return domain.Project{}, err
	}

	created, ok, err := s.GetProject(project.ID)
	if err != nil {
		return domain.Project{}, err
	}
	if !ok {
		return domain.Project{}, fmt.Errorf("project not found after upsert")
	}

	return created, nil
}

func (s *StudioStore) GetProject(projectID string) (domain.Project, bool, error) {
	var project domain.Project
	err := s.db.QueryRow(`
SELECT
  id,
  name,
  DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'),
  DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM studio_projects
WHERE id = ?`, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Project{}, false, nil
	}
	if err != nil {
		return domain.Project{}, false, err
	}

	return project, true, nil
}

func (s *StudioStore) UpdateProject(project domain.Project) (domain.Project, bool, error) {
	current, ok, err := s.GetProject(project.ID)
	if err != nil {
		return domain.Project{}, false, err
	}
	if !ok {
		return domain.Project{}, false, nil
	}
	if current.Name == project.Name {
		return current, true, nil
	}

	_, err = s.db.Exec(`
UPDATE studio_projects
SET name = ?
WHERE id = ?`, project.Name, project.ID)
	if err != nil {
		return domain.Project{}, false, err
	}

	updated, ok, err := s.GetProject(project.ID)
	if err != nil {
		return domain.Project{}, false, err
	}
	return updated, ok, nil
}

func (s *StudioStore) ReplaceFrames(projectID string, frames []domain.Frame) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := upsertProjectTx(tx, projectID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM studio_frames WHERE project_id = ?`, projectID); err != nil {
		return err
	}
	for _, frame := range frames {
		if err := upsertFrame(tx, frame); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *StudioStore) FindFrame(projectID string, frameID string) (domain.Frame, bool, error) {
	var frame domain.Frame
	var kind string
	err := s.db.QueryRow(`
SELECT
  id,
  project_id,
  frame_index,
  image_url,
  thumbnail_url,
  kind,
  COALESCE(note, ''),
  DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM studio_frames
WHERE project_id = ? AND id = ?`, projectID, frameID).Scan(
		&frame.ID,
		&frame.ProjectID,
		&frame.Index,
		&frame.ImageURL,
		&frame.ThumbnailURL,
		&kind,
		&frame.Note,
		&frame.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Frame{}, false, nil
	}
	if err != nil {
		return domain.Frame{}, false, err
	}
	frame.Kind = domain.FrameKind(kind)

	return frame, true, nil
}

func (s *StudioStore) UpsertFrame(frame domain.Frame) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := upsertProjectTx(tx, frame.ProjectID); err != nil {
		return err
	}
	if err := upsertFrame(tx, frame); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *StudioStore) UpdateFrameMetadata(input domain.UpdateFrameMetadataRequest) (domain.Frame, bool, error) {
	frame, ok, err := s.FindFrame(input.ProjectID, input.FrameID)
	if err != nil {
		return domain.Frame{}, false, err
	}
	if !ok {
		return domain.Frame{}, false, nil
	}
	if input.Kind != nil {
		frame.Kind = *input.Kind
	}
	if input.Note != nil {
		frame.Note = *input.Note
	}
	if err := s.UpsertFrame(frame); err != nil {
		return domain.Frame{}, false, err
	}

	updated, ok, err := s.FindFrame(input.ProjectID, input.FrameID)
	if err != nil {
		return domain.Frame{}, false, err
	}
	return updated, ok, nil
}

func (s *StudioStore) DeleteFrame(projectID string, frameID string) (bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`DELETE FROM studio_frames WHERE project_id = ? AND id = ?`, projectID, frameID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}

	rows, err := tx.Query(`
SELECT id
FROM studio_frames
WHERE project_id = ?
ORDER BY frame_index ASC`, projectID)
	if err != nil {
		return false, err
	}
	frameIDs := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return false, err
		}
		frameIDs = append(frameIDs, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return false, err
	}
	rows.Close()

	if err := reorderFramesTx(tx, projectID, frameIDs); err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (s *StudioStore) ReorderFrames(projectID string, frameIDs []string) ([]domain.Frame, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM studio_frames WHERE project_id = ?`, projectID).Scan(&count); err != nil {
		return nil, err
	}
	if count != len(frameIDs) {
		return nil, fmt.Errorf("frameIds must include every frame in the project")
	}
	seen := make(map[string]struct{}, len(frameIDs))
	for _, frameID := range frameIDs {
		if _, ok := seen[frameID]; ok {
			return nil, fmt.Errorf("frameIds must not contain duplicate values")
		}
		seen[frameID] = struct{}{}
		var exists int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM studio_frames WHERE project_id = ? AND id = ?`, projectID, frameID).Scan(&exists); err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, fmt.Errorf("frame not found: %s", frameID)
		}
	}
	if err := reorderFramesTx(tx, projectID, frameIDs); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.ListFrames(projectID)
}

func (s *StudioStore) CreateJob(job domain.Job) error {
	_, err := s.db.Exec(`
INSERT INTO studio_jobs (
  id,
  project_id,
  job_type,
  status,
  progress,
  message
) VALUES (?, ?, ?, ?, ?, ?)`,
		job.ID,
		nullString(job.ProjectID),
		string(job.Type),
		string(job.Status),
		job.Progress,
		job.Message,
	)
	return err
}

func (s *StudioStore) UpdateJob(job domain.Job) error {
	resultJSON, err := json.Marshal(job.Result)
	if err != nil {
		return err
	}
	if string(resultJSON) == "null" {
		resultJSON = nil
	}

	_, err = s.db.Exec(`
UPDATE studio_jobs
SET
  status = ?,
  progress = ?,
  message = ?,
  result_json = ?,
  error = ?
WHERE id = ?`,
		string(job.Status),
		job.Progress,
		job.Message,
		nullBytes(resultJSON),
		nullString(job.Error),
		job.ID,
	)
	return err
}

func (s *StudioStore) GetJob(jobID string) (domain.Job, bool, error) {
	var job domain.Job
	var projectID sql.NullString
	var jobType string
	var status string
	var resultJSON []byte
	var errorMessage sql.NullString
	err := s.db.QueryRow(`
SELECT
  id,
  project_id,
  job_type,
  status,
  progress,
  message,
  result_json,
  error,
  DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'),
  DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM studio_jobs
WHERE id = ?`, jobID).Scan(
		&job.ID,
		&projectID,
		&jobType,
		&status,
		&job.Progress,
		&job.Message,
		&resultJSON,
		&errorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Job{}, false, nil
	}
	if err != nil {
		return domain.Job{}, false, err
	}

	job.ProjectID = projectID.String
	job.Type = domain.JobType(jobType)
	job.Status = domain.JobStatus(status)
	job.Error = errorMessage.String
	if len(resultJSON) > 0 {
		result, err := decodeJobResult(job.Type, resultJSON)
		if err != nil {
			return domain.Job{}, false, fmt.Errorf("decode job result: %w", err)
		}
		job.Result = result
	}

	return job, true, nil
}

func decodeJobResult(jobType domain.JobType, payload []byte) (any, error) {
	switch jobType {
	case domain.JobTypeGeneration:
		var result domain.GenerateFramesResult
		if err := json.Unmarshal(payload, &result); err != nil {
			return nil, err
		}
		return result, nil
	case domain.JobTypeInpainting:
		var result domain.InpaintFrameResult
		if err := json.Unmarshal(payload, &result); err != nil {
			return nil, err
		}
		return result, nil
	case domain.JobTypeExport:
		var result domain.ExportVideoResult
		if err := json.Unmarshal(payload, &result); err != nil {
			return nil, err
		}
		return result, nil
	default:
		var result map[string]any
		if err := json.Unmarshal(payload, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
}

func upsertProjectTx(tx *sql.Tx, projectID string) error {
	_, err := tx.Exec(`
INSERT INTO studio_projects (
  id,
  name
) VALUES (?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name)`, projectID, projectID)
	return err
}

func upsertStudioProject(db *sql.DB, projectID string, name string) error {
	if name == "" {
		name = projectID
	}
	_, err := db.Exec(`
INSERT INTO studio_projects (
  id,
  name
) VALUES (?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name)`, projectID, name)
	return err
}

func upsertFrame(tx *sql.Tx, frame domain.Frame) error {
	_, err := tx.Exec(`
INSERT INTO studio_frames (
  id,
  project_id,
  frame_index,
  image_url,
  thumbnail_url,
  kind,
  note
) VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  image_url = VALUES(image_url),
  thumbnail_url = VALUES(thumbnail_url),
  kind = VALUES(kind),
  note = VALUES(note)`,
		frame.ID,
		frame.ProjectID,
		frame.Index,
		frame.ImageURL,
		frame.ThumbnailURL,
		string(frame.Kind),
		nullString(frame.Note),
	)
	return err
}

func reorderFramesTx(tx *sql.Tx, projectID string, frameIDs []string) error {
	for index, frameID := range frameIDs {
		if _, err := tx.Exec(`
UPDATE studio_frames
SET frame_index = ?
WHERE project_id = ? AND id = ?`, -100000-index, projectID, frameID); err != nil {
			return err
		}
	}
	for index, frameID := range frameIDs {
		if _, err := tx.Exec(`
UPDATE studio_frames
SET frame_index = ?
WHERE project_id = ? AND id = ?`, index, projectID, frameID); err != nil {
			return err
		}
	}
	return nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}

	return value
}
