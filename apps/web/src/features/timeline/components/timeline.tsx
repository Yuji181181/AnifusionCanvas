import { useMutation, useQuery } from '@tanstack/react-query'
import { GripVertical, Pencil, Sparkles, Wand2 } from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { queryClient } from '@/lib/query-client'
import { useFrameStore } from '@/stores/frame-store'

const kindLabels: Record<string, string> = {
  key: '原画',
  generated: '生成',
  inpainted: '修正',
  edited: '編集',
}

const kindIcons: Record<string, React.ReactNode> = {
  key: undefined,
  generated: <Sparkles size={10} />,
  inpainted: <Wand2 size={10} />,
  edited: <Pencil size={10} />,
}

export function Timeline() {
  const projectId = useFrameStore((state) => state.projectId)
  const frames = useFrameStore((state) => state.frames)
  const selectedFrameId = useFrameStore((state) => state.selectedFrameId)
  const setFrames = useFrameStore((state) => state.setFrames)
  const setSelectedFrameId = useFrameStore((state) => state.setSelectedFrameId)
  const [dragIndex, setDragIndex] = useState<number | null>(null)
  const dragOverRef = useRef<number | null>(null)

  const framesQuery = useQuery({
    queryKey: ['frames', projectId],
    queryFn: () => apiClient.listFrames(projectId),
  })

  useEffect(() => {
    if (framesQuery.data) {
      setFrames(framesQuery.data.frames)
    }
  }, [framesQuery.data, setFrames])

  const reorderMutation = useMutation({
    mutationFn: apiClient.reorderFrames,
    onSuccess: (data) => {
      setFrames(data.frames)
      queryClient.invalidateQueries({ queryKey: ['frames', projectId] })
    },
  })

  const handleDragStart = useCallback((e: React.DragEvent, index: number) => {
    setDragIndex(index)
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(index))
  }, [])

  const handleDragOver = useCallback((e: React.DragEvent, index: number) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    dragOverRef.current = index
  }, [])

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      if (dragIndex === null || dragOverRef.current === null || dragIndex === dragOverRef.current) {
        setDragIndex(null)
        dragOverRef.current = null
        return
      }

      const newFrames = [...frames]
      const [moved] = newFrames.splice(dragIndex, 1)
      newFrames.splice(dragOverRef.current, 0, moved)
      setFrames(newFrames)

      reorderMutation.mutate({
        projectId,
        frameIds: newFrames.map((f) => f.id),
      })

      setDragIndex(null)
      dragOverRef.current = null
    },
    [dragIndex, frames, projectId, reorderMutation, setFrames],
  )

  const handleDragEnd = useCallback(() => {
    setDragIndex(null)
    dragOverRef.current = null
  }, [])

  if (frames.length === 0) {
    return (
      <section className="timeline" aria-label="frame timeline" onDragOver={(e) => e.preventDefault()} onDrop={handleDrop}>
        <div className="timeline-empty">2枚の原画から生成すると、ここにフレームが並びます</div>
      </section>
    )
  }

  return (
    <section className="timeline" aria-label="frame timeline">
      {frames.map((frame, index) => (
        <button
          className={frame.id === selectedFrameId ? 'frame-thumb selected' : 'frame-thumb'}
          draggable
          key={frame.id}
          onClick={() => setSelectedFrameId(frame.id)}
          onDragEnd={handleDragEnd}
          onDragOver={(e) => handleDragOver(e, index)}
          onDragStart={(e) => handleDragStart(e, index)}
          type="button"
        >
          <GripVertical aria-hidden="true" className="drag-handle" size={12} />
          <img src={frame.thumbnailUrl || frame.imageUrl} alt={`${frame.index + 1}枚目`} />
          <span className="frame-index">{frame.index + 1}</span>
          {frame.kind !== 'key' && (
            <span className={`frame-kind-tag frame-kind-${frame.kind}`}>
              {kindIcons[frame.kind]}
              {kindLabels[frame.kind] ?? frame.kind}
            </span>
          )}
        </button>
      ))}
    </section>
  )
}
