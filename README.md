# AnifusionCanvas

アニメーション制作における「中割り」工程をAIで支援する Human-in-the-Loop デモアプリケーションです。最新の生成AI（ToonCrafter / SDXL Inpainting）とモダンWeb技術を統合し、クリエイターがAIの生成結果を随時修正できる3ステップのワークフローを提供します。

## ワークフロー

1. **AI中割り生成** — 2枚の原画から間のフレームを自動補間（ToonCrafter）
2. **Inpainting修正** — 破綻箇所をマスクし自然言語で再生成指示（SDXL Inpainting）
3. **手動編集** — ピクセル単位の加筆修正（Fabric.jsフル機能エディタ）
4. **動画書き出し** — フレーム列をMP4にエンコード（FFmpeg）

## Tech Stack

| 領域 | 技術 |
|------|------|
| Frontend | Next.js (App Router) + bun + TailwindCSS |
| UI Components | shadcn/ui + Radix UI |
| State Management | Zustand |
| Data Fetching | useSWR |
| Form | React Hook Form + Zod |
| Canvas | Fabric.js |
| Lint/Format | Biome |
| Unit Test | Vitest |
| E2E Test | Playwright |
| Backend | FastAPI + uv |
| Database | TiDB Serverless |
| Storage | Cloudflare R2 |
| AI Inference | Replicate API (ToonCrafter / SDXL Inpainting) |

## Directory

```
anifusion-canvas/
├── apps/
│   ├── web/          # Next.js フロントエンド
│   └── api/          # FastAPI バックエンド
├── docs/             # ドキュメント
└── package.json      # ルート (bun workspaces)
```

## Prerequisites

- bun
- uv
- Python 3.11+
- ffmpeg

## Quick Start

### 1) Frontend

```bash
cd apps/web
cp .env.example .env.local
bun install
bun run dev
```

### 2) Backend

```bash
cd apps/api
cp .env.example .env
uv sync
uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

- Frontend: http://localhost:3000
- Backend: http://localhost:8000
- API docs: http://localhost:8000/docs

### 3) Replicate API トークン設定

AI推論機能を利用するにはReplicateのAPIトークンが必要です。手順は [setup-replicate-api.md](docs/setup-replicate-api.md) を参照してください。

```bash
# apps/api/.env にトークンを設定
REPLICATE_API_TOKEN=r8_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

## AI推論コスト目安

| モデル | 単価 | 1回あたり |
|--------|------|-----------|
| ToonCrafter (A100) | $0.0014/秒 | ~$0.064 |
| SDXL Inpainting (L40S) | $0.000975/秒 | ~$0.003 |
| **1セッション合計** | | **~$0.07 (約7円)** |

## 詳細

- [アプリケーション要件定義書](docs/アプリケーション要件定義書.md)
- [Replicate API トークン設定手順](docs/setup-replicate-api.md)
