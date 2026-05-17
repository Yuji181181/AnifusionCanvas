import type { Frame, Job } from '@anifusion/contracts'
import { create } from 'zustand'

const projectIdStorageKey = 'anifusion.sessionProjectId'

type FrameState = {
  projectId: string
  frames: Frame[]
  selectedFrameId?: string
  activeJob?: Job
  setFrames: (frames: Frame[]) => void
  upsertFrame: (frame: Frame) => void
  setSelectedFrameId: (frameId: string) => void
  setActiveJob: (job?: Job) => void
}

function createProjectId() {
  const suffix = typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random().toString(36).slice(2)}`

  return `browser-${suffix}`
}

function getBrowserProjectId() {
  if (typeof window === 'undefined') {
    return createProjectId()
  }

  const existing = window.sessionStorage.getItem(projectIdStorageKey)
  if (existing) {
    return existing
  }

  const projectId = createProjectId()
  window.sessionStorage.setItem(projectIdStorageKey, projectId)
  return projectId
}

export const useFrameStore = create<FrameState>((set) => ({
  projectId: getBrowserProjectId(),
  frames: [],
  setFrames: (frames) =>
    set((state) => ({
      frames,
      selectedFrameId: state.selectedFrameId ?? frames[0]?.id,
    })),
  upsertFrame: (frame) =>
    set((state) => ({
      frames: state.frames.some((item) => item.id === frame.id)
        ? state.frames.map((item) => (item.id === frame.id ? frame : item))
        : [...state.frames, frame].sort((a, b) => a.index - b.index),
      selectedFrameId: frame.id,
    })),
  setSelectedFrameId: (selectedFrameId) => set({ selectedFrameId }),
  setActiveJob: (activeJob) => set({ activeJob }),
}))

export function useSelectedFrame() {
  return useFrameStore((state) => state.frames.find((frame) => frame.id === state.selectedFrameId))
}
