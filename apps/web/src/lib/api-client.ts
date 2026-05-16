import type {
  ApiErrorResponse,
  CreateProjectRequest,
  ExportVideoRequest,
  ExportVideoResponse,
  GenerateFramesRequest,
  GenerateFramesResponse,
  GetJobResponse,
  HealthDependenciesResponse,
  HealthResponse,
  InpaintFrameRequest,
  InpaintFrameResponse,
  ListFramesResponse,
  ProjectResponse,
  UpdateFrameRequest,
  UpdateFrameResult,
  UpdateProjectRequest,
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
    const message = await errorMessage(response)
    throw new Error(message || `API request failed: ${response.status}`)
  }

  return response.json() as Promise<T>
}

async function errorMessage(response: Response): Promise<string> {
  const body = await response.text()
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    try {
      const payload = JSON.parse(body) as Partial<ApiErrorResponse>
      if (payload.error?.message) {
        return payload.error.message
      }
    } catch {
      return body
    }
  }

  return body
}

function pathParam(value: string): string {
  return encodeURIComponent(value)
}

export const apiClient = {
  health() {
    return request<HealthResponse>('/health')
  },
  healthDependencies() {
    return request<HealthDependenciesResponse>('/health/dependencies')
  },
  createProject(input: CreateProjectRequest) {
    return request<ProjectResponse>('/projects', {
      method: 'POST',
      body: JSON.stringify(input),
    })
  },
  getProject(projectId: string) {
    return request<ProjectResponse>(`/projects/${pathParam(projectId)}`)
  },
  updateProject(input: UpdateProjectRequest) {
    return request<ProjectResponse>(`/projects/${pathParam(input.id)}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    })
  },
  listFrames(projectId: string) {
    return request<ListFramesResponse>(`/projects/${pathParam(projectId)}/frames`)
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
    return request<UpdateFrameResult>(`/projects/${pathParam(input.projectId)}/frames/${pathParam(input.frameId)}`, {
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
    return request<GetJobResponse<T>>(`/jobs/${pathParam(jobId)}`)
  },
}
