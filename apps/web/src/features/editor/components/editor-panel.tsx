import { useMutation } from '@tanstack/react-query'
import { Circle, Copy, Eye, EyeOff, Hexagon, Layers, MousePointer2, MoveDown, MoveUp, PenLine, Redo, Save, SlidersHorizontal, Square, Sun, Trash2, Type, Undo } from 'lucide-react'
import { Canvas, Circle as FabricCircle, PencilBrush, Polygon, Rect, Textbox, filters, type FabricObject } from 'fabric'
import { useCallback, useEffect, useRef, useState } from 'react'
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
  { id: 'polygon', icon: Hexagon, label: '多角形' },
  { id: 'text', icon: Type, label: 'テキスト' },
] as const

type LayerItem = {
  id: string
  index: number
  isActive: boolean
  isVisible: boolean
  label: string
}

function layerLabel(object: FabricObject) {
  if (object.type === 'textbox') {
    return 'テキスト'
  }
  if (object.type === 'rect') {
    return '四角'
  }
  if (object.type === 'circle') {
    return '円'
  }
  if (object.type === 'polygon') {
    return '多角形'
  }
  if (object.type === 'path') {
    return 'ペン'
  }
  if (object.type === 'image') {
    return '画像'
  }
  return 'レイヤー'
}

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
  const objectIdsRef = useRef<WeakMap<FabricObject, string>>(new WeakMap())
  const nextObjectIdRef = useRef(1)
  const [layers, setLayers] = useState<LayerItem[]>([])

  const syncLayers = useCallback(() => {
    const canvas = fabricRef.current
    if (!canvas) {
      setLayers([])
      return
    }

    const active = canvas.getActiveObject()
    const items = canvas
      .getObjects()
      .map((object, index) => ({ object, index }))
      .filter(({ object }) => object !== frameImageRef.current)
      .map(({ object, index }) => {
        let id = objectIdsRef.current.get(object)
        if (!id) {
          id = `layer-${nextObjectIdRef.current}`
          nextObjectIdRef.current += 1
          objectIdsRef.current.set(object, id)
        }

        return {
          id,
          index,
          isActive: object === active,
          isVisible: object.visible !== false,
          label: layerLabel(object),
        }
      })
      .reverse()

    setLayers(items)
  }, [])

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
    objectIdsRef.current = new WeakMap()
    nextObjectIdRef.current = 1
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
          syncLayers()
        },
        { crossOrigin: 'anonymous' },
      )
    } else {
      undoManager.save()
      syncLayers()
    }

    const saveAndSync = () => {
      undoManager.save()
      syncLayers()
    }
    const syncOnly = () => syncLayers()

    canvas.on('object:added', syncOnly)
    canvas.on('object:removed', syncOnly)
    canvas.on('object:modified', saveAndSync)
    canvas.on('path:created', saveAndSync)
    canvas.on('selection:created', syncOnly)
    canvas.on('selection:updated', syncOnly)
    canvas.on('selection:cleared', syncOnly)

    return () => {
      canvas.off('object:added', syncOnly)
      canvas.off('object:removed', syncOnly)
      canvas.off('object:modified', saveAndSync)
      canvas.off('path:created', saveAndSync)
      canvas.off('selection:created', syncOnly)
      canvas.off('selection:updated', syncOnly)
      canvas.off('selection:cleared', syncOnly)
      canvas.dispose()
      fabricRef.current = null
      undoManagerRef.current = null
      frameImageRef.current = null
      setLayers([])
    }
  }, [frame?.id, frame?.imageUrl, syncLayers])

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
    if (tool === 'polygon') {
      canvas.add(new Polygon([
        { x: 64, y: 0 },
        { x: 128, y: 46 },
        { x: 104, y: 120 },
        { x: 24, y: 120 },
        { x: 0, y: 46 },
      ], { fill: color, left: 120, top: 90 }))
    }
    if (tool === 'text') {
      canvas.add(new Textbox('修正メモ', { fill: color, fontSize: 42, left: 120, top: 120, width: 260 }))
    }
    canvas.renderAll()
    undoManagerRef.current?.save()
    syncLayers()
  }, [color, syncLayers, tool])

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
      syncLayers()
    }
  }

  function selectedEditableObject(): FabricObject | undefined {
    const active = fabricRef.current?.getActiveObject()
    if (!active || active === frameImageRef.current) {
      return undefined
    }
    return active
  }

  function keepAboveFrameImage(object: FabricObject) {
    const canvas = fabricRef.current
    const frameImage = frameImageRef.current
    if (!canvas || !frameImage) {
      return
    }

    const frameIndex = canvas.getObjects().indexOf(frameImage)
    const objectIndex = canvas.getObjects().indexOf(object)
    if (objectIndex <= frameIndex) {
      canvas.moveObjectTo(object, frameIndex + 1)
    }
  }

  function objectAtLayerIndex(index: number) {
    return fabricRef.current?.getObjects()[index]
  }

  function selectLayer(index: number) {
    const canvas = fabricRef.current
    const object = objectAtLayerIndex(index)
    if (!canvas || !object || object.visible === false) {
      return
    }

    canvas.setActiveObject(object)
    canvas.renderAll()
    syncLayers()
  }

  function toggleLayerVisibility(index: number) {
    const canvas = fabricRef.current
    const object = objectAtLayerIndex(index)
    if (!canvas || !object || object === frameImageRef.current) {
      return
    }

    const nextVisible = object.visible === false
    object.set('visible', nextVisible)
    if (!nextVisible && canvas.getActiveObject() === object) {
      canvas.discardActiveObject()
    }
    canvas.renderAll()
    undoManagerRef.current?.save()
    syncLayers()
  }

  async function duplicateSelected() {
    const canvas = fabricRef.current
    const active = selectedEditableObject()
    if (!canvas || !active) {
      return
    }

    const clone = await active.clone()
    clone.set({
      left: (active.left ?? 0) + 24,
      top: (active.top ?? 0) + 24,
    })
    canvas.add(clone)
    canvas.setActiveObject(clone)
    canvas.renderAll()
    undoManagerRef.current?.save()
    syncLayers()
  }

  function bringSelectedForward() {
    const canvas = fabricRef.current
    const active = selectedEditableObject()
    if (!canvas || !active) {
      return
    }

    canvas.bringObjectForward(active)
    canvas.renderAll()
    undoManagerRef.current?.save()
    syncLayers()
  }

  function sendSelectedBackward() {
    const canvas = fabricRef.current
    const active = selectedEditableObject()
    if (!canvas || !active) {
      return
    }

    canvas.sendObjectBackwards(active)
    keepAboveFrameImage(active)
    canvas.renderAll()
    undoManagerRef.current?.save()
    syncLayers()
  }

  function bringSelectedToFront() {
    const canvas = fabricRef.current
    const active = selectedEditableObject()
    if (!canvas || !active) {
      return
    }

    canvas.bringObjectToFront(active)
    canvas.renderAll()
    undoManagerRef.current?.save()
    syncLayers()
  }

  function handleUndo() {
    undoManagerRef.current?.undo()
    window.setTimeout(syncLayers, 0)
  }

  function handleRedo() {
    undoManagerRef.current?.redo()
    window.setTimeout(syncLayers, 0)
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
        <button className="icon-button" onClick={duplicateSelected} title="選択オブジェクトを複製" type="button">
          <Copy aria-hidden="true" />
        </button>
        <button className="icon-button" onClick={sendSelectedBackward} title="選択オブジェクトを背面へ" type="button">
          <MoveDown aria-hidden="true" />
        </button>
        <button className="icon-button" onClick={bringSelectedForward} title="選択オブジェクトを前面へ" type="button">
          <MoveUp aria-hidden="true" />
        </button>
        <button className="icon-button" onClick={bringSelectedToFront} title="選択オブジェクトを最前面へ" type="button">
          <Layers aria-hidden="true" />
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
      <div className="editor-workspace">
        <div className="panel editor-canvas-panel">
          <canvas ref={canvasElementRef} />
        </div>
        <aside className="panel layer-panel" aria-label="レイヤー一覧">
          <div className="layer-panel-heading">
            <Layers aria-hidden="true" size={16} />
            <strong>レイヤー</strong>
          </div>
          <div className="layer-list">
            {layers.length === 0 ? (
              <p className="layer-empty">追加した編集レイヤーはありません</p>
            ) : (
              layers.map((layer, visibleIndex) => (
                <div className={layer.isActive ? 'layer-item active' : 'layer-item'} key={layer.id}>
                  <button
                    aria-pressed={layer.isActive}
                    className="layer-select-button"
                    onClick={() => selectLayer(layer.index)}
                    type="button"
                  >
                    <span>{layer.label}</span>
                    <small>#{layers.length - visibleIndex}</small>
                  </button>
                  <button
                    className={layer.isVisible ? 'icon-button' : 'icon-button muted'}
                    onClick={() => toggleLayerVisibility(layer.index)}
                    title={layer.isVisible ? 'レイヤーを非表示' : 'レイヤーを表示'}
                    type="button"
                  >
                    {layer.isVisible ? <Eye aria-hidden="true" /> : <EyeOff aria-hidden="true" />}
                  </button>
                </div>
              ))
            )}
          </div>
        </aside>
      </div>
    </section>
  )
}
