# AnifusionCanvas

AI 画像生成をブラウザ上で行うアプリケーション。

- **Frontend**: React + Vite (Cloudflare Pages)
- **Backend**: Go + Echo (Cloud Run)
- **DB**: TiDB Cloud
- **Storage**: Cloudflare R2

## ローカル開発

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

## デプロイ

GitHub Actions から手動実行する。

1. **CI** → lint / build / test
2. **Deploy API to Cloud Run** → Docker build & push → Cloud Run deploy
3. **Deploy Web to Cloudflare Pages** → build & deploy

詳細は `docs/deployment-guide.md` を参照。

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