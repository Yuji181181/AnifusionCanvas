import type { InpaintFrameResult } from '@anifusion/contracts'
import { inpaintingFormSchema, type InpaintingFormValues } from '@/lib/form-schemas'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery } from '@tanstack/react-query'
import { AlertTriangle, Eraser, Eye, EyeOff, Undo, Wand2 } from 'lucide-react'
import { Canvas, PencilBrush } from 'fabric'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
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
  const historyRef = useRef<string[]>([])
  const historyIndexRef = useRef(-1)
  const [brushSize, setBrushSize] = useState(24)
  const [isMaskVisible, setIsMaskVisible] = useState(true)

  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<InpaintingFormValues>({
    resolver: zodResolver(inpaintingFormSchema),
    defaultValues: {
      projectId,
      frameId: frame?.id ?? '',
      prompt: '手の形を自然な握りこぶしに修正',
      maskDataUrl: '',
      strength: 0.72,
    },
  })

  const syncMaskValue = useCallback((shouldValidate = false) => {
    const canvas = fabricRef.current
    if (!canvas || canvas.getObjects().length === 0) {
      setValue('maskDataUrl', '', { shouldValidate })
      return
    }

    setValue('maskDataUrl', canvas.toDataURL({ format: 'png', multiplier: 1 }), { shouldValidate })
  }, [setValue])

  const saveMaskHistory = useCallback(() => {
    const canvas = fabricRef.current
    if (!canvas) {
      return
    }

    const nextHistory = historyRef.current.slice(0, historyIndexRef.current + 1)
    nextHistory.push(JSON.stringify(canvas.toJSON()))
    historyRef.current = nextHistory.slice(-30)
    historyIndexRef.current = historyRef.current.length - 1
    syncMaskValue(true)
  }, [syncMaskValue])

  useEffect(() => {
    setValue('projectId', projectId)
    setValue('frameId', frame?.id ?? '')
  }, [projectId, frame?.id, setValue])

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
    brush.width = brushSize
    canvas.freeDrawingBrush = brush
    fabricRef.current = canvas
    historyRef.current = [JSON.stringify(canvas.toJSON())]
    historyIndexRef.current = 0
    syncMaskValue()
    canvas.on('path:created', saveMaskHistory)

    return () => {
      canvas.off('path:created', saveMaskHistory)
      canvas.dispose()
      fabricRef.current = null
      historyRef.current = []
      historyIndexRef.current = -1
    }
  }, [frame?.id, saveMaskHistory, syncMaskValue])

  useEffect(() => {
    const canvas = fabricRef.current
    if (!canvas?.freeDrawingBrush) {
      return
    }

    canvas.freeDrawingBrush.width = brushSize
  }, [brushSize])

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

  function runInpaint(values: InpaintingFormValues) {
    if (!fabricRef.current) {
      return
    }

    const maskDataUrl = fabricRef.current.toDataURL({ format: 'png', multiplier: 1 })
    mutation.mutate({
      projectId: values.projectId,
      frameId: values.frameId,
      maskDataUrl,
      prompt: values.prompt,
      strength: values.strength,
    })
  }

  function clearMask() {
    const canvas = fabricRef.current
    if (!canvas || canvas.getObjects().length === 0) {
      return
    }

    canvas.clear()
    canvas.renderAll()
    saveMaskHistory()
  }

  function undoMask() {
    const canvas = fabricRef.current
    if (!canvas || historyIndexRef.current <= 0) {
      return
    }

    historyIndexRef.current -= 1
    canvas.loadFromJSON(JSON.parse(historyRef.current[historyIndexRef.current])).then(() => {
      canvas.renderAll()
      syncMaskValue(true)
    })
  }

  const isRunning = mutation.isPending || Boolean(activeJob?.type === 'inpainting' && activeJob?.status !== 'succeeded' && activeJob?.status !== 'failed')

  return (
    <section className="work-grid editor-grid">
      <div className={isMaskVisible ? 'panel canvas-panel' : 'panel canvas-panel mask-hidden'}>
        {frame ? <img className="canvas-backdrop" src={frame.imageUrl} alt="選択中フレーム" /> : null}
        <canvas ref={canvasElementRef} />
      </div>
      <div className="panel side-panel">
        <div className="section-heading">
          <span>Step 2</span>
          <h1>AIで破綻部分だけ修正</h1>
          <p>黒いブラシでマスクを描き、自然言語で修正内容を指定します。</p>
        </div>
        <form onSubmit={handleSubmit(runInpaint)}>
          <input type="hidden" {...register('maskDataUrl')} />
          <label>
            修正プロンプト
            <textarea {...register('prompt')} rows={5} />
            {errors.prompt && (
              <span className="field-error"><AlertTriangle aria-hidden="true" size={14} /> {errors.prompt.message}</span>
            )}
          </label>
          <label>
            変化量
            <input
              {...register('strength', { valueAsNumber: true })}
              max={1}
              min={0.1}
              step={0.01}
              type="range"
            />
            {errors.strength && (
              <span className="field-error"><AlertTriangle aria-hidden="true" size={14} /> {errors.strength.message}</span>
            )}
          </label>
          <label>
            ブラシサイズ
            <input
              aria-label="マスクのブラシサイズ"
              max={72}
              min={4}
              onChange={(event) => setBrushSize(Number(event.target.value))}
              step={1}
              type="range"
              value={brushSize}
            />
          </label>
          {errors.frameId && (
            <p className="field-error"><AlertTriangle aria-hidden="true" size={14} /> {errors.frameId.message}</p>
          )}
          {errors.maskDataUrl && (
            <p className="field-error"><AlertTriangle aria-hidden="true" size={14} /> {errors.maskDataUrl.message}</p>
          )}
          <div className="button-row">
            <button className="icon-button" onClick={undoMask} title="マスクを元に戻す" type="button">
              <Undo aria-hidden="true" />
            </button>
            <button className="icon-button" onClick={clearMask} title="マスクを消去" type="button">
              <Eraser aria-hidden="true" />
            </button>
            <button
              className={isMaskVisible ? 'icon-button active' : 'icon-button'}
              onClick={() => setIsMaskVisible((value) => !value)}
              title={isMaskVisible ? 'マスクを非表示' : 'マスクを表示'}
              type="button"
            >
              {isMaskVisible ? <Eye aria-hidden="true" /> : <EyeOff aria-hidden="true" />}
            </button>
            <button className="command-button" disabled={!frame || isRunning} type="submit">
              <Wand2 aria-hidden="true" />
              修正を実行
            </button>
          </div>
        </form>
        <JobStatusPanel job={activeJob?.type === 'inpainting' ? activeJob : undefined} />
      </div>
    </section>
  )
}
