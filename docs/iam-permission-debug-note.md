# IAM 権限エラー時の確認メモ

## 今回のエラー

```text
Permission 'iam.serviceAccounts.getAccessToken' denied
```

## 原因

Workload Identity User binding が未設定だった。

GitHub Actions が Workload Identity Federation で認証しても、service account を偽装（impersonate）する権限がないと token 取得に失敗する。

## 修正内容

次の binding を追加した。

```bash
gcloud iam service-accounts add-iam-policy-binding \
  anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com \
  --project=anifusioncanvas \
  --role='roles/iam.workloadIdentityUser' \
  --member='principalSet://iam.googleapis.com/projects/976317870900/locations/global/workloadIdentityPools/github-actions-live/attribute.repository/Yuji181181/AnifusionCanvas'
```

## 確認方法

```bash
gcloud iam service-accounts get-iam-policy \
  anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com \
  --project=anifusioncanvas
```

`roles/iam.workloadIdentityUser` に `principalSet://.../attribute.repository/Yuji181181/AnifusionCanvas` が含まれていれば OK。

## service account に必要なロール

| ロール | 用途 |
|---|---|
| `roles/run.admin` | Cloud Run デプロイ |
| `roles/artifactregistry.writer` | Docker image push |
| `roles/iam.serviceAccountUser` | service account 偽装 |
| `roles/iam.workloadIdentityUser` | Workload Identity からの偽装 |

これらは `anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com` に付与済み。