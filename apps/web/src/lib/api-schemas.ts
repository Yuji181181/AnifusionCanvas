import type {
  ExportVideoResponse,
  GenerateFramesResponse,
  GetJobResponse,
  HealthDependenciesResponse,
  HealthResponse,
  InpaintFrameResponse,
  ListFramesResponse,
  ProjectResponse,
  ReorderFramesResult,
  UpdateFrameResult,
} from '@anifusion/contracts'
import { z } from 'zod'

export const apiErrorResponseSchema = z.object({
  error: z.object({
    code: z.string(),
    message: z.string(),
  }),
})

const frameKindSchema = z.enum(['key', 'generated', 'inpainted', 'edited'])

export const frameSchema = z.object({
  id: z.string(),
  projectId: z.string(),
  index: z.number(),
  imageUrl: z.string(),
  thumbnailUrl: z.string(),
  kind: frameKindSchema,
  note: z.string().optional(),
  updatedAt: z.string(),
})

export const projectSchema = z.object({
  id: z.string(),
  name: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const storageObjectSchema = z.object({
  key: z.string(),
  url: z.string(),
  contentType: z.string(),
  size: z.number(),
})

export const exportArtifactSchema = storageObjectSchema.extend({
  frameCount: z.number(),
  fps: z.number(),
})

const jobStatusSchema = z.enum(['queued', 'running', 'succeeded', 'failed'])
const jobTypeSchema = z.enum(['generation', 'inpainting', 'export'])

export function jobSchema<T extends z.ZodTypeAny>(resultSchema: T = z.unknown() as T) {
  return z.object({
    id: z.string(),
    projectId: z.string().optional(),
    type: jobTypeSchema,
    status: jobStatusSchema,
    progress: z.number(),
    message: z.string(),
    result: resultSchema.optional(),
    error: z.string().optional(),
    createdAt: z.string(),
    updatedAt: z.string(),
  })
}

export const generateFramesResultSchema = z.object({
  frames: z.array(frameSchema),
})

export const inpaintFrameResultSchema = z.object({
  frame: frameSchema,
})

export const exportVideoResultSchema = z.object({
  videoUrl: z.string(),
  artifact: exportArtifactSchema.optional(),
})

export const healthResponseSchema = z.object({
  status: z.string(),
}) satisfies z.ZodType<HealthResponse>

const dependencyCheckResultSchema = z.object({
  name: z.enum(['database', 'replicate', 'r2', 'ffmpeg']),
  status: z.enum(['ok', 'skipped', 'error']),
  message: z.string(),
})

export const healthDependenciesResponseSchema = z.object({
  status: z.enum(['ok', 'degraded']),
  results: z.array(dependencyCheckResultSchema),
}) satisfies z.ZodType<HealthDependenciesResponse>

export const projectResponseSchema = z.object({
  project: projectSchema,
}) satisfies z.ZodType<ProjectResponse>

export const listFramesResponseSchema = z.object({
  frames: z.array(frameSchema),
}) satisfies z.ZodType<ListFramesResponse>

export const generateFramesResponseSchema = z.object({
  job: jobSchema(generateFramesResultSchema),
}) satisfies z.ZodType<GenerateFramesResponse>

export const inpaintFrameResponseSchema = z.object({
  job: jobSchema(inpaintFrameResultSchema),
}) satisfies z.ZodType<InpaintFrameResponse>

export const updateFrameResultSchema = z.object({
  frame: frameSchema,
}) satisfies z.ZodType<UpdateFrameResult>

export const reorderFramesResultSchema = z.object({
  frames: z.array(frameSchema),
}) satisfies z.ZodType<ReorderFramesResult>

export const exportVideoResponseSchema = z.object({
  job: jobSchema(exportVideoResultSchema),
}) satisfies z.ZodType<ExportVideoResponse>

export const getJobResponseSchema = z.object({
  job: jobSchema(),
}) satisfies z.ZodType<GetJobResponse>
