import type { Frame, Job } from '@anifusion/contracts'
import { beforeEach, describe, expect, it } from 'vitest'
import { useFrameStore } from '@/stores/frame-store'

function frame(overrides: Partial<Frame>): Frame {
  return {
    id: 'frame-1',
    projectId: 'demo-project',
    index: 0,
    imageUrl: 'data:image/png;base64,frame',
    thumbnailUrl: 'data:image/png;base64,frame',
    kind: 'generated',
    updatedAt: '2026-05-16T00:00:00Z',
    ...overrides,
  }
}

function job(overrides: Partial<Job>): Job {
  return {
    id: 'job-1',
    projectId: 'demo-project',
    type: 'generation',
    status: 'queued',
    progress: 0,
    message: 'queued',
    createdAt: '2026-05-16T00:00:00Z',
    updatedAt: '2026-05-16T00:00:00Z',
    ...overrides,
  }
}

describe('useFrameStore', () => {
  beforeEach(() => {
    useFrameStore.setState({
      projectId: 'demo-project',
      frames: [],
      selectedFrameId: undefined,
      activeJob: undefined,
    })
  })

  it('selects the first frame when frames are set without an existing selection', () => {
    useFrameStore.getState().setFrames([
      frame({ id: 'frame-a', index: 0 }),
      frame({ id: 'frame-b', index: 1 }),
    ])

    expect(useFrameStore.getState().selectedFrameId).toBe('frame-a')
  })

  it('keeps the current selection when replacing the frame list', () => {
    useFrameStore.getState().setSelectedFrameId('frame-b')

    useFrameStore.getState().setFrames([
      frame({ id: 'frame-a', index: 0 }),
      frame({ id: 'frame-b', index: 1 }),
    ])

    expect(useFrameStore.getState().selectedFrameId).toBe('frame-b')
  })

  it('inserts new frames in timeline order and selects the inserted frame', () => {
    useFrameStore.getState().setFrames([
      frame({ id: 'frame-a', index: 0 }),
      frame({ id: 'frame-c', index: 2 }),
    ])

    useFrameStore.getState().upsertFrame(frame({ id: 'frame-b', index: 1 }))

    expect(useFrameStore.getState().frames.map((item) => item.id)).toEqual(['frame-a', 'frame-b', 'frame-c'])
    expect(useFrameStore.getState().selectedFrameId).toBe('frame-b')
  })

  it('replaces an existing frame without changing its timeline position', () => {
    useFrameStore.getState().setFrames([
      frame({ id: 'frame-a', index: 0, kind: 'key' }),
      frame({ id: 'frame-b', index: 1, kind: 'generated' }),
    ])

    useFrameStore.getState().upsertFrame(frame({ id: 'frame-b', index: 1, kind: 'edited', note: 'cleanup' }))

    expect(useFrameStore.getState().frames).toHaveLength(2)
    expect(useFrameStore.getState().frames[1]).toMatchObject({
      id: 'frame-b',
      kind: 'edited',
      note: 'cleanup',
    })
    expect(useFrameStore.getState().selectedFrameId).toBe('frame-b')
  })

  it('tracks and clears the active job', () => {
    useFrameStore.getState().setActiveJob(job({ id: 'job-generation', progress: 35, status: 'running' }))

    expect(useFrameStore.getState().activeJob).toMatchObject({
      id: 'job-generation',
      progress: 35,
      status: 'running',
    })

    useFrameStore.getState().setActiveJob(undefined)

    expect(useFrameStore.getState().activeJob).toBeUndefined()
  })
})
