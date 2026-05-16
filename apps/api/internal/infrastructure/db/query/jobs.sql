-- name: GetJob :one
SELECT
  id,
  project_id,
  job_type,
  status,
  progress,
  message,
  result_json,
  error,
  created_at,
  updated_at
FROM jobs
WHERE id = ?;

-- name: CreateJob :exec
INSERT INTO jobs (
  id,
  project_id,
  job_type,
  status,
  progress,
  message
) VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateJobState :exec
UPDATE jobs
SET
  status = ?,
  progress = ?,
  message = ?,
  result_json = ?,
  error = ?
WHERE id = ?;
