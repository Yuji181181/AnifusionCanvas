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
