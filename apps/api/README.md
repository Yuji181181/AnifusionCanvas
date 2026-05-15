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
