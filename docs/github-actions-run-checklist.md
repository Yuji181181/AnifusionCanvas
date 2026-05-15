# GitHub Actions 実行チェックリスト

## 先に実行する順番

1. `CI`
2. `Deploy API to Cloud Run`
3. `Deploy Web to Cloudflare Pages`

## いまの前提

- `GCP_WORKLOAD_IDENTITY_PROVIDER` は登録済み
- `GCP_SERVICE_ACCOUNT_EMAIL` は登録済み
- Cloud Run 用のアプリ secrets は GitHub Secrets に登録済み
- Cloudflare API Token は GitHub Secrets に登録済み

## 実行方法

1. GitHub リポジトリを開く
2. `Actions` タブを開く
3. 左で `CI` を選ぶ
4. `Run workflow` を押す
5. 完了後、`Deploy API to Cloud Run` を実行する
6. 完了後、`Deploy Web to Cloudflare Pages` を実行する

## 成功条件

### CI

- lint 成功
- build 成功
- test 成功

### Deploy API to Cloud Run

- `google-github-actions/auth` が成功する
- Docker build / push が成功する
- Cloud Run deploy が成功する

### Deploy Web to Cloudflare Pages

- Bun install が成功する
- Web build が成功する
- Pages deploy が成功する

## 失敗しやすいポイント

### API deploy で auth 失敗

確認するもの:

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_SERVICE_ACCOUNT_EMAIL`
- Workload Identity User binding

### API deploy で push 失敗

確認するもの:

- service account に `Artifact Registry Writer` があるか

### API deploy で Cloud Run deploy 失敗

確認するもの:

- service account に `Cloud Run Admin` があるか
- `.env` 由来の GitHub Secrets がすべて埋まっているか

### Pages deploy で失敗

確認するもの:

- `CLOUDFLARE_API_TOKEN`
- Pages project `anifusion-canvas` が存在するか
