# Cloudflare Workers Builds 失敗時のメモ

`Workers Builds: anifusioncanvas` が失敗していても、今回の構成では必ずしも問題ではない。

## 理由

このリポジトリでは、Cloudflare 側の Git 自動ビルドに依存せず、次のどちらかでデプロイできる。

1. `wrangler pages deploy`
2. GitHub Actions の `Deploy Web to Cloudflare Pages`

実際に CLI では Pages デプロイ成功を確認済み。

## よくある失敗原因

- Cloudflare Dashboard 側の Build command が古い
- Root directory / Output directory がずれている
- 環境変数 `VITE_API_BASE_URL` が未設定
- Git integration 側の project 名や framework preset がずれている

## このプロジェクトで正しい値

- Build command: `bun install && bun run build:web`
- Build output directory: `apps/web/dist`
- Root directory: `/`
- Environment variable: `VITE_API_BASE_URL=https://anifusion-api-976317870900.asia-northeast1.run.app`

## どれを正とするか

今後の運用では、次を正とする。

- GitHub Actions: `Deploy Web to Cloudflare Pages`

Cloudflare Dashboard 側の `Workers Builds` が失敗していても、GitHub Actions / CLI deploy が成功していれば運用上は進められる。

## もし Cloudflare Dashboard 側の自動ビルドも成功させたい場合

Cloudflare Dashboard で `anifusion-canvas` project の Build settings を次に合わせる。

- Framework preset: `None` または `Vite`
- Build command: `bun install && bun run build:web`
- Build output directory: `apps/web/dist`
- Root directory: `/`
- Environment variable: `VITE_API_BASE_URL`

そのうえで再デプロイする。
