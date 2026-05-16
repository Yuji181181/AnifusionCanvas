import type { ExportVideoResult } from '@anifusion/contracts'
import type { ExportFormValues } from '@/lib/form-schemas'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery } from '@tanstack/react-query'
import { AlertTriangle, Download, Film, Play, RotateCcw } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { JobStatusPanel } from '@/components/shared/job-status'
import { apiClient } from '@/lib/api-client'
import { useFrameStore } from '@/stores/frame-store'

export function ExportPanel() {
  const projectId = useFrameStore((state) => state.projectId)
  const frames = useFrameStore((state) => state.frames)
  const activeJob = useFrameStore((state) => state.activeJob)
  const setActiveJob = useFrameStore((state) => state.setActiveJob)
  const exportJob = activeJob?.type === 'export' ? activeJob : undefined

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<ExportFormValues>({
    resolver: zodResolver(exportFormSchema),
    defaultValues: {
      projectId,
      fps: 8,
    },
  })

  const fps = watch('fps')
  const [previewIndex, setPreviewIndex] = useState(0)
  const [videoPlaying, setVideoPlaying] = useState(false)
  const videoRef = useRef<HTMLVideoElement | null>(null)

  useEffect(() => {
    if (frames.length === 0) {
      return
    }

    const timer = window.setInterval(() => {
      setPreviewIndex((value) => (value + 1) % frames.length)
    }, Math.max(80, 1000 / fps))

    return () => window.clearInterval(timer)
  }, [fps, frames.length])

  const jobQuery = useQuery({
    queryKey: ['job', exportJob?.id],
    enabled: Boolean(exportJob && exportJob.status !== 'succeeded' && exportJob.status !== 'failed'),
    queryFn: () => apiClient.getJob<ExportVideoResult>(exportJob!.id),
    refetchInterval: 900,
  })

  useEffect(() => {
    if (jobQuery.data?.job) {
      setActiveJob(jobQuery.data.job)
    }
  }, [jobQuery.data, setActiveJob])

  const mutation = useMutation({
    mutationFn: apiClient.exportVideo,
    onSuccess: (data) => setActiveJob(data.job),
  })

  const videoUrl = useMemo(() => {
    if (exportJob?.status === 'succeeded' && exportJob.result?.videoUrl) {
      return exportJob.result.videoUrl
    }
    return undefined
  }, [exportJob])

  function runExport(values: ExportFormValues) {
    mutation.mutate({ projectId: values.projectId, fps: values.fps })
  }

  function handleReExport() {
    setActiveJob(undefined)
    videoRef.current?.pause()
  }

  function handlePlayVideo() {
    videoRef.current?.play()
    setVideoPlaying(true)
  }

  const isExporting = mutation.isPending || Boolean(exportJob && exportJob.status !== 'succeeded' && exportJob.status !== 'failed')

  return (
    <section className="work-grid">
      <div className="panel primary-panel">
        <div className="section-heading">
          <span>Step 4</span>
          <h1>完成フレームを動画に書き出し</h1>
          <p>自動プレビューと最終MP4書き出しを分け、明示的な操作だけで動画化します。</p>
        </div>
        <form onSubmit={handleSubmit(runExport)}>
          <label>
            FPS {fps}
            <input {...register('fps', { valueAsNumber: true })} max={24} min={4} type="range" />
            {errors.fps && (
              <span className="field-error"><AlertTriangle aria-hidden="true" size={14} /> {errors.fps.message}</span>
            )}
          </label>
          {videoUrl ? (
            <div className="export-actions">
              <a className="command-button" download href={videoUrl}>
                <Download aria-hidden="true" />
                MP4をダウンロード
              </a>
              <button className="icon-button" onClick={handleReExport} type="button">
                <RotateCcw aria-hidden="true" />
                再書き出し
              </button>
            </div>
          ) : (
            <button
              className="command-button"
              disabled={frames.length === 0 || isExporting}
              type="submit"
            >
              <Film aria-hidden="true" />
              MP4を書き出す
            </button>
          )}
        </form>
        <JobStatusPanel job={exportJob} />
      </div>
      <div className="panel playback-panel">
        {videoUrl ? (
          <>
            <video
              ref={videoRef}
              controls
              onPause={() => setVideoPlaying(false)}
              onPlay={() => setVideoPlaying(true)}
              src={videoUrl}
              style={{ height: '100%', width: '100%' }}
            />
            {!videoPlaying && (
              <button className="play-overlay" onClick={handlePlayVideo} type="button">
                <Play aria-hidden="true" size={48} />
              </button>
            )}
          </>
        ) : (
          frames[previewIndex] ? (
            <img src={frames[previewIndex].imageUrl} alt="プレビュー" />
          ) : (
            <div>フレームがありません</div>
          )
        )}
      </div>
    </section>
  )
}
