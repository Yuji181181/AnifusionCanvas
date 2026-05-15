# Artifact Registry push 認証トラブルシュート

Cloud Run デプロイ時に次のようなエラーが出る場合の対処をまとめる。

```text
Unauthenticated request. Unauthenticated requests do not have permission "artifactregistry.repositories.uploadArtifacts"
```

## 原因

このエラーは、次のどちらかで起きることが多い。

1. `gcloud auth configure-docker` は済んでいるが、Docker Desktop / WSL の認証連携がずれている
2. Docker が `gcloud` credential helper を正しく呼べていない

## 事前確認

### 1. gcloud 認証

```bash
gcloud auth list
```

### 2. Docker config

```bash
cat ~/.docker/config.json
```

期待する例:

```json
{
  "credsStore": "desktop.exe",
  "credHelpers": {
    "asia-northeast1-docker.pkg.dev": "gcloud"
  }
}
```

### 3. Artifact Registry の存在確認

```bash
gcloud artifacts repositories describe anifusion \
  --location=asia-northeast1 \
  --project=anifusioncanvas
```

## 最も確実な回避方法

credential helper が効かない場合は、明示的に Docker login する。

```bash
gcloud auth print-access-token | docker login -u oauth2accesstoken --password-stdin https://asia-northeast1-docker.pkg.dev
```

成功すると次のような表示になる。

```text
Login Succeeded
```

## その後にやること

```bash
export GCP_PROJECT_ID=anifusioncanvas
export GCP_REGION=asia-northeast1
export GCP_ARTIFACT_REPOSITORY=anifusion
export GCP_SERVICE_NAME=anifusion-api

bash infra/cloud-run/deploy.sh
```

## まだ失敗する場合

次を確認する。

- Google Cloud プロジェクトに課金が有効か
- 自分のアカウントに Artifact Registry へ push する権限があるか
- push 先の `project / region / repository` が一致しているか
- Docker Desktop を再起動したか
- `gcloud auth print-access-token` が実行できるか
