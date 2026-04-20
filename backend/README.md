# Backend (FastAPI + uv)

## Setup

```bash
uv sync
cp .env.example .env
```

## Run

```bash
uv run uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

## Available Endpoints

- `GET /health`
- `POST /v1/jobs`
- `GET /v1/jobs/{job_id}`
- `POST /v1/jobs/{job_id}/export`

## Pre-development Helpers

```bash
# External services and env preflight
uv run python scripts/preflight_check.py

# Download model artifacts from Hugging Face
uv run python scripts/download_models.py
```

## Faster Model Download (optional)

```bash
export ARIA2_SPLIT=4
export ARIA2_MAX_CONNECTION_PER_SERVER=4
export ARIA2_MAX_CONCURRENT_DOWNLOADS=1
export ARIA2_RETRY_WAIT=5
export ARIA2_MAX_TRIES=20
export ARIA2_TIMEOUT=30
uv run python scripts/download_models.py
```
