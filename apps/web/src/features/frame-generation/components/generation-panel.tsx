import type { GenerateFramesResult } from '@anifusion/contracts'
import { useMutation, useQuery } from '@tanstack/react-query'
import { ImagePlus, Play } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { JobStatusPanel } from '@/components/shared/job-status'
import { apiClient } from '@/lib/api-client'
import { createDemoFrameDataUrl } from '@/lib/demo-images'
import { queryClient } from '@/lib/query-client'
import { useFrameStore } from '@/stores/frame-store'

export function GenerationPanel() {
  const projectId = useFrameStore((state) => state.projectId)
  const setFrames = useFrameStore((state) => state.setFrames)
  const activeJob = useFrameStore((state) => state.activeJob)
  const setActiveJob = useFrameStore((state) => state.setActiveJob)
  const [prompt, setPrompt] = useState('風に髪がなびきながら振り向く')
  const [frameCount, setFrameCount] = useState(6)
  const [startImage, setStartImage] = useState('')
  const [endImage, setEndImage] = useState('')

  const demoImages = useMemo(
    () => ({
      start: createDemoFrameDataUrl('KEY 1', 204),
      end: createDemoFrameDataUrl('KEY 2', 342),
    }),
    [],
  )

  useEffect(() => {
    setStartImage(demoImages.start)
    setEndImage(demoImages.end)
  }, [demoImages])

  const jobQuery = useQuery({
    queryKey: ['job', activeJob?.id],
    enabled: Boolean(activeJob && activeJob.status !== 'succeeded' && activeJob.status !== 'failed'),
    queryFn: () => apiClient.getJob<GenerateFramesResult>(activeJob!.id),
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

  function runGeneration() {
    mutation.mutate({
      projectId,
      prompt,
      frameCount,
      startImageDataUrl: startImage,
      endImageDataUrl: endImage,
      negativePrompt: 'broken hands, distorted lines',
    })
  }

  return (
    <section className="work-grid">
      <div className="panel primary-panel">
        <div className="section-heading">
          <span>Step 1</span>
          <h1>AIで中割りを生成</h1>
          <p>2枚の原画と動きの指示から、編集可能なフレーム列を生成します。</p>
        </div>
        <div className="form-grid">
          <label>
            動きの指示
            <textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} rows={4} />
          </label>
          <label>
            生成枚数
            <input
              max={12}
              min={2}
              onChange={(event) => setFrameCount(Number(event.target.value))}
              type="number"
              value={frameCount}
            />
          </label>
        </div>
        <button className="command-button" disabled={mutation.isPending} onClick={runGeneration} type="button">
          <Play aria-hidden="true" />
          生成を開始
        </button>
      </div>
      <div className="panel">
        <div className="preview-pair">
          <FrameUploadPreview imageUrl={startImage} label="原画 1" onChange={setStartImage} />
          <FrameUploadPreview imageUrl={endImage} label="原画 2" onChange={setEndImage} />
        </div>
        <JobStatusPanel job={activeJob} />
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
