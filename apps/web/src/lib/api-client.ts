import type {
  ExportVideoRequest,
  ExportVideoResponse,
  GenerateFramesRequest,
  GenerateFramesResponse,
  GetJobResponse,
  InpaintFrameRequest,
  InpaintFrameResponse,
  ListFramesResponse,
  UpdateFrameRequest,
  UpdateFrameResult,
} from '@anifusion/contracts'
import { env } from './env'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${env.apiBaseUrl}${path}`, {
    ...options,
    headers: {
      'content-type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || `API request failed: ${response.status}`)
  }

  return response.json() as Promise<T>
}

export const apiClient = {
  listFrames(projectId: string) {
    return request<ListFramesResponse>(`/projects/${projectId}/frames`)
  },
  generateFrames(input: GenerateFramesRequest) {
    return request<GenerateFramesResponse>('/inference/generate', {
      method: 'POST',
      body: JSON.stringify(input),
    })
  },
  inpaintFrame(input: InpaintFrameRequest) {
    return request<InpaintFrameResponse>('/inference/inpaint', {
      method: 'POST',
      body: JSON.stringify(input),
    })
  },
  updateFrame(input: UpdateFrameRequest) {
    return request<UpdateFrameResult>(`/projects/${input.projectId}/frames/${input.frameId}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    })
  },
  exportVideo(input: ExportVideoRequest) {
    return request<ExportVideoResponse>('/export/video', {
      method: 'POST',
      body: JSON.stringify(input),
    })
  },
  getJob<T>(jobId: string) {
    return request<GetJobResponse<T>>(`/jobs/${jobId}`)
  },
}
