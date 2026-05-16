import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from '@/lib/api-client'

function mockFetch(response: unknown, init?: ResponseInit) {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(JSON.stringify(response), {
      headers: { 'content-type': 'application/json' },
      status: 200,
      ...init,
    }),
  )
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

describe('apiClient', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('fetches dependency health checks', async () => {
    const fetchMock = mockFetch({
      status: 'degraded',
      results: [
        { name: 'database', status: 'skipped', message: 'DATABASE_URL is not set' },
        { name: 'ffmpeg', status: 'ok', message: 'ffmpeg found at /usr/bin/ffmpeg' },
      ],
    })

    await apiClient.healthDependencies()

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/health/dependencies',
      expect.objectContaining({
        headers: expect.objectContaining({ 'content-type': 'application/json' }),
      }),
    )
  })

  it('sends project creation requests as JSON', async () => {
    const fetchMock = mockFetch({
      project: {
        id: 'project-1',
        name: 'Demo project',
        createdAt: '2026-05-17T00:00:00Z',
        updatedAt: '2026-05-17T00:00:00Z',
      },
    })

    await apiClient.createProject({ id: 'project-1', name: 'Demo project' })

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/projects',
      expect.objectContaining({
        body: JSON.stringify({ id: 'project-1', name: 'Demo project' }),
        headers: expect.objectContaining({ 'content-type': 'application/json' }),
        method: 'POST',
      }),
    )
  })

  it('encodes project path parameters when updating a project', async () => {
    const fetchMock = mockFetch({
      project: {
        id: 'project with space',
        name: 'Renamed project',
        createdAt: '2026-05-17T00:00:00Z',
        updatedAt: '2026-05-17T00:00:01Z',
      },
    })

    await apiClient.updateProject({ id: 'project with space', name: 'Renamed project' })

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/projects/project%20with%20space',
      expect.objectContaining({
        body: JSON.stringify({ id: 'project with space', name: 'Renamed project' }),
        headers: expect.objectContaining({ 'content-type': 'application/json' }),
        method: 'PUT',
      }),
    )
  })

  it('encodes project and frame path parameters when updating a frame', async () => {
    const fetchMock = mockFetch({
      frame: {
        id: 'frame/1',
        projectId: 'project with space',
        index: 1,
        imageUrl: 'data:image/png;base64,edited',
        thumbnailUrl: 'data:image/png;base64,edited',
        kind: 'edited',
        updatedAt: '2026-05-16T00:00:00Z',
      },
    })

    await apiClient.updateFrame({
      projectId: 'project with space',
      frameId: 'frame/1',
      imageDataUrl: 'data:image/png;base64,edited',
      note: 'manual edit',
    })

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/projects/project%20with%20space/frames/frame%2F1',
      expect.objectContaining({
        body: JSON.stringify({
          projectId: 'project with space',
          frameId: 'frame/1',
          imageDataUrl: 'data:image/png;base64,edited',
          note: 'manual edit',
        }),
        headers: expect.objectContaining({ 'content-type': 'application/json' }),
        method: 'PUT',
      }),
    )
  })

  it('sends frame metadata update requests as JSON', async () => {
    const fetchMock = mockFetch({
      frame: {
        id: 'frame/1',
        projectId: 'project with space',
        index: 1,
        imageUrl: 'data:image/png;base64,edited',
        thumbnailUrl: 'data:image/png;base64,edited',
        kind: 'edited',
        note: 'cleanup',
        updatedAt: '2026-05-17T00:00:00Z',
      },
    })

    await apiClient.updateFrameMetadata({
      projectId: 'project with space',
      frameId: 'frame/1',
      kind: 'edited',
      note: 'cleanup',
    })

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/projects/project%20with%20space/frames/frame%2F1/metadata',
      expect.objectContaining({
        body: JSON.stringify({
          projectId: 'project with space',
          frameId: 'frame/1',
          kind: 'edited',
          note: 'cleanup',
        }),
        headers: expect.objectContaining({ 'content-type': 'application/json' }),
        method: 'PUT',
      }),
    )
  })

  it('sends frame reorder requests as JSON', async () => {
    const fetchMock = mockFetch({ frames: [] })

    await apiClient.reorderFrames({
      projectId: 'project with space',
      frameIds: ['frame/2', 'frame/1'],
    })

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/projects/project%20with%20space/frames/reorder',
      expect.objectContaining({
        body: JSON.stringify({
          projectId: 'project with space',
          frameIds: ['frame/2', 'frame/1'],
        }),
        headers: expect.objectContaining({ 'content-type': 'application/json' }),
        method: 'PUT',
      }),
    )
  })

  it('handles no-content frame deletion responses', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(apiClient.deleteFrame('project with space', 'frame/1')).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/projects/project%20with%20space/frames/frame%2F1',
      expect.objectContaining({
        headers: expect.objectContaining({ 'content-type': 'application/json' }),
        method: 'DELETE',
      }),
    )
  })

  it('encodes job IDs when fetching a job', async () => {
    const fetchMock = mockFetch({
      job: {
        id: 'job/1',
        type: 'generation',
        status: 'succeeded',
        progress: 100,
        message: 'done',
        version: 1,
        createdAt: '2026-05-16T00:00:00Z',
        updatedAt: '2026-05-16T00:00:00Z',
      },
    })

    await apiClient.getJob('job/1')

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/jobs/job%2F1',
      expect.objectContaining({
        headers: expect.objectContaining({ 'content-type': 'application/json' }),
      }),
    )
  })

  it('sends generation requests as JSON', async () => {
    const fetchMock = mockFetch({
      job: {
        id: 'job-1',
        type: 'generation',
        status: 'queued',
        progress: 0,
        message: 'queued',
        version: 1,
        createdAt: '2026-05-16T00:00:00Z',
        updatedAt: '2026-05-16T00:00:00Z',
      },
    })

    await apiClient.generateFrames({
      projectId: 'project-1',
      prompt: 'turn around',
      frameCount: 4,
      startImageDataUrl: 'data:image/png;base64,start',
      endImageDataUrl: 'data:image/png;base64,end',
    })

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/inference/generate',
      expect.objectContaining({
        body: JSON.stringify({
          projectId: 'project-1',
          prompt: 'turn around',
          frameCount: 4,
          startImageDataUrl: 'data:image/png;base64,start',
          endImageDataUrl: 'data:image/png;base64,end',
        }),
        headers: expect.objectContaining({ 'content-type': 'application/json' }),
        method: 'POST',
      }),
    )
  })

  it('throws API error response messages for failed requests', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: { code: 'Bad Request', message: 'projectId is required' } }), {
          headers: { 'content-type': 'application/json' },
          status: 400,
        }),
      ),
    )

    await expect(apiClient.listFrames('')).rejects.toThrow('projectId is required')
  })

  it('rejects API responses that do not match the schema', async () => {
    mockFetch({
      frames: [
        {
          id: 'frame-1',
          projectId: 'project-1',
          index: 0,
          imageUrl: 'data:image/png;base64,frame',
          thumbnailUrl: 'data:image/png;base64,frame',
          kind: 'unexpected',
          updatedAt: '2026-05-17T00:00:00Z',
        },
      ],
    })

    await expect(apiClient.listFrames('project-1')).rejects.toThrow()
  })

  it('falls back to response text for non-JSON failed requests', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('service unavailable', { status: 503 })))

    await expect(apiClient.listFrames('project-1')).rejects.toThrow('service unavailable')
  })
})
