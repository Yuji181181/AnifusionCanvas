export type JobStatus = 'queued' | 'running' | 'succeeded' | 'failed'

export type FrameKind = 'key' | 'generated' | 'inpainted' | 'edited'

export type Frame = {
  id: string
  projectId: string
  index: number
  imageUrl: string
  thumbnailUrl: string
  kind: FrameKind
  note?: string
  updatedAt: string
}

export type Project = {
  id: string
  name: string
  createdAt: string
  updatedAt: string
}

export type Job<T = unknown> = {
  id: string
  projectId?: string
  type: 'generation' | 'inpainting' | 'export'
  status: JobStatus
  progress: number
  message: string
  result?: T
  error?: string
  createdAt: string
  updatedAt: string
}

export type CreateProjectRequest = {
  id: string
  name: string
}

export type ProjectResponse = {
  project: Project
}

export type UpdateProjectRequest = {
  id: string
  name: string
}

export type GenerateFramesRequest = {
  projectId: string
  prompt: string
  negativePrompt?: string
  frameCount: number
  startImageDataUrl: string
  endImageDataUrl: string
}

export type GenerateFramesResult = {
  frames: Frame[]
}

export type GenerateFramesResponse = {
  job: Job<GenerateFramesResult>
}

export type InpaintFrameRequest = {
  projectId: string
  frameId: string
  prompt: string
  maskDataUrl: string
  strength: number
}

export type InpaintFrameResult = {
  frame: Frame
}

export type InpaintFrameResponse = {
  job: Job<InpaintFrameResult>
}

export type UpdateFrameRequest = {
  projectId: string
  frameId: string
  imageDataUrl: string
  note?: string
}

export type UpdateFrameResult = {
  frame: Frame
}

export type UpdateFrameMetadataRequest = {
  projectId: string
  frameId: string
  kind?: FrameKind
  note?: string
}

export type ReorderFramesRequest = {
  projectId: string
  frameIds: string[]
}

export type ReorderFramesResult = {
  frames: Frame[]
}

export type ExportVideoRequest = {
  projectId: string
  fps: number
}

export type ExportVideoResult = {
  videoUrl: string
}

export type ExportVideoResponse = {
  job: Job<ExportVideoResult>
}

export type ListFramesResponse = {
  frames: Frame[]
}

export type GetJobResponse<T = unknown> = {
  job: Job<T>
}

export type HealthResponse = {
  status: string
}

export type DependencyCheckStatus = 'ok' | 'skipped' | 'error'

export type DependencyCheckResult = {
  name: 'database' | 'replicate' | 'r2' | 'ffmpeg'
  status: DependencyCheckStatus
  message: string
}

export type HealthDependenciesResponse = {
  status: 'ok' | 'degraded'
  results: DependencyCheckResult[]
}

export type ApiErrorResponse = {
  error: {
    code: string
    message: string
  }
}
