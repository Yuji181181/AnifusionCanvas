# Backend Conventions (apps/api)

## Layered Architecture

```
Router → Service → Infrastructure
```

- **Routers** (`routers/`): HTTP request handling only. Call services, return responses.
- **Services** (`services/`): Business logic. Call infrastructure, operate on models.
- **Infrastructure** (`infrastructure/`): External I/O — database, storage, Replicate API.

Routers must NOT import infrastructure directly. Services must NOT import routers.

## Async Everywhere

- All DB operations use `AsyncSession` via `get_db()` dependency
- All external API calls use `httpx.AsyncClient` or `replicate.async_run`
- Never use sync DB drivers or sync HTTP clients

## Configuration

- `app/config.py` uses `pydantic-settings` with `.env` file
- All secrets via environment variables — never hardcode credentials
- Settings instance available via FastAPI dependency injection

## Database

- SQLAlchemy models in `models/` (DB representation)
- Pydantic schemas in `schemas/` (API contracts)
- Keep them separate — models for DB, schemas for API
- Alembic for migrations — always generate and review before applying
- Migration command: `PYTHONPATH=. .venv/bin/alembic revision --autogenerate -m "description"`

## Job Pattern (Async Inference)

1. POST endpoint creates Job record (status=pending), returns job_id immediately
2. FastAPI BackgroundTasks runs the actual Replicate API call
3. Client polls GET /api/jobs/{id} via useSWR (refreshInterval: 2000ms)
4. On completion: update job status, store result_url, update related frame records

## Package Management

Always `uv add <package>` to add dependencies. Never use `pip install`.

## Testing

- **Unit tests** (`tests/unit/`): services, schemas, config, infrastructure methods
- **Integration tests** (`tests/integration/`): full API endpoints with test DB
- Mock ONLY external services (Replicate, R2); use real DB with fixtures
- SQLite in-memory for test DB (no external dependency needed)
- `@pytest.mark.asyncio` on all async test functions
- Fixtures in `tests/conftest.py`: `db_session`, `client`, `mock_replicate`, `mock_r2`
