-- name: UpsertProject :exec
INSERT INTO studio_projects (
  id,
  name
) VALUES (?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name);
