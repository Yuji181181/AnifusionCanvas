export function createDemoFrameDataUrl(label: string, hue: number): string {
  const canvas = document.createElement('canvas')
  canvas.width = 640
  canvas.height = 360
  const ctx = canvas.getContext('2d')

  if (!ctx) {
    return ''
  }

  const gradient = ctx.createLinearGradient(0, 0, 640, 360)
  gradient.addColorStop(0, `hsl(${hue}, 72%, 58%)`)
  gradient.addColorStop(1, `hsl(${(hue + 72) % 360}, 68%, 42%)`)
  ctx.fillStyle = gradient
  ctx.fillRect(0, 0, 640, 360)

  ctx.fillStyle = 'rgba(255,255,255,0.9)'
  ctx.beginPath()
  ctx.arc(320, 180, 88, 0, Math.PI * 2)
  ctx.fill()

  ctx.fillStyle = '#18202f'
  ctx.font = '700 42px system-ui'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(label, 320, 180)

  return canvas.toDataURL('image/png')
}
