import type { GenerateFramesResult } from '@anifusion/contracts'
import { generationFormSchema, type GenerationFormValues } from '@/lib/form-schemas'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery } from '@tanstack/react-query'
import { AlertTriangle, ImagePlus, Play } from 'lucide-react'
import { useEffect, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { JobStatusPanel } from '@/components/shared/job-status'
import { readableError, RecoveryPanel } from '@/components/shared/recovery-panel'
import { apiClient } from '@/lib/api-client'
import { createDemoFrameDataUrl } from '@/lib/demo-images'
import { queryClient } from '@/lib/query-client'
import { useFrameStore } from '@/stores/frame-store'

export function GenerationPanel() {
  const projectId = useFrameStore((state) => state.projectId)
  const setFrames = useFrameStore((state) => state.setFrames)
  const activeJob = useFrameStore((state) => state.activeJob)
  const setActiveJob = useFrameStore((state) => state.setActiveJob)
  const generationJob = activeJob?.type === 'generation' ? activeJob : undefined

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<GenerationFormValues>({
    resolver: zodResolver(generationFormSchema),
    defaultValues: {
      projectId,
      prompt: '風に髪がなびきながら振り向く',
      negativePrompt: 'broken hands, distorted lines',
      frameCount: 6,
      startImageDataUrl: '',
      endImageDataUrl: '',
    },
  })

  const demoImages = useMemo(
    () => ({
      start: createDemoFrameDataUrl('KEY 1', 204),
      end: createDemoFrameDataUrl('KEY 2', 342),
    }),
    [],
  )

  useEffect(() => {
    setValue('startImageDataUrl', demoImages.start)
    setValue('endImageDataUrl', demoImages.end)
    setValue('projectId', projectId)
  }, [demoImages, projectId, setValue])

  const jobQuery = useQuery({
    queryKey: ['job', generationJob?.id],
    enabled: Boolean(generationJob && generationJob.status !== 'succeeded' && generationJob.status !== 'failed'),
    queryFn: () => apiClient.getJob<GenerateFramesResult>(generationJob!.id),
    refetchInterval: 900,
  })

  useEffect(() => {
    if (!jobQuery.data?.job) {
      return
    }

    setActiveJob(jobQuery.data.job)
    if (jobQuery.data.job.status === 'succeeded' && jobQuery.data.job.result?.frames) {
      setFrames(jobQuery.data.job.result.frames)
      queryClient.invalidateQueries({ queryKey: ['frames', projectId] })
    }
  }, [jobQuery.data, projectId, setActiveJob, setFrames])

  const mutation = useMutation({
    mutationFn: apiClient.generateFrames,
    onSuccess: (data) => setActiveJob(data.job),
  })

  function runGeneration(values: GenerationFormValues) {
    mutation.mutate({
      projectId: values.projectId,
      prompt: values.prompt,
      frameCount: values.frameCount,
      startImageDataUrl: values.startImageDataUrl,
      endImageDataUrl: values.endImageDataUrl,
      negativePrompt: values.negativePrompt,
    })
  }

  function clearFailure() {
    mutation.reset()
    if (generationJob?.status === 'failed') {
      setActiveJob(undefined)
    }
  }

  const startImage = watch('startImageDataUrl')
  const endImage = watch('endImageDataUrl')
  const isGenerating = mutation.isPending || Boolean(generationJob && generationJob.status !== 'failed' && generationJob.status !== 'succeeded')
  const failureMessage = readableError(mutation.error)
    ?? (generationJob?.status === 'failed' ? generationJob.error || generationJob.message : undefined)

  return (
    <section className="work-grid">
      <div className="panel primary-panel">
        <div className="section-heading">
          <span>Step 1</span>
          <h1>AIで中割りを生成</h1>
          <p>2枚の原画と動きの指示から、編集可能なフレーム列を生成します。</p>
        </div>
        <form className="form-grid" id="generation-form" onSubmit={handleSubmit(runGeneration)}>
          <label>
            動きの指示
            <textarea {...register('prompt')} rows={4} />
            {errors.prompt && (
              <span className="field-error"><AlertTriangle aria-hidden="true" size={14} /> {errors.prompt.message}</span>
            )}
          </label>
          <label>
            生成枚数
            <input
              {...register('frameCount', { valueAsNumber: true })}
              max={12}
              min={2}
              type="number"
            />
            {errors.frameCount && (
              <span className="field-error"><AlertTriangle aria-hidden="true" size={14} /> {errors.frameCount.message}</span>
            )}
          </label>
          <button className="command-button" disabled={isGenerating} type="submit">
            <Play aria-hidden="true" />
            生成を開始
          </button>
        </form>
        <RecoveryPanel
          formId="generation-form"
          message={failureMessage}
          onDismiss={clearFailure}
          title="生成を完了できませんでした"
        />
      </div>
      <div className="panel">
        <div className="preview-pair">
          <FrameUploadPreview
            imageUrl={startImage}
            label="原画 1"
            onChange={(url) => setValue('startImageDataUrl', url, { shouldValidate: true })}
          />
          <FrameUploadPreview
            imageUrl={endImage}
            label="原画 2"
            onChange={(url) => setValue('endImageDataUrl', url, { shouldValidate: true })}
          />
        </div>
        {errors.startImageDataUrl && (
          <p className="field-error"><AlertTriangle aria-hidden="true" size={14} /> {errors.startImageDataUrl.message}</p>
        )}
        {errors.endImageDataUrl && (
          <p className="field-error"><AlertTriangle aria-hidden="true" size={14} /> {errors.endImageDataUrl.message}</p>
        )}
        <JobStatusPanel job={generationJob} />
      </div>
    </section>
  )
}

type PreviewProps = {
  imageUrl: string
  label: string
  onChange: (value: string) => void
}

function FrameUploadPreview({ imageUrl, label, onChange }: PreviewProps) {
  function handleFile(file?: File) {
    if (!file) {
      return
    }

    const reader = new FileReader()
    reader.onload = () => onChange(String(reader.result))
    reader.readAsDataURL(file)
  }

  return (
    <label className="upload-preview">
      <input accept="image/*" onChange={(event) => handleFile(event.target.files?.[0])} type="file" />
      {imageUrl ? <img src={imageUrl} alt={label} /> : <ImagePlus aria-hidden="true" />}
      <span>{label}</span>
    </label>
  )
}
