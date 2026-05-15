# Docker 認証メモ

この環境では `docker-credential-gcloud` helper が使えず、Artifact Registry への push で認証エラーが出た。

そのため、現時点では次の方式を採用している。

```bash
gcloud auth print-access-token | docker login -u oauth2accesstoken --password-stdin https://asia-northeast1-docker.pkg.dev
```

`infra/cloud-run/deploy.sh` は、この login を内部で実行してから push する。

## 補足

- Docker 設定ファイルの helper は一時的に外した
- そのため、`~/.docker/config.json` の資格情報は平文保存になる場合がある
- 長期的には、helper の復旧か CI/CD への移行でこのローカル依存を減らすのがよい
