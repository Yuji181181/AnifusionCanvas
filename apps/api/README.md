# API

## Run

```bash
source ~/.bashrc
go mod tidy
go run ./cmd/server
```

## Health Check Example

```bash
curl http://127.0.0.1:8080/health
```

## Health Check

- `GET /health`
- `GET /health/dependencies`

`/health/dependencies` は外部依存の状態を返す。未設定の依存は `skipped`、接続失敗は `error` になる。

```bash
curl http://127.0.0.1:8080/health/dependencies
```

## Database Migrations

TiDB/MySQL 向けのマイグレーションは `internal/infrastructure/db/migrations` に置く。

```bash
migrate -path internal/infrastructure/db/migrations -database "$DATABASE_URL" up
```

SQL query は `internal/infrastructure/db/query` に置き、sqlc 生成対象にする。

DB-backed store を有効化する場合は次を設定する。

```bash
STUDIO_STORE=database
```

既存DBのスキーマが合わない場合は `docs/database-setup.md` を参照する。

## Required Environment Variables

- `DATABASE_URL`
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET`
- `R2_ENDPOINT_URL`
- `REPLICATE_API_TOKEN`

## Optional Environment Variables

- `APP_ENV`
- `API_PORT`
- `FRONTEND_ORIGIN`
- `R2_PUBLIC_BASE_URL`
- `R2_REGION`
