# AnifusionCanvas

## Local Development

```bash
bun install
source ~/.bashrc
cd apps/api && go mod tidy && cd ../..
docker compose up --build
```

- Frontend: http://localhost:3000
- API health: http://localhost:8080/health

停止:

```bash
docker compose down
```

## Deployment

- Cloud Run / Pages の設定は `docs/deployment-guide.md` を参照
- 事前確認は `docs/deployment-checklist.md` を参照
- CLI 認証手順は `docs/cli-authentication-guide.md` を参照
- GitHub Actions 化は `docs/github-actions-setup.md` を参照
- GitHub Actions 実行確認は `docs/github-actions-run-checklist.md` を参照
- Docker 認証の補足は `docs/docker-auth-note.md` を参照
- Secret Manager 化は `docs/secret-manager-setup.md` を参照
- Cloud Run 運用強化は `docs/cloud-run-ops.md` を参照
- 残りの環境構築は `docs/remaining-env-setup.md` を参照
- 外部サービスの画面操作手順は `docs/external-console-setup-guide.md` を参照
- Workload Identity Provider エラー時は `docs/workload-identity-error-recovery.md` を参照
- Cloudflare Workers Builds 失敗時は `docs/cloudflare-workers-builds-note.md` を参照
- `.env` は `deploy.sh` から直接読み込まれるため、コメント行を含んでもそのまま使える

## Direct Run

### Frontend

```bash
bun run dev:web
```

### Backend

```bash
source ~/.bashrc
cd apps/api
go run ./cmd/server
```
