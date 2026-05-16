import type { ExportVideoResult } from '@anifusion/contracts'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Download, Film } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { JobStatusPanel } from '@/components/shared/job-status'
import { apiClient } from '@/lib/api-client'
import { useFrameStore } from '@/stores/frame-store'

export function ExportPanel() {
  const projectId = useFrameStore((state) => state.projectId)
  const frames = useFrameStore((state) => state.frames)
  const activeJob = useFrameStore((state) => state.activeJob)
  const setActiveJob = useFrameStore((state) => state.setActiveJob)
  const [fps, setFps] = useState(8)
  const [previewIndex, setPreviewIndex] = useState(0)
  const exportJob = activeJob?.type === 'export' ? activeJob : undefined

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

  return (
    <section className="work-grid">
      <div className="panel primary-panel">
        <div className="section-heading">
          <span>Step 4</span>
          <h1>完成フレームを動画に書き出し</h1>
          <p>自動プレビューと最終MP4書き出しを分け、明示的な操作だけで動画化します。</p>
        </div>
        <label>
          FPS {fps}
          <input max={24} min={4} onChange={(event) => setFps(Number(event.target.value))} type="range" value={fps} />
        </label>
        <button
          className="command-button"
          disabled={frames.length === 0 || mutation.isPending}
          onClick={() => mutation.mutate({ projectId, fps })}
          type="button"
        >
          <Film aria-hidden="true" />
          MP4を書き出す
        </button>
        <JobStatusPanel job={exportJob} />
        {videoUrl ? (
          <a className="download-link" href={videoUrl}>
            <Download aria-hidden="true" />
            書き出し結果を開く
          </a>
        ) : null}
      </div>
      <div className="panel playback-panel">
        {frames[previewIndex] ? <img src={frames[previewIndex].imageUrl} alt="プレビュー" /> : <div>フレームがありません</div>}
      </div>
    </section>
  )
}
