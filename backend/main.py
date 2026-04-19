from __future__ import annotations

import asyncio
from datetime import UTC, datetime
from enum import StrEnum
from typing import Any
from uuid import uuid4

from fastapi import BackgroundTasks, FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field


class JobType(StrEnum):
    INBETWEEN = "inbetween"
    INPAINT = "inpaint"


class JobStatus(StrEnum):
    PENDING = "pending"
    RUNNING = "running"
    SUCCEEDED = "succeeded"
    FAILED = "failed"


class CreateJobRequest(BaseModel):
    job_type: JobType
    payload: dict[str, Any] = Field(default_factory=dict)


class CreateJobResponse(BaseModel):
    job_id: str
    status: JobStatus


class JobResponse(BaseModel):
    job_id: str
    job_type: JobType
    status: JobStatus
    result: dict[str, Any] | None = None
    created_at: datetime
    updated_at: datetime


app = FastAPI(title="Anifusion Canvas API", version="0.1.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://127.0.0.1:3000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

jobs: dict[str, JobResponse] = {}


@app.get("/health")
async def health() -> dict[str, str]:
    return {
        "service": "anifusion-canvas-backend",
        "status": "ok",
        "timestamp": datetime.now(UTC).isoformat(),
    }


async def _simulate_job(job_id: str) -> None:
    job = jobs[job_id]
    jobs[job_id] = job.model_copy(
        update={
            "status": JobStatus.RUNNING,
            "updated_at": datetime.now(UTC),
        }
    )

    await asyncio.sleep(2)

    completed = jobs[job_id].model_copy(
        update={
            "status": JobStatus.SUCCEEDED,
            "result": {
                "frames": [
                    "https://example.invalid/frame-001.png",
                    "https://example.invalid/frame-002.png",
                ],
                "preview_video_url": "https://example.invalid/preview.mp4",
            },
            "updated_at": datetime.now(UTC),
        }
    )
    jobs[job_id] = completed


@app.post("/v1/jobs", response_model=CreateJobResponse, status_code=202)
async def create_job(
    payload: CreateJobRequest, background_tasks: BackgroundTasks
) -> CreateJobResponse:
    now = datetime.now(UTC)
    job_id = str(uuid4())

    jobs[job_id] = JobResponse(
        job_id=job_id,
        job_type=payload.job_type,
        status=JobStatus.PENDING,
        created_at=now,
        updated_at=now,
    )

    background_tasks.add_task(_simulate_job, job_id)
    return CreateJobResponse(job_id=job_id, status=JobStatus.PENDING)


@app.get("/v1/jobs/{job_id}", response_model=JobResponse)
async def get_job(job_id: str) -> JobResponse:
    job = jobs.get(job_id)
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")
    return job
