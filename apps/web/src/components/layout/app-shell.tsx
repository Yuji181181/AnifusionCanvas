import { Link } from '@tanstack/react-router'
import { Clapperboard, Download, Eraser, PenTool, Sparkles } from 'lucide-react'
import type { ReactNode } from 'react'
import { Timeline } from '@/features/timeline/components/timeline'

const steps = [
  { to: '/step1', label: 'AI中割り', icon: Sparkles },
  { to: '/step2', label: 'Inpainting', icon: Eraser },
  { to: '/step3', label: '手動編集', icon: PenTool },
  { to: '/export', label: '書き出し', icon: Download },
] as const

type Props = {
  children: ReactNode
}

export function AppShell({ children }: Props) {
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <Clapperboard aria-hidden="true" />
          <div>
            <strong>Anifusion Canvas</strong>
            <span>Human-in-the-Loop animation studio</span>
          </div>
        </div>
        <nav className="step-nav" aria-label="workflow">
          {steps.map((step) => (
            <Link key={step.to} to={step.to} className="step-link" activeProps={{ className: 'step-link active' }}>
              <step.icon aria-hidden="true" />
              <span>{step.label}</span>
            </Link>
          ))}
        </nav>
      </aside>
      <main className="workspace">
        <div className="workspace-inner">{children}</div>
        <Timeline />
      </main>
    </div>
  )
}
