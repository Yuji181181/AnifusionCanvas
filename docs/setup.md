# Setup

## Frontend

```bash
bun install
bun run dev:web
```

## Backend

```bash
cd apps/api
go mod tidy
go run ./cmd/server
```

## Environment Variables

ルートの `.env` を使用する場合でも、実運用ではフロントエンド用・バックエンド用に分離して管理する。

### Backend Required

- `DATABASE_URL`
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET`
- `R2_ENDPOINT_URL`
- `REPLICATE_API_TOKEN`

### Backend Optional

- `APP_ENV`
- `API_PORT`
- `FRONTEND_ORIGIN`
- `R2_PUBLIC_BASE_URL`
- `R2_REGION`

### Frontend Required

- `VITE_API_BASE_URL`

## Docker Compose

```bash
docker compose up --build
```

確認先:

- Frontend: `http://127.0.0.1:3000`
- API health: `http://127.0.0.1:8080/health`

停止:

```bash
docker compose down
```

## Notes

- Go は `~/.local/share/go/bin` に導入済みで、`.bashrc` に PATH を追加済み
- Go 側の `DATABASE_URL` は Go MySQL ドライバ互換のDSN形式へ変換済み
- Cloud Run 本番では `FRONTEND_ORIGIN` をCloudflare PagesのURLへ設定する
- `R2_PUBLIC_BASE_URL` を使う場合は公開配信URLを設定する
- Docker Desktop を WSL2 から使うには、Docker Desktop 側で WSL integration を有効にする必要がある
