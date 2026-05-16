import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { apiClient } from '@/lib/api-client'
import { useFrameStore } from '@/stores/frame-store'

export function Timeline() {
  const projectId = useFrameStore((state) => state.projectId)
  const frames = useFrameStore((state) => state.frames)
  const selectedFrameId = useFrameStore((state) => state.selectedFrameId)
  const setFrames = useFrameStore((state) => state.setFrames)
  const setSelectedFrameId = useFrameStore((state) => state.setSelectedFrameId)

  const framesQuery = useQuery({
    queryKey: ['frames', projectId],
    queryFn: () => apiClient.listFrames(projectId),
  })

  useEffect(() => {
    if (framesQuery.data) {
      setFrames(framesQuery.data.frames)
    }
  }, [framesQuery.data, setFrames])

  return (
    <section className="timeline" aria-label="frame timeline">
      {frames.length === 0 ? (
        <div className="timeline-empty">2枚の原画から生成すると、ここにフレームが並びます</div>
      ) : (
        frames.map((frame) => (
          <button
            className={frame.id === selectedFrameId ? 'frame-thumb selected' : 'frame-thumb'}
            key={frame.id}
            onClick={() => setSelectedFrameId(frame.id)}
            type="button"
          >
            <img src={frame.thumbnailUrl || frame.imageUrl} alt={`${frame.index + 1}枚目`} />
            <span>{frame.index + 1}</span>
          </button>
        ))
      )}
    </section>
  )
}
