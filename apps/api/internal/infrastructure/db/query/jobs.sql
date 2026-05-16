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
FROM studio_jobs
WHERE id = ?;

-- name: CreateJob :exec
INSERT INTO studio_jobs (
  id,
  project_id,
  job_type,
  status,
  progress,
  message
) VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateJobState :exec
UPDATE studio_jobs
SET
  status = ?,
  progress = ?,
  message = ?,
  result_json = ?,
  error = ?
WHERE id = ?;
