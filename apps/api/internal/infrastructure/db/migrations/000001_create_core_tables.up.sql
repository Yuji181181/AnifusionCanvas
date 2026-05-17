CREATE TABLE studio_projects (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE studio_frames (
  id VARCHAR(96) PRIMARY KEY,
  project_id VARCHAR(64) NOT NULL,
  frame_index INT NOT NULL,
  image_url MEDIUMTEXT NOT NULL,
  thumbnail_url MEDIUMTEXT NOT NULL,
  kind VARCHAR(32) NOT NULL,
  note TEXT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_studio_frames_project_index (project_id, frame_index),
  KEY idx_studio_frames_project_id (project_id)
);

CREATE TABLE studio_jobs (
  id VARCHAR(96) PRIMARY KEY,
  project_id VARCHAR(64) NULL,
  job_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  progress INT NOT NULL DEFAULT 0,
  message TEXT NOT NULL,
  result_json JSON NULL,
  error TEXT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_studio_jobs_project_id (project_id),
  KEY idx_studio_jobs_status (status)
);
