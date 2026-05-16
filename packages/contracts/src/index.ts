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

export type Job<T = unknown> = {
  id: string
  type: 'generation' | 'inpainting' | 'export'
  status: JobStatus
  progress: number
  message: string
  result?: T
  error?: string
  createdAt: string
  updatedAt: string
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

export type UpdateFrameRequest = {
  projectId: string
  frameId: string
  imageDataUrl: string
  note?: string
}

export type ExportVideoRequest = {
  projectId: string
  fps: number
}

export type ExportVideoResult = {
  videoUrl: string
}

export type HealthResponse = {
  status: string
}
