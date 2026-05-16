import type { InpaintFrameResult } from '@anifusion/contracts'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Eraser, Wand2 } from 'lucide-react'
import { Canvas, PencilBrush } from 'fabric'
import { useEffect, useRef, useState } from 'react'
import { JobStatusPanel } from '@/components/shared/job-status'
import { apiClient } from '@/lib/api-client'
import { queryClient } from '@/lib/query-client'
import { useFrameStore, useSelectedFrame } from '@/stores/frame-store'

export function InpaintingPanel() {
  const projectId = useFrameStore((state) => state.projectId)
  const upsertFrame = useFrameStore((state) => state.upsertFrame)
  const activeJob = useFrameStore((state) => state.activeJob)
  const setActiveJob = useFrameStore((state) => state.setActiveJob)
  const frame = useSelectedFrame()
  const canvasElementRef = useRef<HTMLCanvasElement | null>(null)
  const fabricRef = useRef<Canvas | null>(null)
  const [prompt, setPrompt] = useState('手の形を自然な握りこぶしに修正')
  const [strength, setStrength] = useState(0.72)

  useEffect(() => {
    if (!canvasElementRef.current) {
      return
    }

    const canvas = new Canvas(canvasElementRef.current, {
      backgroundColor: 'rgba(255,255,255,0)',
      height: 360,
      isDrawingMode: true,
      width: 640,
    })
    const brush = new PencilBrush(canvas)
    brush.color = 'rgba(0,0,0,0.86)'
    brush.width = 24
    canvas.freeDrawingBrush = brush
    fabricRef.current = canvas

    return () => {
      canvas.dispose()
      fabricRef.current = null
    }
  }, [frame?.id])

  const jobQuery = useQuery({
    queryKey: ['job', activeJob?.id],
    enabled: Boolean(activeJob && activeJob.type === 'inpainting' && activeJob.status !== 'succeeded' && activeJob.status !== 'failed'),
    queryFn: () => apiClient.getJob<InpaintFrameResult>(activeJob!.id),
    refetchInterval: 900,
  })

  useEffect(() => {
    if (!jobQuery.data?.job) {
      return
    }

    setActiveJob(jobQuery.data.job)
    if (jobQuery.data.job.status === 'succeeded' && jobQuery.data.job.result?.frame) {
      upsertFrame(jobQuery.data.job.result.frame)
      queryClient.invalidateQueries({ queryKey: ['frames', projectId] })
    }
  }, [jobQuery.data, projectId, setActiveJob, upsertFrame])

  const mutation = useMutation({
    mutationFn: apiClient.inpaintFrame,
    onSuccess: (data) => setActiveJob(data.job),
  })

  function runInpaint() {
    if (!frame || !fabricRef.current) {
      return
    }

    mutation.mutate({
      projectId,
      frameId: frame.id,
      maskDataUrl: fabricRef.current.toDataURL({ format: 'png', multiplier: 1 }),
      prompt,
      strength,
    })
  }

  function clearMask() {
    fabricRef.current?.clear()
  }

  return (
    <section className="work-grid editor-grid">
      <div className="panel canvas-panel">
        {frame ? <img className="canvas-backdrop" src={frame.imageUrl} alt="選択中フレーム" /> : null}
        <canvas ref={canvasElementRef} />
      </div>
      <div className="panel side-panel">
        <div className="section-heading">
          <span>Step 2</span>
          <h1>AIで破綻部分だけ修正</h1>
          <p>黒いブラシでマスクを描き、自然言語で修正内容を指定します。</p>
        </div>
        <label>
          修正プロンプト
          <textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} rows={5} />
        </label>
        <label>
          変化量 {strength.toFixed(2)}
          <input
            max={1}
            min={0.1}
            onChange={(event) => setStrength(Number(event.target.value))}
            step={0.01}
            type="range"
            value={strength}
          />
        </label>
        <div className="button-row">
          <button className="icon-button" onClick={clearMask} title="マスクを消去" type="button">
            <Eraser aria-hidden="true" />
          </button>
          <button className="command-button" disabled={!frame || mutation.isPending} onClick={runInpaint} type="button">
            <Wand2 aria-hidden="true" />
            修正を実行
          </button>
        </div>
        <JobStatusPanel job={activeJob?.type === 'inpainting' ? activeJob : undefined} />
      </div>
    </section>
  )
}
