import type { Job } from '@anifusion/contracts'

type Props = {
  job?: Job
}

export function JobStatusPanel({ job }: Props) {
  if (!job) {
    return <div className="status-panel muted">ジョブはまだありません</div>
  }

  return (
    <div className="status-panel">
      <div>
        <strong>{job.message}</strong>
        <span>{job.status}</span>
      </div>
      <progress value={job.progress} max={100} />
    </div>
  )
}
