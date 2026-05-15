# GitHub Actions で `unauthorized_client` が出る場合

次のエラーが出る場合の確認メモ。

```text
The given credential is rejected by the attribute condition.
```

## 原因

Workload Identity Provider の `attributeCondition` と、実際の GitHub repository が一致していない。

今回の実例:

- 誤: `haseg/AnifusionCanvas`
- 正: `Yuji181181/AnifusionCanvas`

## 確認方法

### GitHub remote を見る

```bash
git remote get-url origin
```

今回の結果:

```text
git@github.com:Yuji181181/AnifusionCanvas.git
```

### Provider の condition を見る

```bash
gcloud iam workload-identity-pools providers describe github \
  --project=anifusioncanvas \
  --location=global \
  --workload-identity-pool=github-actions-live
```

## 正しい condition

```text
attribute.repository=="Yuji181181/AnifusionCanvas"
```

## 修正コマンド

```bash
gcloud iam workload-identity-pools providers update-oidc github \
  --project=anifusioncanvas \
  --location=global \
  --workload-identity-pool=github-actions-live \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --attribute-condition='attribute.repository=="Yuji181181/AnifusionCanvas"'
```
