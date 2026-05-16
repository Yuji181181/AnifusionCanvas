import type { Frame, Job } from '@anifusion/contracts'
import { create } from 'zustand'

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

export const useFrameStore = create<FrameState>((set) => ({
  projectId: 'demo-project',
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
