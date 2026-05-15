# Cloud Run 運用強化メモ

## 必要性

本格運用前には必要。
PoC の段階では最低限でも動くが、継続運用するなら早めに整えたほうがよい。

## 推奨項目

- 専用 service account を作る
- Secret Manager 連携へ移行する
- Cloud Logging / Error Reporting を確認する
- 最低限の監視アラートを設定する
- max instances / timeout / memory を見直す

## service account 例

```bash
gcloud iam service-accounts create anifusion-api \
  --display-name="Anifusion API"
```

## 推奨ロール

- `roles/run.admin` または必要最小限の Cloud Run 更新権限
- `roles/artifactregistry.writer`
- `roles/secretmanager.secretAccessor`
- 必要に応じて `roles/logging.logWriter`

## 監視で見るもの

- 5xx エラー率
- レイテンシ
- インスタンス起動回数
- メモリ使用量
- デプロイ失敗履歴
