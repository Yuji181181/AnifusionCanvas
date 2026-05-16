import type {
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
  ReorderFramesRequest,
  UpdateFrameRequest,
  UpdateFrameMetadataRequest,
  UpdateProjectRequest,
} from '@anifusion/contracts'
import { z } from 'zod'
import {
  apiErrorResponseSchema,
  exportVideoResponseSchema,
  generateFramesResponseSchema,
  getJobResponseSchema,
  healthDependenciesResponseSchema,
  healthResponseSchema,
  inpaintFrameResponseSchema,
  listFramesResponseSchema,
  projectResponseSchema,
  reorderFramesResultSchema,
  updateFrameResultSchema,
} from './api-schemas'
import { env } from './env'

async function request<T>(path: string, options: RequestInit | undefined, schema: z.ZodType<T>): Promise<T> {
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

  if (response.status === 204) {
    return undefined as T
  }

  return schema.parse(await response.json())
}

async function errorMessage(response: Response): Promise<string> {
  const body = await response.text()
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    try {
      const payload = apiErrorResponseSchema.safeParse(JSON.parse(body))
      if (!payload.success) {
        return body
      }
      return payload.data.error.message
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
    return request<HealthResponse>('/health', undefined, healthResponseSchema)
  },
  healthDependencies() {
    return request<HealthDependenciesResponse>('/health/dependencies', undefined, healthDependenciesResponseSchema)
  },
  createProject(input: CreateProjectRequest) {
    return request<ProjectResponse>(
      '/projects',
      {
        method: 'POST',
        body: JSON.stringify(input),
      },
      projectResponseSchema,
    )
  },
  getProject(projectId: string) {
    return request<ProjectResponse>(`/projects/${pathParam(projectId)}`, undefined, projectResponseSchema)
  },
  updateProject(input: UpdateProjectRequest) {
    return request<ProjectResponse>(
      `/projects/${pathParam(input.id)}`,
      {
        method: 'PUT',
        body: JSON.stringify(input),
      },
      projectResponseSchema,
    )
  },
  listFrames(projectId: string) {
    return request<ListFramesResponse>(`/projects/${pathParam(projectId)}/frames`, undefined, listFramesResponseSchema)
  },
  generateFrames(input: GenerateFramesRequest) {
    return request<GenerateFramesResponse>(
      '/inference/generate',
      {
        method: 'POST',
        body: JSON.stringify(input),
      },
      generateFramesResponseSchema,
    )
  },
  inpaintFrame(input: InpaintFrameRequest) {
    return request<InpaintFrameResponse>(
      '/inference/inpaint',
      {
        method: 'POST',
        body: JSON.stringify(input),
      },
      inpaintFrameResponseSchema,
    )
  },
  updateFrame(input: UpdateFrameRequest) {
    return request(
      `/projects/${pathParam(input.projectId)}/frames/${pathParam(input.frameId)}`,
      {
        method: 'PUT',
        body: JSON.stringify(input),
      },
      updateFrameResultSchema,
    )
  },
  updateFrameMetadata(input: UpdateFrameMetadataRequest) {
    return request(
      `/projects/${pathParam(input.projectId)}/frames/${pathParam(input.frameId)}/metadata`,
      {
        method: 'PUT',
        body: JSON.stringify(input),
      },
      updateFrameResultSchema,
    )
  },
  deleteFrame(projectId: string, frameId: string) {
    return request<void>(
      `/projects/${pathParam(projectId)}/frames/${pathParam(frameId)}`,
      {
        method: 'DELETE',
      },
      z.undefined(),
    )
  },
  reorderFrames(input: ReorderFramesRequest) {
    return request(
      `/projects/${pathParam(input.projectId)}/frames/reorder`,
      {
        method: 'PUT',
        body: JSON.stringify(input),
      },
      reorderFramesResultSchema,
    )
  },
  exportVideo(input: ExportVideoRequest) {
    return request<ExportVideoResponse>(
      '/export/video',
      {
        method: 'POST',
        body: JSON.stringify(input),
      },
      exportVideoResponseSchema,
    )
  },
  getJob<T>(jobId: string) {
    return request<GetJobResponse<T>>(`/jobs/${pathParam(jobId)}`, undefined, getJobResponseSchema as z.ZodType<GetJobResponse<T>>)
  },
}
