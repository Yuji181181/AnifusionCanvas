import { create } from 'zustand'

type Tool = 'select' | 'pen' | 'rect' | 'circle' | 'polygon' | 'text'

type EditorState = {
  tool: Tool
  color: string
  brushSize: number
  setTool: (tool: Tool) => void
  setColor: (color: string) => void
  setBrushSize: (brushSize: number) => void
}

export const useEditorStore = create<EditorState>((set) => ({
  tool: 'pen',
  color: '#111827',
  brushSize: 8,
  setTool: (tool) => set({ tool }),
  setColor: (color) => set({ color }),
  setBrushSize: (brushSize) => set({ brushSize }),
}))
