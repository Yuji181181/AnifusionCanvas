-- name: UpsertProject :exec
INSERT INTO projects (
  id,
  name
) VALUES (?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name);
