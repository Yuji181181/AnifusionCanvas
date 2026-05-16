# API Contract

このドキュメントは、AnifusionCanvas の web / api 間で共有する現在のHTTP契約をまとめる。

## Source Of Truth

| 領域 | ファイル |
| --- | --- |
| Go domain | `apps/api/internal/domain/types.go` |
| Go routes | `apps/api/internal/http/router/router.go` |
| TypeScript contracts | `packages/contracts/src/index.ts` |
| Web API client | `apps/web/src/lib/api-client.ts` |

GoのJSON tagとTypeScript contractのプロパティ名を一致させる。片方だけを変更しない。

## Common Types

| Concept | Go | TypeScript |
| --- | --- | --- |
| Project | `domain.Project` | `Project` |
| Storage object | `storage.StoredObject` | `StorageObject` |
| Frame | `domain.Frame` | `Frame` |
| Frame kind | `domain.FrameKind` | `FrameKind` |
| Job | `domain.Job` | `Job<T>` |
| Job type | `domain.JobType` | `JobType` |
| Job status | `domain.JobStatus` | `JobStatus` |
| Export artifact | `domain.ExportArtifact` | `ExportArtifact` |
| Dependency check result | `dependency.CheckResult` | `DependencyCheckResult` |
| API error | `handler.ErrorBody` | `ApiErrorResponse` |

### Project

| JSON field | Type | Notes |
| --- | --- | --- |
| `id` | string | Project ID |
| `name` | string | User-facing project name |
| `createdAt` | string | RFC3339 timestamp |
| `updatedAt` | string | RFC3339 timestamp |

### Frame

| JSON field | Type | Notes |
| --- | --- | --- |
| `id` | string | Frame ID |
| `projectId` | string | Project ID |
| `index` | number | Timeline order |
| `imageUrl` | string | Current frame image URL or data URL |
| `thumbnailUrl` | string | Thumbnail URL. Falls back to `imageUrl` in web UI |
| `kind` | `key` / `generated` / `inpainted` / `edited` | Frame origin |
| `note` | string, optional | Prompt or edit note |
| `updatedAt` | string | RFC3339 timestamp |

### Storage Object

| JSON field | Type | Notes |
| --- | --- | --- |
| `key` | string | R2 object key |
| `url` | string | Public URL when `R2_PUBLIC_BASE_URL` is set, otherwise an internal `r2://` URI |
| `contentType` | string | Object media type |
| `size` | number | Object size in bytes |

### Export Artifact

`ExportArtifact` extends `StorageObject` with video export metadata.

| JSON field | Type | Notes |
| --- | --- | --- |
| `key` | string | R2 object key |
| `url` | string | Public URL when `R2_PUBLIC_BASE_URL` is set, otherwise an internal `r2://` URI |
| `contentType` | string | Exported video media type |
| `size` | number | Object size in bytes |
| `frameCount` | number | Number of frames encoded into the video |
| `fps` | number | Frames per second used for the export |

### Job

| JSON field | Type | Notes |
| --- | --- | --- |
| `id` | string | Job ID |
| `projectId` | string, optional | Project ID |
| `type` | `generation` / `inpainting` / `export` | Job category |
| `status` | `queued` / `running` / `succeeded` / `failed` | Job state |
| `progress` | number | 0 to 100 |
| `message` | string | User-facing status |
| `result` | object, optional | Typed by job category |
| `error` | string, optional | Failure detail |
| `createdAt` | string | RFC3339 timestamp |
| `updatedAt` | string | RFC3339 timestamp |

Job result types:

| `type` | Result type |
| --- | --- |
| `generation` | `GenerateFramesResult` |
| `inpainting` | `InpaintFrameResult` |
| `export` | `ExportVideoResult` |

Typed job aliases:

```ts
type GenerateFramesJob = Job<GenerateFramesResult> & { type: 'generation' }
type InpaintFrameJob = Job<InpaintFrameResult> & { type: 'inpainting' }
type ExportVideoJob = Job<ExportVideoResult> & { type: 'export' }
type StudioJob = GenerateFramesJob | InpaintFrameJob | ExportVideoJob
```

### Dependency Check Result

| JSON field | Type | Notes |
| --- | --- | --- |
| `name` | `database` / `replicate` / `r2` / `ffmpeg` | Checked dependency |
| `status` | `ok` / `skipped` / `error` | `skipped` means the dependency is not configured in the current environment |
| `message` | string | Human-readable check detail |

## Error Response

All API errors should use this shape.

```json
{
  "error": {
    "code": "Bad Request",
    "message": "projectId is required"
  }
}
```

`code` currently uses `http.StatusText(status)`.

The web client reads `error.message` first and falls back to plain response text for non-JSON failures.

## Endpoints

### `GET /health`

Response:

```ts
type HealthResponse = {
  status: string
}
```

### `GET /health/dependencies`

Checks database, Replicate, R2, and FFmpeg availability.

Response:

```ts
type HealthDependenciesResponse = {
  status: 'ok' | 'degraded'
  results: DependencyCheckResult[]
}
```

Behavior:

- `status` is `ok` when no dependency check has `status: 'error'`.
- `status` is `degraded` when at least one dependency check has `status: 'error'`.
- Missing optional local-development configuration returns a dependency result with `status: 'skipped'`.
- Production readiness expects `database`, `replicate`, `r2`, and `ffmpeg` to all return `status: 'ok'`.

### `POST /projects`

Request:

```ts
type CreateProjectRequest = {
  id: string
  name: string
}
```

Response:

```ts
type ProjectResponse = {
  project: Project
}
```

Validation:

- `id` is required.
- `name` is required.

Behavior:

- Creates the project when it does not exist.
- Updates the project name when the same `id` already exists.

### `GET /projects/:projectId`

Response:

```ts
type ProjectResponse = {
  project: Project
}
```

Validation:

- `projectId` is required.

Errors:

- Returns `404` when the project does not exist.

### `PUT /projects/:projectId`

Request:

```ts
type UpdateProjectRequest = {
  id: string
  name: string
}
```

`id` is overwritten from the `:projectId` path parameter by the API layer.

Response:

```ts
type ProjectResponse = {
  project: Project
}
```

Validation:

- `projectId` is required.
- `name` is required.

Errors:

- Returns `404` when the project does not exist.

### `GET /projects/:projectId/frames`

Response:

```ts
type ListFramesResponse = {
  frames: Frame[]
}
```

Validation:

- `projectId` is required.

### `PUT /projects/:projectId/frames/:frameId`

Request:

```ts
type UpdateFrameRequest = {
  projectId: string
  frameId: string
  imageDataUrl: string
  note?: string
}
```

Response:

```ts
type UpdateFrameResult = {
  frame: Frame
}
```

Validation:

- `projectId` is required.
- `frameId` is required.
- `imageDataUrl` is required.
- `imageDataUrl` must start with `data:`.

### `PUT /projects/:projectId/frames/:frameId/metadata`

Request:

```ts
type UpdateFrameMetadataRequest = {
  projectId: string
  frameId: string
  kind?: FrameKind
  note?: string
}
```

`projectId` and `frameId` are overwritten from the path parameters by the API layer.

Response:

```ts
type UpdateFrameResult = {
  frame: Frame
}
```

Validation:

- `projectId` is required.
- `frameId` is required.
- At least one of `kind` or `note` is required.
- `kind`, when present, must be one of `key`, `generated`, `inpainted`, or `edited`.

Errors:

- Returns `404` when the frame does not exist.

### `DELETE /projects/:projectId/frames/:frameId`

Response:

```ts
type DeleteFrameResponse = void
```

Validation:

- `projectId` is required.
- `frameId` is required.

Behavior:

- Deletes the frame.
- Compacts the remaining frame indexes in ascending timeline order.

Errors:

- Returns `404` when the frame does not exist.

### `PUT /projects/:projectId/frames/reorder`

Request:

```ts
type ReorderFramesRequest = {
  projectId: string
  frameIds: string[]
}
```

`projectId` is overwritten from the `:projectId` path parameter by the API layer.

Response:

```ts
type ReorderFramesResult = {
  frames: Frame[]
}
```

Validation:

- `projectId` is required.
- `frameIds` must not be empty.
- `frameIds` must not contain blank values.
- `frameIds` must not contain duplicate values.
- `frameIds` must include every current frame in the project exactly once.

### `POST /inference/generate`

Request:

```ts
type GenerateFramesRequest = {
  projectId: string
  prompt: string
  negativePrompt?: string
  frameCount: number
  startImageDataUrl: string
  endImageDataUrl: string
}
```

Response:

```ts
type GenerateFramesResponse = {
  job: Job<GenerateFramesResult>
}
```

Validation:

- `projectId` is required.
- `prompt` is required.
- `frameCount` must be between 2 and 12.
- `startImageDataUrl` must start with `data:`.
- `endImageDataUrl` must start with `data:`.

### `POST /inference/inpaint`

Request:

```ts
type InpaintFrameRequest = {
  projectId: string
  frameId: string
  prompt: string
  maskDataUrl: string
  strength: number
}
```

Response:

```ts
type InpaintFrameResponse = {
  job: Job<InpaintFrameResult>
}
```

Validation:

- `projectId` is required.
- `frameId` is required.
- `prompt` is required.
- `maskDataUrl` is required.
- `maskDataUrl` must start with `data:`.
- `strength` must be between 0.1 and 1.

### `POST /export/video`

Request:

```ts
type ExportVideoRequest = {
  projectId: string
  fps: number
}
```

Response:

```ts
type ExportArtifact = StorageObject & {
  frameCount: number
  fps: number
}

type ExportVideoResult = {
  videoUrl: string
  artifact?: ExportArtifact
}

type ExportVideoResponse = {
  job: Job<ExportVideoResult>
}
```

Validation:

- `projectId` is required.
- `fps` must be greater than 0.
- `fps` must be 60 or less.

### `GET /jobs/:jobId`

Response:

```ts
type GetJobResponse<T = unknown> = {
  job: Job<T>
}
```

For general studio job polling, web clients may use:

```ts
type GetStudioJobResponse = {
  job: StudioJob
}
```

Validation:

- `jobId` is required.

## Change Rules

- When a JSON field changes, update Go domain, TypeScript contract, web API client, and tests in the same PR.
- When a route changes, update this document and `apps/api/internal/http/router/router_test.go`.
- When validation changes, update both this document and HTTP tests.
- For new external I/O behavior, keep HTTP handler contracts stable and add infrastructure behind usecase boundaries.
