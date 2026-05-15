# デプロイチェックリスト

## Cloud Run

- [x] Google Cloud プロジェクト作成済み
- [ ] 課金有効化済み
- [x] Cloud Run API 有効化済み
- [x] Artifact Registry API 有効化済み
- [x] Cloud Build API 有効化済み
- [x] Artifact Registry 作成済み
- [ ] `infra/cloud-run/env.production.yaml` の `REPLACE_ME` を置換済み
- [ ] `gcloud auth configure-docker asia-northeast1-docker.pkg.dev` 実行済み
- [ ] `bash infra/cloud-run/deploy.sh` 実行済み
- [ ] `/health` 確認済み

## Cloudflare Pages

- [ ] Pages プロジェクト作成済み
- [ ] GitHub 連携済み
- [ ] Build command 設定済み
- [ ] Output directory 設定済み
- [ ] `VITE_API_BASE_URL` 設定済み
- [x] デプロイ成功済み

## 現在の補足

- [x] Cloudflare Pages project 名は `anifusion-canvas`
- [ ] Cloud Run URL は未確定
- [ ] Pages の初回デプロイは未実施

## Cross Check

- [ ] `FRONTEND_ORIGIN` が Pages URL になっている
- [ ] CORS エラーが出ない
- [ ] API にフロントから到達できる
