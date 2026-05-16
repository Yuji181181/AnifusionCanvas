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

  it('encodes job IDs when fetching a job', async () => {
    const fetchMock = mockFetch({
      job: {
        id: 'job/1',
        type: 'generation',
        status: 'succeeded',
        progress: 100,
        message: 'done',
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

  it('falls back to response text for non-JSON failed requests', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('service unavailable', { status: 503 })))

    await expect(apiClient.listFrames('project-1')).rejects.toThrow('service unavailable')
  })
})
