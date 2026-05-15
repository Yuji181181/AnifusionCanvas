# デプロイ設定ガイド

このドキュメントでは、AnifusionCanvas を **Cloud Run（API）** と **Cloudflare Pages（フロントエンド）** にデプロイするために必要な設定と、あなたが手作業で行うべき操作を順番にまとめる。

## 1. 全体構成

- フロントエンド: Cloudflare Pages
- バックエンド: Google Cloud Run
- コンテナイメージ: Artifact Registry
- 環境変数管理: Cloud Run の環境変数、または必要に応じて Secret Manager

---

## 2. 先に私が用意したもの

以下のファイルはすでに作成済み。

- `infra/cloud-run/service.yaml`
- `infra/cloud-run/env.production.yaml`
- `infra/cloud-run/deploy.sh`
- `infra/cloudflare-pages/wrangler.toml`
- `infra/cloudflare-pages/README.md`

あなたはこれらの `REPLACE_ME` を実際の値に置き換えるか、Cloud コンソール上で同等の設定を行えばよい。

---

## 3. あなたがやる必要があること

これは私では実行できない。理由は、あなたの Google Cloud / Cloudflare アカウントへのログイン、課金設定、プロジェクト作成、権限付与が必要だから。

### 3-1. Google Cloud 側

#### 目的

Cloud Run に API を載せるためのプロジェクトと Artifact Registry を準備する。

#### 手順

1. Google Cloud Console にログインする
2. 新しいプロジェクトを作成する、または既存プロジェクトを選ぶ
3. 課金を有効化する
4. 以下の API を有効化する
   - Cloud Run API
   - Artifact Registry API
   - Cloud Build API
5. Artifact Registry に Docker リポジトリを作成する
   - 形式: Docker
   - リージョン: Cloud Run と同じリージョン
   - 名前例: `anifusion`
6. ローカルで `gcloud auth login` を行う
7. ローカルで対象プロジェクトを設定する

```bash
gcloud config set project YOUR_PROJECT_ID
```

8. Docker が Artifact Registry に push できるよう認証する

```bash
gcloud auth configure-docker REGION-docker.pkg.dev
```

#### あなたが決める値

- `GCP_PROJECT_ID`
- `GCP_REGION`
- `GCP_ARTIFACT_REPOSITORY`
- `GCP_SERVICE_NAME`

---

### 3-2. Cloud Run の環境変数設定

#### 目的

API が本番で利用する認証情報や接続先を設定する。

#### 設定対象

`infra/cloud-run/env.production.yaml` の以下を実値に置き換える。

- `FRONTEND_ORIGIN`
- `DATABASE_URL`
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET`
- `R2_PUBLIC_BASE_URL`
- `R2_ENDPOINT_URL`
- `REPLICATE_API_TOKEN`

#### 注意

- `DATABASE_URL` は **Go MySQL ドライバ互換のDSN** を使う
- 機密情報は将来的には Secret Manager に寄せたほうがよい
- 最初は env file 運用で問題ない

---

### 3-3. Cloudflare Pages 側

#### 目的

Vite SPA を静的配信する。

#### 手順

1. Cloudflare Dashboard にログインする
2. Workers & Pages を開く
3. 新規 Pages プロジェクトを作成する
4. GitHub リポジトリを接続する
5. Build 設定を入力する
   - Build command: `bun install && bun run build:web`
   - Build output directory: `apps/web/dist`
   - Root directory: `/`
6. 環境変数 `VITE_API_BASE_URL` を設定する
   - 値: Cloud Run の公開URL
7. デプロイを実行する

#### あなたが決める値

- Pages の公開URL
- `VITE_API_BASE_URL`

補足:

- Cloudflare Pages プロジェクト名は `anifusion-canvas` で作成済み

---

## 4. デプロイ順序

1. Google Cloud プロジェクト準備
2. Artifact Registry 作成
3. `env.production.yaml` の実値設定
4. Cloud Run API デプロイ
5. Cloud Run URL の確認
6. Cloudflare Pages に `VITE_API_BASE_URL` を設定
7. Cloudflare Pages デプロイ
8. Cloud Run 側の `FRONTEND_ORIGIN` を Pages URL に更新
9. 最終疎通確認

---

## 5. Cloud Run のデプロイ方法

### 先に確認すること

Artifact Registry への push で認証エラーが出る場合は、先に次を実行する。

```bash
gcloud auth configure-docker asia-northeast1-docker.pkg.dev
cat ~/.docker/config.json
```

`credHelpers` に `asia-northeast1-docker.pkg.dev: gcloud` が入っていることを確認する。

それでも `Unauthenticated request` が出る場合は、Docker Desktop / WSL 側で参照している Docker 設定がずれている可能性がある。

その場合は、credential helper に頼らず、一度だけ明示的に login する。

```bash
gcloud auth print-access-token | docker login -u oauth2accesstoken --password-stdin https://asia-northeast1-docker.pkg.dev
```

これで `Login Succeeded` が出たら、再度 `bash infra/cloud-run/deploy.sh` を実行する。


### 事前に必要な環境変数

```bash
export GCP_PROJECT_ID=YOUR_PROJECT_ID
export GCP_REGION=YOUR_REGION
export GCP_ARTIFACT_REPOSITORY=YOUR_REPOSITORY
export GCP_SERVICE_NAME=anifusion-api
```

### 実行コマンド

```bash
bash infra/cloud-run/deploy.sh
```

### デプロイ後に確認すること

```bash
curl https://YOUR_CLOUD_RUN_URL/health
```

期待レスポンス:

```json
{"status":"ok"}
```

---

## 6. Cloudflare Pages の設定値

`VITE_API_BASE_URL` には Cloud Run の公開URLを入れる。

例:

```text
https://anifusion-api-xxxxx-an.a.run.app
```

---

## 7. 本番公開前に確認すること

- Cloud Run の `FRONTEND_ORIGIN` が Pages のURLになっている
- Pages の `VITE_API_BASE_URL` が Cloud Run URL になっている
- `R2_PUBLIC_BASE_URL` を使う場合は公開URLが正しい
- API の `/health` が通る
- フロントから API への CORS エラーが出ない

---

## 8. 現在この環境で確認できた情報

以下はすでに CLI から確認できている。

- GCP project ID: `anifusioncanvas`
- GCP region: `asia-northeast1`
- Artifact Registry repository: `anifusion`
- Cloud Run API: 有効
- Artifact Registry API: 有効
- Cloud Build API: 有効
- Cloudflare account ID: `fcd71b61f590ce4fc12a03c221e027cd`

現時点で確定したURLは次。

- Cloud Run: `https://anifusion-api-976317870900.asia-northeast1.run.app`
- Cloudflare Pages preview deployment: `https://197449f8.anifusion-canvas.pages.dev`
- Cloudflare Pages production domain: `https://anifusion-canvas.pages.dev`

引き続き未確定なのは次。

- `infra/cloud-run/env.production.yaml` に入れる本番値の恒久管理方法

## 9. 私が次にできること

あなたが以下を終えたら、次は私が続けられる。

- `env.production.yaml` の実値反映
- Cloudflare Pages の project 名の確定
- 本番URLの確定

その後は、私がデプロイ用設定の最終調整や、必要なら GitHub Actions 化まで進められる。
