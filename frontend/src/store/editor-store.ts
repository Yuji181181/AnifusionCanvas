import { create } from "zustand";

type EditorState = {
  brushSize: number;
  prompt: string;
  setBrushSize: (brushSize: number) => void;
  setPrompt: (prompt: string) => void;
};

export const useEditorStore = create<EditorState>((set) => ({
  brushSize: 24,
  prompt: "握りこぶし",
  setBrushSize: (brushSize) => set({ brushSize }),
  setPrompt: (prompt) => set({ prompt }),
}));
