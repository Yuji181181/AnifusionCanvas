# AnifusionCanvas - Claude Code Configuration

## Project Overview

Personal demo app for anime in-between frame generation using AI.
Human-in-the-Loop workflow: AI generates, human corrects, repeat.

**4-Step Workflow:**
1. AI In-between Generation — Upload 2 keyframes, ToonCrafter interpolates intermediate frames
2. Inpainting Correction — Mask broken areas, describe fix in natural language, SDXL Inpainting regenerates
3. Manual Pixel Editing — Full Fabric.js editor for pixel-level fixes
4. Video Export — Encode frame sequence to MP4 via FFmpeg

## Development Workflow (MANDATORY)

All implementation tasks MUST follow this 3-phase workflow unless the user explicitly says to skip.

### Phase 1: Plan

- Analyze the requirement and identify affected files
- List test cases BEFORE implementation (what should be tested and why)
- Define implementation order (types → data layer → business logic → UI/API)
- Output: Implementation plan with file changes + test case list + implementation order

### Phase 2: Generate (Test-First)

- **Write tests FIRST (Red)** — create test files with assertions before implementation exists
- **Run tests to confirm they FAIL** — this proves the test actually exercises the target code
- **Implement the code (Green)** — write minimum code to make tests pass
- **Run tests after each step** — never accumulate failures
- NEVER write implementation code without a corresponding test

### Phase 3: Evaluate

- Run the full test suite
- Verify implementation matches the plan
- Check code quality and architecture compliance
- Report gaps — if critical gaps exist, loop back to Phase 2

## Monorepo Structure

```
anifusion-canvas/
├── apps/web/              # Next.js 16 App Router frontend (bun)
│   ├── app/               # Pages: /, /step1, /step2, /step3, /export
│   ├── components/        # ui/ (shadcn), editor/, timeline/, forms/
│   ├── hooks/             # Custom React hooks
│   ├── lib/               # api.ts, utils.ts
│   ├── stores/            # Zustand stores
│   ├── types/             # TypeScript type definitions
│   ├── tests/             # unit/ and e2e/
│   └── public/            # Static assets
├── apps/api/              # FastAPI backend (uv)
│   ├── app/
│   │   ├── main.py        # FastAPI entry point
│   │   ├── config.py      # pydantic-settings
│   │   ├── models/        # SQLAlchemy ORM models
│   │   ├── schemas/       # Pydantic request/response schemas
│   │   ├── routers/       # API route handlers
│   │   ├── services/      # Business logic layer
│   │   ├── infrastructure/ # database.py, storage.py, replicate.py
│   │   └── workers/       # Background task definitions
│   ├── alembic/           # Database migrations
│   └── tests/             # unit/ and integration/
├── docs/                  # Requirements and setup documentation
└── package.json           # Root workspace config (bun workspaces)
```

## Tech Stack

| Layer | Technology | Notes |
|-------|-----------|-------|
| Frontend Framework | Next.js 16 (App Router) | **Breaking changes** — read `node_modules/next/dist/docs/` before coding |
| Package Manager | bun | `bun install`, `bun add`, `bun run dev` |
| UI Components | shadcn/ui + @base-ui/react | Add via `bunx shadcn@latest add <component>` |
| Icons | lucide-react | Only icon library |
| Styling | TailwindCSS v4 | PostCSS plugin, no tailwind.config.ts |
| State Management | Zustand | `create()`, multiple small stores |
| Data Fetching | useSWR v2 | Polling for async jobs, mutate for revalidation |
| Forms | React Hook Form + Zod | `@hookform/resolvers` for zodResolver |
| Canvas | Fabric.js v7 | Step2: mask brush only; Step3: full editor |
| Lint/Format | Biome | No ESLint. `bun run lint`, `bun run format` |
| Unit Test (FE) | Vitest | jsdom, globals: true |
| E2E Test | Playwright | Chromium primary |
| Backend Framework | FastAPI | Async throughout |
| Python Package Mgr | uv | `uv add <pkg>`, NOT pip |
| ORM | SQLAlchemy 2.x (async) | aiomysql driver |
| Database | TiDB Serverless | MySQL-compatible, SSL required |
| Migrations | Alembic | Async migrations |
| Object Storage | Cloudflare R2 | S3-compatible via boto3 |
| AI Inference | Replicate API | fofr/tooncrafter, lucataco/sdxl-inpainting |
| Video Processing | FFmpeg | Frame extraction and MP4 encoding |

## Development Commands

```bash
# Frontend
cd apps/web
bun install
bun run dev          # http://localhost:3000
bun run lint         # Biome check
bun run format       # Biome format --write
bun run check        # Biome check + build
bun run test         # Vitest (watch)
bun run test:run     # Vitest (single run)
bun run test:e2e     # Playwright

# Backend
cd apps/api
uv sync
uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
uv add <package>     # Add dependency
uv run pytest        # Run tests
```

## Environment Variables

**Frontend (apps/web/.env.local):**
- `NEXT_PUBLIC_API_URL` — FastAPI base URL (default: http://localhost:8000)

**Backend (apps/api/.env):**
- `DATABASE_URL` — TiDB connection string
- `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`, `R2_PUBLIC_BASE_URL`, `R2_ENDPOINT_URL`, `R2_REGION`
- `REPLICATE_API_TOKEN` — Required for AI inference
- `HF_TOKEN`, `MODEL_CACHE_DIR`, `TOONCRAFTER_MODEL_ID`, `SD_INPAINT_MODEL_ID`

## AI Inference Costs

| Model | Rate | Per-call |
|-------|------|----------|
| ToonCrafter (A100) | $0.0014/sec | ~$0.064 |
| SDXL Inpainting (L40S) | $0.000975/sec | ~$0.003 |
| 1 session total | | ~$0.07 (~7 JPY) |

## Code Style

**TypeScript/React (Biome):**
- Tabs, double quotes, semicolons, line width 100
- `@/` path alias for imports
- Import organization enabled

**Python:**
- PEP 8 with type hints on all signatures
- `async def` for all I/O-bound functions
- Pydantic v2: `model_config = ConfigDict(...)`
- SQLAlchemy 2.0: `Mapped[]`, `mapped_column()`, no `Column()`

## Test Strategy (Pragmatic)

No coverage thresholds. Focus on critical business logic.

**Frontend (Vitest):**
- Unit: hooks, stores, lib utilities, Zod schemas
- Component: non-trivial components with logic (skip pure layout)
- Mock `lib/api.ts` in unit tests

**Frontend (Playwright):**
- E2E: 4-step critical path only (inbetween → inpaint → edit → export)

**Backend (pytest):**
- Unit: services, schemas, config, infrastructure (mock Replicate/R2)
- Integration: API endpoints with test DB (SQLite in-memory, NOT mocked DB)
- Mock ONLY external services (Replicate, R2); use real DB with fixtures
