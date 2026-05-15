# Secret Manager セットアップ

このドキュメントでは、Cloud Run の環境変数を `.env` 直書きから外して、Google Secret Manager へ移す手順をまとめる。

## 必要性

本番運用では必要。

理由:
- 機密情報を GitHub Secrets や `.env` に長く置き続けないため
- Cloud Run の secret 連携で監査しやすくするため
- ローテーションしやすくするため

## 作成する Secret 名

- `DATABASE_URL`
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET`
- `R2_PUBLIC_BASE_URL`
- `R2_ENDPOINT_URL`
- `REPLICATE_API_TOKEN`

## 例: Secret 作成

```bash
echo -n 'YOUR_VALUE' | gcloud secrets create DATABASE_URL --data-file=-
```

すでに存在する場合は version を追加する。

```bash
echo -n 'YOUR_VALUE' | gcloud secrets versions add DATABASE_URL --data-file=-
```

## Cloud Run service account に権限付与

```bash
gcloud secrets add-iam-policy-binding DATABASE_URL \
  --member="serviceAccount:YOUR_SERVICE_ACCOUNT_EMAIL" \
  --role="roles/secretmanager.secretAccessor"
```

これを各 secret に対して実施する。
