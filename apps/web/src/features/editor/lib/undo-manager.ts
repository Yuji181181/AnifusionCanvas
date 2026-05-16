import type { Canvas } from 'fabric'

export interface HistoryState {
  json: string
}

export class UndoManager {
  private undoStack: HistoryState[] = []
  private redoStack: HistoryState[] = []
  private maxSize: number
  private canvas: Canvas

  constructor(canvas: Canvas, maxSize = 50) {
    this.canvas = canvas
    this.maxSize = maxSize
  }

  save() {
    this.undoStack.push({ json: JSON.stringify(this.canvas.toJSON()) })
    if (this.undoStack.length > this.maxSize) {
      this.undoStack.shift()
    }
    this.redoStack = []
  }

  undo() {
    if (this.undoStack.length === 0) {
      return
    }

    this.redoStack.push({ json: JSON.stringify(this.canvas.toJSON()) })
    const state = this.undoStack.pop()!
    this.loadState(state)
  }

  redo() {
    if (this.redoStack.length === 0) {
      return
    }

    this.undoStack.push({ json: JSON.stringify(this.canvas.toJSON()) })
    const state = this.redoStack.pop()!
    this.loadState(state)
  }

  canUndo(): boolean {
    return this.undoStack.length > 0
  }

  canRedo(): boolean {
    return this.redoStack.length > 0
  }

  reset() {
    this.undoStack = []
    this.redoStack = []
  }

  private loadState(state: HistoryState) {
    this.canvas.loadFromJSON(JSON.parse(state.json)).then(() => {
      this.canvas.renderAll()
    })
  }
}
