import { useMutation } from '@tanstack/react-query'
import { Circle, MousePointer2, PenLine, Save, Square, Type } from 'lucide-react'
import { Canvas, Circle as FabricCircle, PencilBrush, Rect, Textbox } from 'fabric'
import { useEffect, useRef } from 'react'
import { apiClient } from '@/lib/api-client'
import { queryClient } from '@/lib/query-client'
import { useEditorStore } from '@/stores/editor-store'
import { useFrameStore, useSelectedFrame } from '@/stores/frame-store'

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

    if (frame?.imageUrl) {
      canvas.backgroundImage = undefined
      const image = new Image()
      image.crossOrigin = 'anonymous'
      image.onload = () => {
        canvas.getContext().drawImage(image, 0, 0, 960, 540)
        canvas.renderAll()
      }
      image.src = frame.imageUrl
    }

    return () => {
      canvas.dispose()
      fabricRef.current = null
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

  function addObject() {
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
      <div className="panel editor-canvas-panel">
        <canvas ref={canvasElementRef} />
      </div>
    </section>
  )
}
