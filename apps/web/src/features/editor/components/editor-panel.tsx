import { useMutation } from '@tanstack/react-query'
import { Circle, MousePointer2, PenLine, Redo, Save, SlidersHorizontal, Square, Sun, Trash2, Type, Undo } from 'lucide-react'
import { Canvas, Circle as FabricCircle, PencilBrush, Rect, Textbox, filters } from 'fabric'
import { useCallback, useEffect, useRef } from 'react'
import { apiClient } from '@/lib/api-client'
import { queryClient } from '@/lib/query-client'
import { useEditorStore } from '@/stores/editor-store'
import { useFrameStore, useSelectedFrame } from '@/stores/frame-store'
import { UndoManager } from '../lib/undo-manager'

const tools = [
  { id: 'select', icon: MousePointer2, label: '選択' },
  { id: 'pen', icon: PenLine, label: 'ペン' },
  { id: 'rect', icon: Square, label: '四角' },
  { id: 'circle', icon: Circle, label: '円' },
  { id: 'text', icon: Type, label: 'テキスト' },
] as const

export function EditorPanel() {
  const projectId = useFrameStore((state) => state.projectId)
  const upsertFrame = useFrameStore((state) => state.upsertFrame)
  const frame = useSelectedFrame()
  const tool = useEditorStore((state) => state.tool)
  const color = useEditorStore((state) => state.color)
  const brushSize = useEditorStore((state) => state.brushSize)
  const setTool = useEditorStore((state) => state.setTool)
  const setColor = useEditorStore((state) => state.setColor)
  const setBrushSize = useEditorStore((state) => state.setBrushSize)
  const canvasElementRef = useRef<HTMLCanvasElement | null>(null)
  const fabricRef = useRef<Canvas | null>(null)
  const undoManagerRef = useRef<UndoManager | null>(null)
  const frameImageRef = useRef<fabric.Image | null>(null)

  useEffect(() => {
    if (!canvasElementRef.current) {
      return
    }

    const canvas = new Canvas(canvasElementRef.current, {
      backgroundColor: '#f8fafc',
      height: 540,
      width: 960,
    })
    fabricRef.current = canvas
    const undoManager = new UndoManager(canvas)
    undoManagerRef.current = undoManager

    if (frame?.imageUrl) {
      fabric.Image.fromURL(
        frame.imageUrl,
        (img) => {
          img.set({ selectable: true, evented: true })
          img.scaleToWidth(960)
          canvas.add(img)
          canvas.sendObjectToBack(img)
          frameImageRef.current = img
          canvas.renderAll()
          undoManager.save()
        },
        { crossOrigin: 'anonymous' },
      )
    } else {
      undoManager.save()
    }

    canvas.on('object:modified', () => undoManager.save())
    canvas.on('path:created', () => undoManager.save())

    return () => {
      canvas.dispose()
      fabricRef.current = null
      undoManagerRef.current = null
      frameImageRef.current = null
    }
  }, [frame?.id, frame?.imageUrl])

  useEffect(() => {
    const canvas = fabricRef.current
    if (!canvas) {
      return
    }

    canvas.isDrawingMode = tool === 'pen'
    if (tool === 'pen') {
      const brush = new PencilBrush(canvas)
      brush.color = color
      brush.width = brushSize
      canvas.freeDrawingBrush = brush
    }
  }, [brushSize, color, tool])

  const mutation = useMutation({
    mutationFn: apiClient.updateFrame,
    onSuccess: (data) => {
      upsertFrame(data.frame)
      queryClient.invalidateQueries({ queryKey: ['frames', projectId] })
    },
  })

  const addObject = useCallback(() => {
    const canvas = fabricRef.current
    if (!canvas) {
      return
    }

    if (tool === 'rect') {
      canvas.add(new Rect({ fill: color, height: 88, left: 80, top: 80, width: 132 }))
    }
    if (tool === 'circle') {
      canvas.add(new FabricCircle({ fill: color, left: 120, radius: 52, top: 90 }))
    }
    if (tool === 'text') {
      canvas.add(new Textbox('修正メモ', { fill: color, fontSize: 42, left: 120, top: 120, width: 260 }))
    }
    canvas.renderAll()
    undoManagerRef.current?.save()
  }, [color, tool])

  function deleteSelected() {
    const canvas = fabricRef.current
    if (!canvas) {
      return
    }

    const active = canvas.getActiveObject()
    if (active) {
      canvas.remove(active)
      canvas.discardActiveObject()
      canvas.renderAll()
      undoManagerRef.current?.save()
    }
  }

  function handleUndo() {
    undoManagerRef.current?.undo()
  }

  function handleRedo() {
    undoManagerRef.current?.redo()
  }

  function applyFilter(filterType: string, value: number) {
    const canvas = fabricRef.current
    if (!canvas) {
      return
    }

    const active = canvas.getActiveObject() || frameImageRef.current
    if (!active || active.type !== 'image') {
      return
    }

    const img = active as fabric.Image
    const existing = (img.filters ?? []).filter((f) => f.type !== filterType)
    let newFilter: filters.BaseFilter | undefined

    switch (filterType) {
      case 'Brightness':
        newFilter = new filters.Brightness({ brightness: (value - 50) / 50 })
        break
      case 'Contrast':
        newFilter = new filters.Contrast({ contrast: (value - 50) / 50 })
        break
      case 'Saturation':
        newFilter = new filters.Saturation({ saturation: (value - 50) / 50 })
        break
      case 'Blur':
        newFilter = new filters.Blur({ blur: value / 10 })
        break
      default:
        return
    }

    img.filters = [...existing, newFilter]
    img.applyFilters()
    canvas.renderAll()
    undoManagerRef.current?.save()
  }

  function saveFrame() {
    const canvas = fabricRef.current
    if (!canvas || !frame) {
      return
    }

    mutation.mutate({
      projectId,
      frameId: frame.id,
      imageDataUrl: canvas.toDataURL({ format: 'png', multiplier: 1 }),
      note: 'manual edit',
    })
  }

  return (
    <section className="editor-page">
      <div className="toolbar panel">
        {tools.map((item) => (
          <button
            className={tool === item.id ? 'icon-button active' : 'icon-button'}
            key={item.id}
            onClick={() => setTool(item.id)}
            title={item.label}
            type="button"
          >
            <item.icon aria-hidden="true" />
          </button>
        ))}
        <button className="icon-button" onClick={handleUndo} title="元に戻す" type="button">
          <Undo aria-hidden="true" />
        </button>
        <button className="icon-button" onClick={handleRedo} title="やり直す" type="button">
          <Redo aria-hidden="true" />
        </button>
        <button className="icon-button" onClick={deleteSelected} title="選択オブジェクトを削除" type="button">
          <Trash2 aria-hidden="true" />
        </button>
        <input aria-label="色" onChange={(event) => setColor(event.target.value)} type="color" value={color} />
        <input
          aria-label="ブラシサイズ"
          max={48}
          min={1}
          onChange={(event) => setBrushSize(Number(event.target.value))}
          type="range"
          value={brushSize}
        />
        <button className="command-button compact" onClick={addObject} type="button">
          追加
        </button>
        <button className="command-button compact" disabled={!frame || mutation.isPending} onClick={saveFrame} type="button">
          <Save aria-hidden="true" />
          保存
        </button>
      </div>
      <div className="toolbar panel">
        <span className="toolbar-label">
          <Sun aria-hidden="true" size={16} />
          明度
          <input max={100} min={0} onChange={(e) => applyFilter('Brightness', Number(e.target.value))} type="range" />
        </span>
        <span className="toolbar-label">
          <SlidersHorizontal aria-hidden="true" size={16} />
          コントラスト
          <input max={100} min={0} onChange={(e) => applyFilter('Contrast', Number(e.target.value))} type="range" />
        </span>
        <span className="toolbar-label">
          <SlidersHorizontal aria-hidden="true" size={16} />
          彩度
          <input max={100} min={0} onChange={(e) => applyFilter('Saturation', Number(e.target.value))} type="range" />
        </span>
        <span className="toolbar-label">
          <SlidersHorizontal aria-hidden="true" size={16} />
          ブラー
          <input max={50} min={0} onChange={(e) => applyFilter('Blur', Number(e.target.value))} type="range" />
        </span>
      </div>
      <div className="panel editor-canvas-panel">
        <canvas ref={canvasElementRef} />
      </div>
    </section>
  )
}
