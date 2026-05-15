# Workload Identity Provider 作成エラー時の復旧手順

このドキュメントは、次のエラーが出る場合のための **実行順つき手順書**。

```text
INVALID_ARGUMENT: The attribute condition must reference one of the provider's claims.
```

このエラーは、Console 側で condition を入れていないつもりでも、**以前の作成失敗状態** や **既存 provider / pool の壊れた設定** が残っていると起きることがある。

---

## 手順 1: まず Pool が存在するか確認する

次をそのまま実行する。

```bash
gcloud iam workload-identity-pools list \
  --project="anifusioncanvas" \
  --location="global"
```

### 結果の見方

- `github-actions` が見つかったら、手順 1-1 へ進む
- `github-actions` が見つからなければ、手順 1-2 へ進む

### 手順 1-1: `github-actions` が DELETED 残留していないか確認する

次をそのまま実行する。

```bash
gcloud iam workload-identity-pools describe github-actions \
  --project="anifusioncanvas" \
  --location="global"
```

もし `state: DELETED` と出たら、その pool 名はしばらく再利用できないことがある。
その場合は `github-actions` を使わず、**新しい pool 名** を使う。

このプロジェクトでは次を推奨する。

- Pool ID: `github-actions-live`
- Display name: `GitHub Actions Live`

その場合は、以降の手順に出てくる `github-actions` を **すべて `github-actions-live` に読み替えて実行する**。

### 手順 1-2: Pool を新規作成する

次をそのまま実行する。

```bash
gcloud iam workload-identity-pools create github-actions \
  --project="anifusioncanvas" \
  --location="global" \
  --display-name="GitHub Actions"
```

作成できたら、手順 3 へ進む。

---

## 手順 2: 既存 provider を確認する

次をそのまま実行する。

```bash
gcloud iam workload-identity-pools providers list \
  --project="anifusioncanvas" \
  --location="global" \
  --workload-identity-pool="github-actions"
```

### 結果の見方

- 何も出なければ、手順 3 へ進む
- `github` など既存 provider が出たら、手順 2-2 へ進む

### 手順 2-2: 既存 provider を削除する

`github` provider が残っている場合は、次をそのまま実行する。

```bash
gcloud iam workload-identity-pools providers delete github \
  --project="anifusioncanvas" \
  --location="global" \
  --workload-identity-pool="github-actions" \
  --quiet
```

削除後、もう一度 provider 一覧を確認する。

```bash
gcloud iam workload-identity-pools providers list \
  --project="anifusioncanvas" \
  --location="global" \
  --workload-identity-pool="github-actions"
```

一覧が空になったら、手順 3 へ進む。

---

## 手順 3: Provider を CLI で作成する

次を **1行のまま** そのまま実行する。

```bash
gcloud iam workload-identity-pools providers create-oidc github --project="anifusioncanvas" --location="global" --workload-identity-pool="github-actions-live" --display-name="GitHub provider" --issuer-uri="https://token.actions.githubusercontent.com" --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" --attribute-condition="attribute.repository==\"Yuji181181/AnifusionCanvas\""
```

### 成功したら

そのまま手順 6 へ進む。

### まだ同じエラーが出たら

手順 4 へ進む。

---

## 手順 4: Workload Identity Pool を確認する

次をそのまま実行する。

```bash
gcloud iam workload-identity-pools list \
  --project="anifusioncanvas" \
  --location="global"
```

### 結果の見方

- `github-actions` がある → 手順 5 へ進む
- `github-actions` がない → 手順 5 の「Pool を作成する」だけ実行する

---

## 手順 5: Pool を削除して作り直す

### 5-1. Pool を削除する

```bash
gcloud iam workload-identity-pools delete github-actions \
  --project="anifusioncanvas" \
  --location="global" \
  --quiet
```

### 5-2. Pool を作成する

```bash
gcloud iam workload-identity-pools create github-actions \
  --project="anifusioncanvas" \
  --location="global" \
  --display-name="GitHub Actions"
```

### 5-3. Provider を再作成する

```bash
gcloud iam workload-identity-pools providers create-oidc github --project="anifusioncanvas" --location="global" --workload-identity-pool="github-actions-live" --display-name="GitHub provider" --issuer-uri="https://token.actions.githubusercontent.com" --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" --attribute-condition="attribute.repository==\"Yuji181181/AnifusionCanvas\""
```

---

## 手順 6: `GCP_WORKLOAD_IDENTITY_PROVIDER` を取得する

Provider 作成後、次をそのまま実行する。

```bash
gcloud iam workload-identity-pools providers describe github \
  --project="anifusioncanvas" \
  --location="global" \
  --workload-identity-pool="github-actions-live" \
  --format="value(name)"
```

今回実際に取得できた値:

```text
projects/976317870900/locations/global/workloadIdentityPools/github-actions-live/providers/github
```

出力例:

```text
projects/976317870900/locations/global/workloadIdentityPools/github-actions/providers/github
```

この出力を、そのまま GitHub Secret `GCP_WORKLOAD_IDENTITY_PROVIDER` に登録する。

---

## 手順 7: まだダメな場合に確認すること

### 7-1. issuer URI が正しいか

正しい値:

```text
https://token.actions.githubusercontent.com
```

### 7-2. mapping が最小構成か

正しい値:

```text
google.subject=assertion.sub,attribute.repository=assertion.repository
```

### 7-3. Console ではなく CLI を使っているか

Console 入力が壊れている場合があるので、まずはこのドキュメントの CLI 手順を優先する。

---

## 手順 8: 作成後に次にやること

Provider が作れたら、次に必要なのはこれ。

1. service account を作る
2. `GCP_SERVICE_ACCOUNT_EMAIL` を取得する
3. Workload Identity User の binding を設定する
4. GitHub Secrets に両方登録する

この続きは `docs/external-console-setup-guide.md` に従う。
