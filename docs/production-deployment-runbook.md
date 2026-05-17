# 本番デプロイ Runbook

この runbook は、AnifusionCanvas を Cloud Run + Cloudflare Pages + TiDB + R2 + Replicate で本番運用するための操作手順です。Google Cloud、Cloudflare、TiDB、Replicate のアカウント操作と課金設定はユーザー本人だけが実行できます。

## 1. 確定している本番値

| 項目 | 値 |
| --- | --- |
| GCP project | `anifusioncanvas` |
| GCP region | `asia-northeast1` |
| Artifact Registry repository | `anifusion` |
| Cloud Run service | `anifusion-api` |
| Cloud Run URL | `https://anifusion-api-976317870900.asia-northeast1.run.app` |
| Cloudflare Pages production URL | `https://anifusion-canvas.pages.dev` |
| Frontend API env | `VITE_API_BASE_URL=https://anifusion-api-976317870900.asia-northeast1.run.app` |
| Backend CORS env | `FRONTEND_ORIGIN=https://anifusion-canvas.pages.dev` |

## 2. ユーザーが用意する Secret

Google Secret Manager に次の secret を作成します。secret 名は deploy script が参照しているため、この名前から変えないでください。

| Secret 名 | 取得元 | 用途 |
| --- | --- | --- |
| `DATABASE_URL` | TiDB Cloud | Studio state の永続化 |
| `R2_ACCOUNT_ID` | Cloudflare dashboard | R2 API 接続 |
| `R2_ACCESS_KEY_ID` | Cloudflare R2 API token | Object upload |
| `R2_SECRET_ACCESS_KEY` | Cloudflare R2 API token | Object upload |
| `R2_BUCKET` | Cloudflare R2 | 画像と MP4 の保存先 bucket |
| `R2_ENDPOINT_URL` | Cloudflare R2 | S3 compatible endpoint |
| `REPLICATE_API_TOKEN` | Replicate account | ToonCrafter / SDXL Inpainting |

作成例:

```bash
gcloud config set project anifusioncanvas

printf '%s' 'YOUR_DATABASE_DSN' | gcloud secrets create DATABASE_URL --data-file=-
printf '%s' 'YOUR_R2_ACCOUNT_ID' | gcloud secrets create R2_ACCOUNT_ID --data-file=-
printf '%s' 'YOUR_R2_ACCESS_KEY_ID' | gcloud secrets create R2_ACCESS_KEY_ID --data-file=-
printf '%s' 'YOUR_R2_SECRET_ACCESS_KEY' | gcloud secrets create R2_SECRET_ACCESS_KEY --data-file=-
printf '%s' 'YOUR_R2_BUCKET' | gcloud secrets create R2_BUCKET --data-file=-
printf '%s' 'YOUR_R2_ENDPOINT_URL' | gcloud secrets create R2_ENDPOINT_URL --data-file=-
printf '%s' 'YOUR_REPLICATE_API_TOKEN' | gcloud secrets create REPLICATE_API_TOKEN --data-file=-
```

更新例:

```bash
printf '%s' 'NEW_VALUE' | gcloud secrets versions add DATABASE_URL --data-file=-
```

## 3. Google Cloud 権限

GitHub Actions の deploy 用 service account と、Cloud Run runtime 用 service account を分けると運用しやすくなります。既存の `GCP_SERVICE_ACCOUNT_EMAIL` を runtime service account として使う場合も、最低限次を満たしてください。

Runtime service account:

```bash
export PROJECT_ID=anifusioncanvas
export RUNTIME_SA=YOUR_RUNTIME_SERVICE_ACCOUNT_EMAIL

for SECRET in DATABASE_URL R2_ACCOUNT_ID R2_ACCESS_KEY_ID R2_SECRET_ACCESS_KEY R2_BUCKET R2_ENDPOINT_URL REPLICATE_API_TOKEN; do
  gcloud secrets add-iam-policy-binding "$SECRET" \
    --member="serviceAccount:${RUNTIME_SA}" \
    --role="roles/secretmanager.secretAccessor"
done
```

GitHub Actions deploy service account:

```bash
export DEPLOY_SA=YOUR_DEPLOY_SERVICE_ACCOUNT_EMAIL

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${DEPLOY_SA}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${DEPLOY_SA}" \
  --role="roles/artifactregistry.writer"

gcloud iam service-accounts add-iam-policy-binding "$RUNTIME_SA" \
  --member="serviceAccount:${DEPLOY_SA}" \
  --role="roles/iam.serviceAccountUser"
```

GitHub Actions の Workload Identity Federation は GitHub 側 secret として次を設定します。

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_SERVICE_ACCOUNT_EMAIL`

## 4. Cloud Run 推奨設定

長時間の Replicate polling / download / FFmpeg export を前提に、deploy script は次の値を既定にしています。

| 設定 | 既定値 | 理由 |
| --- | --- | --- |
| CPU | `2` | FFmpeg と画像処理の余裕を持たせる |
| Memory | `2Gi` | base64 data URL と MP4 export の一時メモリを吸収する |
| Timeout | `900s` | Cloud Run 最大値。長時間ジョブの HTTP 起点処理を許容する |
| Concurrency | `4` | 重いジョブの同時実行を抑える |
| CPU throttling | disabled | response 後の background goroutine が止まりにくい状態にする |

必要に応じて deploy 実行時だけ上書きできます。

```bash
export CLOUD_RUN_CPU=2
export CLOUD_RUN_MEMORY=2Gi
export CLOUD_RUN_TIMEOUT=900s
export CLOUD_RUN_CONCURRENCY=4
```

## 5. R2 bucket 設定

R2 bucket では、Cloud Run からの S3 compatible API upload と、ブラウザから成果物を取得する公開 URL の扱いを分けて確認します。

必須:

- bucket 名が `R2_BUCKET` と一致している
- R2 API token が対象 bucket に read/write できる
- `R2_ENDPOINT_URL` が `https://<account-id>.r2.cloudflarestorage.com` 形式である

公開 URL を使う場合:

- R2 custom domain または public development URL を有効化する
- その base URL を `R2_PUBLIC_BASE_URL` として Cloud Run に設定する
- 未設定でも API は R2 object の upload 自体は実行できる

CORS を設定する場合の例:

```json
[
  {
    "AllowedOrigins": ["https://anifusion-canvas.pages.dev"],
    "AllowedMethods": ["GET", "HEAD"],
    "AllowedHeaders": ["*"],
    "ExposeHeaders": ["ETag"],
    "MaxAgeSeconds": 3600
  }
]
```

ライフサイクルはデモ用途なら 7 日から 30 日で古い生成物を削除する設定を推奨します。

## 6. デプロイ手順

Secret Manager 版を優先します。

```bash
gh workflow run "Deploy API to Cloud Run via Secret Manager"
```

ローカルから手動実行する場合:

```bash
export GCP_PROJECT_ID=anifusioncanvas
export GCP_REGION=asia-northeast1
export GCP_ARTIFACT_REPOSITORY=anifusion
export GCP_SERVICE_NAME=anifusion-api
export GCP_SERVICE_ACCOUNT_EMAIL=YOUR_RUNTIME_SERVICE_ACCOUNT_EMAIL

bash infra/cloud-run/deploy-secret-manager.sh
```

Cloudflare Pages は dashboard で次を設定してから production deploy を実行します。

```text
VITE_API_BASE_URL=https://anifusion-api-976317870900.asia-northeast1.run.app
```

## 7. 本番 smoke test

API:

```bash
export API_URL=https://anifusion-api-976317870900.asia-northeast1.run.app

curl -fsS "$API_URL/health"
curl -fsS "$API_URL/health/dependencies" | jq
```

期待:

- `/health` が `{"status":"ok"}` を返す
- `/health/dependencies` の `database`, `replicate`, `r2`, `ffmpeg` がすべて `ok`

CORS:

```bash
curl -i -X OPTIONS "$API_URL/projects" \
  -H 'Origin: https://anifusion-canvas.pages.dev' \
  -H 'Access-Control-Request-Method: POST'
```

期待:

- `access-control-allow-origin: https://anifusion-canvas.pages.dev` が返る

アプリ:

1. `https://anifusion-canvas.pages.dev` を開く
2. デモ画像を読み込む
3. frame generation を 1 回実行する
4. 生成 job が succeeded になることを確認する
5. inpaint を 1 回実行する
6. export video を 1 回実行する
7. 生成された画像または MP4 URL が開けることを確認する

実行結果は `docs/operations-guide.md` に日付、環境、結果、失敗時の対応を追記します。

## 8. Rollback

Cloud Run の revision rollback:

```bash
gcloud run revisions list \
  --service anifusion-api \
  --region asia-northeast1 \
  --project anifusioncanvas

gcloud run services update-traffic anifusion-api \
  --region asia-northeast1 \
  --project anifusioncanvas \
  --to-revisions REVISION_NAME=100
```

Cloudflare Pages は dashboard の Deployments から直前の成功 deployment を選び、Promote to production を実行します。

Secret の rollback は古い version を確認してから、必要に応じて新しい version として再投入します。

```bash
gcloud secrets versions list DATABASE_URL
gcloud secrets versions access VERSION_NUMBER --secret DATABASE_URL
printf '%s' 'ROLLBACK_VALUE' | gcloud secrets versions add DATABASE_URL --data-file=-
```

## 9. 残る運用判断

現行実装は Cloud Run 内の background goroutine で AI 生成と export job を進めます。`--no-cpu-throttling` と `timeout=900s` でデモ用途の安定性は上がりますが、instance shutdown、二重実行、長時間 retry は完全には解決しません。

本番運用品質まで上げる場合は、次のいずれかに移行します。

- Cloud Tasks で job 起動を明示的に retry する
- Pub/Sub + worker service に分ける
- Cloud Run Jobs で長時間処理を分離する

ポートフォリオ用途では、まず本 runbook の smoke test を固定し、失敗時に手動再実行できる状態にすることを優先します。
