# GitHub Actions セットアップ

このドキュメントでは、Cloud Run / Cloudflare Pages のデプロイを GitHub Actions から実行するための設定をまとめる。

## 1. 追加済み workflow

- `.github/workflows/ci.yml`
- `.github/workflows/deploy-api.yml`
- `.github/workflows/deploy-api-secret-manager.yml`
- `.github/workflows/deploy-pages.yml`

## 2. あなたが GitHub に設定する Secrets

### Cloud Run 用

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_SERVICE_ACCOUNT_EMAIL`
- `DATABASE_URL`
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET`
- `R2_PUBLIC_BASE_URL`
- `R2_ENDPOINT_URL`
- `REPLICATE_API_TOKEN`

### Cloudflare Pages 用

- `CLOUDFLARE_API_TOKEN`

## 3. あなたがやる操作

### 3-1. GitHub Secrets の登録

1. GitHub リポジトリを開く
2. Settings → Secrets and variables → Actions
3. New repository secret から必要な secret を追加する

### 3-2. GCP Workload Identity の作成

GitHub Actions から安全に GCP へデプロイするには Workload Identity Federation を使う。ここはあなたの GCP 側 IAM 設定が必要。

最低限必要なのは次。

- Workload Identity Pool
- Workload Identity Provider
- GitHub Actions 用 service account
- その service account への Cloud Run / Artifact Registry 権限付与

## 4. 実行方法

GitHub の Actions タブから手動実行する。

まずは次の順で実行する。

- `CI`
- `Deploy API to Cloud Run`
- `Deploy Web to Cloudflare Pages`

`Deploy API to Cloud Run via Secret Manager` は、Secret Manager 登録と service account 権限付与まで終わってから実行する。

## 5. 補足

- 今回の workflow は `workflow_dispatch` のみ
- まずは手動実行で安定させ、その後 main push トリガーへ広げるのが安全
トリガーへ広げるのが安全
