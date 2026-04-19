# AnifusionCanvas

アニメ制作向け Human-in-the-Loop 中割り支援デモの開発環境です。

## Directory

- `frontend`: Next.js (App Router) + bun + TailwindCSS + Biome
- `backend`: FastAPI + uv

## Prerequisites

- bun
- uv
- Python 3.11+
- ffmpeg (将来の動画エンコード用)

## Quick Start

### 1) Frontend

```bash
cd frontend
cp .env.example .env.local
bun install
bun run dev
```

### 2) Backend

```bash
cd backend
cp .env.example .env
uv sync
uv run uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

- Frontend: http://localhost:3000
- Backend: http://localhost:8000
- API docs: http://localhost:8000/docs
