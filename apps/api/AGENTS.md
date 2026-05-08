# Agent Rules: Python / FastAPI

## Pydantic v2

Use `model_config = ConfigDict(...)` not `class Config`.
Schema fields use `Field()` for validation and documentation.
Always define both request and response schemas for API endpoints.

## SQLAlchemy 2.0

Use `Mapped[type]` and `mapped_column()` — never `Column()`.
Use `DeclarativeBase` and `async_sessionmaker`.
Relationships use `relationship()` with `back_populates`.
Models must have `__tablename__` explicitly set.

## FastAPI

Use `Depends()` for dependency injection (DB sessions, services, settings).
Return Pydantic schemas from endpoints, never ORM models directly.
Use `BackgroundTasks` for async job execution.
Handle errors with `HTTPException` and appropriate status codes.

## Type Hints

All function signatures must have type hints.
Use `async def` for all I/O-bound functions.
Use `AsyncGenerator` for async dependency generators.
Import types from `typing` or `collections.abc` as appropriate.
