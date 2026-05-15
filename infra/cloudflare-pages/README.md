# Cloudflare Pages

## Build Settings

- Framework preset: None
- Build command: `bun install && bun run build:web`
- Build output directory: `apps/web/dist`
- Root directory: `/`

## Required Environment Variables

- `VITE_API_BASE_URL`: Cloud Run の公開URL

## Notes

- `VITE_` プレフィックス付き変数はクライアントに露出する
- Secrets は Cloudflare Pages ではなく Cloud Run 側へ置く
