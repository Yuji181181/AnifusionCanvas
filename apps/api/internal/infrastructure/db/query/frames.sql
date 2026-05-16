-- name: ListFramesByProject :many
SELECT
  id,
  project_id,
  frame_index,
  image_url,
  thumbnail_url,
  kind,
  note,
  updated_at
FROM frames
WHERE project_id = ?
ORDER BY frame_index ASC;

-- name: UpsertFrame :exec
INSERT INTO frames (
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
  note = VALUES(note);
