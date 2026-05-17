import { AlertTriangle, RotateCcw, X } from 'lucide-react'

type Props = {
  actionLabel?: string
  formId: string
  message?: string
  onDismiss: () => void
  title: string
}

export function RecoveryPanel({ actionLabel = '再試行', formId, message, onDismiss, title }: Props) {
  if (!message) {
    return null
  }

  return (
    <div className="recovery-panel" role="alert">
      <div>
        <AlertTriangle aria-hidden="true" size={16} />
        <strong>{title}</strong>
      </div>
      <p>{message}</p>
      <div className="button-row">
        <button className="command-button compact" form={formId} type="submit">
          <RotateCcw aria-hidden="true" />
          {actionLabel}
        </button>
        <button className="icon-button" onClick={onDismiss} title="失敗表示を閉じる" type="button">
          <X aria-hidden="true" />
        </button>
      </div>
    </div>
  )
}

export function readableError(error: unknown): string | undefined {
  if (!error) {
    return undefined
  }
  if (error instanceof Error) {
    return error.message
  }
  return String(error)
}
