# 残りの環境構築と必要性

## 1. GitHub Actions / Secrets 本番化

必要: 高

理由:
- 再デプロイを安定化するため
- 手元依存を減らすため
- 誰が見ても再現できるようにするため

## 2. CI の build / lint / test

必要: 高

理由:
- 壊れた変更を main に入れにくくするため
- 実装フェーズへ進む前の土台になるため

## 3. Docker 認証の恒久運用

必要: 高

理由:
- 今は一時対応の login で運用しているため
- ローカル差異を減らしたいため

## 4. Secret Manager 化

必要: 中〜高

理由:
- 本番ではほぼ必須
- ただし PoC 段階なら後回しも可能

## 5. Cloud Run 運用強化

必要: 中

理由:
- すぐ必須ではないが、本番公開なら必要
- 監視、権限、運用性を上げるため

## 今回追加したもの

- `.github/workflows/ci.yml`
- `.github/workflows/deploy-api-secret-manager.yml`
- `infra/cloud-run/deploy-secret-manager.sh`
- `docs/secret-manager-setup.md`
- `docs/cloud-run-ops.md`
