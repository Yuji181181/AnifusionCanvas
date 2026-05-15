# 外部サービス側の詳細セットアップ手順

このドキュメントでは、**リポジトリ外であなたが実施する必要がある環境構築** を、できるだけ細かく順番にまとめる。

対象は次の5つ。

1. GitHub Secrets の登録
2. Google Cloud の Workload Identity Federation 設定
3. Cloud Run 用 service account と IAM 権限設定
4. Secret Manager への本番シークレット登録
5. Cloud Run の監視・運用設定

このドキュメントに沿って進めれば、GitHub Actions から **Cloud Run** と **Cloudflare Pages** へデプロイできる状態まで持っていける。

---

## 0. 事前に確定している情報

このプロジェクトでは、現時点で次を前提にしている。

- **GCP project ID**: `anifusioncanvas`
- **GCP region**: `asia-northeast1`
- **Artifact Registry repository**: `anifusion`
- **Cloud Run service name**: `anifusion-api`
- **Cloudflare Pages project name**: `anifusion-canvas`
- **Cloudflare account ID**: `fcd71b61f590ce4fc12a03c221e027cd`
- **Cloud Run URL**: `https://anifusion-api-976317870900.asia-northeast1.run.app`
- **Cloudflare Pages production domain**: `https://anifusion-canvas.pages.dev`

---

## 1. GitHub Secrets の登録

ここでは、GitHub Actions がデプロイ時に使う値を GitHub に登録する。

### 1-1. GitHub の画面を開く

1. GitHub で対象リポジトリを開く
2. 上部メニューの **Settings** を開く
3. 左メニューから **Secrets and variables** → **Actions** を開く
4. **Repository secrets** タブで **New repository secret** を押す

### 1-2. 登録する Secret 一覧

#### GCP 認証系

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_SERVICE_ACCOUNT_EMAIL`

#### Cloud Run 用アプリ設定

- `DATABASE_URL`
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET`
- `R2_PUBLIC_BASE_URL`
- `R2_ENDPOINT_URL`
- `REPLICATE_API_TOKEN`

#### Cloudflare Pages 用

- `CLOUDFLARE_API_TOKEN`

### 1-3. `.env` にある値と、`.env` にない値の違い

GitHub Secrets に入れる値のうち、次は **すでに `.env` にある値** なので、その値を使ってよい。

- `DATABASE_URL`
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET`
- `R2_PUBLIC_BASE_URL`
- `R2_ENDPOINT_URL`
- `REPLICATE_API_TOKEN`

一方で、次は **`.env` には入っていない値** なので、外部サービス上で取得または作成する必要がある。

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_SERVICE_ACCOUNT_EMAIL`
- `CLOUDFLARE_API_TOKEN`

これらは取得方法を知らないと埋められないので、以下で詳しく説明する。

### 1-4. `.env` にない値の取得方法

#### `GCP_WORKLOAD_IDENTITY_PROVIDER`

これは **Workload Identity Provider を作成したあとに GCP が発行する識別子**。
自分で適当に決める値ではない。

取得手順:

1. Google Cloud Console を開く
2. `anifusioncanvas` プロジェクトを選ぶ
3. **IAM & Admin** → **Workload Identity Federation** を開く
4. Pool と Provider を作成する
5. 作成後に表示される **Provider resource name** をコピーする

値の形式例:

```text
projects/976317870900/locations/global/workloadIdentityPools/github-actions-live/providers/github
```

これを GitHub Secret `GCP_WORKLOAD_IDENTITY_PROVIDER` に入れる。

#### `GCP_SERVICE_ACCOUNT_EMAIL`

これは **GitHub Actions が GCP 上でなりすます service account のメールアドレス**。
service account を作成したあとに確定する。

取得手順:

1. Google Cloud Console を開く
2. **IAM & Admin** → **Service Accounts** を開く
3. `anifusion-api-deployer` などの service account を作る
4. 作成後、一覧に表示される **メールアドレス** をコピーする

値の形式例:

```text
anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com
```

これを GitHub Secret `GCP_SERVICE_ACCOUNT_EMAIL` に入れる。

#### `CLOUDFLARE_API_TOKEN`

これは **Cloudflare Dashboard で手動発行する API Token**。
`.env` からは取得できない。

取得手順:

1. Cloudflare Dashboard にログインする
2. 右上プロフィール → **My Profile**
3. **API Tokens** を開く
4. **Create Token** を押す
5. Pages デプロイに必要な権限を持つ token を作る
6. 発行直後に表示される token をコピーする

これを GitHub Secret `CLOUDFLARE_API_TOKEN` に入れる。

### 1-5. GitHub へ入れるときの注意

### 1-3. 入力時の注意

- secret 名は **大文字・アンダースコア区切り** で正確に入力する
- 改行が混ざらないように注意する
- `DATABASE_URL` は **Go MySQL DSN形式** を使う
- `R2_PUBLIC_BASE_URL` を使わないなら空文字でもよいが、GitHub Secrets では空値管理しづらいので、未使用なら一旦ダミーや未設定運用でもよい

---

## 2. GCP Workload Identity Federation 設定

これは、GitHub Actions から GCP に安全にログインするための仕組み。

### 2-1. なぜ必要か

GitHub Actions で GCP 認証をする方法は大きく2つある。

- サービスアカウントキー JSON を GitHub Secrets に置く
- **Workload Identity Federation** を使う

このプロジェクトでは、後者を使う。理由は、**秘密鍵ファイルを GitHub に置かなくてよい**から。

### 2-2. GCP コンソールでやること

#### 手順

1. Google Cloud Console を開く
2. 対象プロジェクト `anifusioncanvas` を選ぶ
3. 左メニューから **IAM & Admin** → **Workload Identity Federation** を開く
4. **Create Pool** を押す

#### Pool 作成例

- Name: `github-actions-live`
- ID: `github-actions-live`

補足:

- `github-actions` という ID は、過去に作成して削除した pool が `DELETED` 状態で残っていると再利用できないことがある
- その場合は最初から `github-actions-live` のような新しい ID を使うほうが安全

作成後、次に **Provider** を作る。

5. 作成した Pool を開く
6. **Add Provider** を押す

#### Provider 作成例

- Provider type: **OpenID Connect (OIDC)**
- Provider name: `github`
- Issuer (URI): `https://token.actions.githubusercontent.com`

#### まず試すべき最小構成

もし **条件を何も入れていないのに同じエラーが出る** なら、Console 上の入力が崩れているか、mapping を増やしすぎている可能性がある。

その場合は、まず **最小構成** で作る。

Attribute mapping は **これだけ** にする。

- `google.subject` → `assertion.sub`
- `attribute.repository` → `assertion.repository`

Attribute condition は **空欄のまま** にする。

これで作成できるかをまず確認する。

#### 重要

- `attribute.actor`
- `attribute.ref`

は、最初は入れなくてよい。
まず Provider を作ることを優先する。

#### それでも同じエラーが出る場合

考えられるのは次のどちらか。

1. Console の入力欄に、見えない condition が残っている
2. Mapping の入力形式が間違っている

#### Console で見直すポイント

1. **Attribute condition** 欄が本当に空欄か確認する
   - スペースだけ入っていてもだめ
   - 以前入力した condition が UI 上に残っていないか確認する
2. **Attribute mapping** を全部消して入れ直す
3. 最小構成の2行だけにする

```text
google.subject=assertion.sub
attribute.repository=assertion.repository
```

4. 余計な空白や全角文字を入れない
5. Issuer URI が正確に次になっているか確認する

```text
https://token.actions.githubusercontent.com
```

#### 重要: Console では `key=value` 形式で入れる

GCP Console の mapping 入力欄では、画面によっては次のように **1行ずつ key=value** で入れる必要がある。

```text
google.subject=assertion.sub
attribute.repository=assertion.repository
```

ドキュメント上で `→` と書いていても、Console ではその記号は入力しない。

#### それでも解決しない場合の最短回避

Console を使わず、CLI で作成する。

ただし、**同じ名前の provider を作り直している途中** だったり、以前の試行で壊れた設定が残っていると、`attribute condition` を指定していないつもりでも同じエラーが出ることがある。

そのため、次の順でやるのが安全。

### 1. 既存 provider があるか確認する

```bash
gcloud iam workload-identity-pools providers list \
  --project="anifusioncanvas" \
  --location="global" \
  --workload-identity-pool="github-actions"
```

### 2. `github` provider が残っていたら削除する

```bash
gcloud iam workload-identity-pools providers delete github \
  --project="anifusioncanvas" \
  --location="global" \
  --workload-identity-pool="github-actions" \
  --quiet
```

### 3. そのままコピーできる 1行コマンドで再作成する

```bash
gcloud iam workload-identity-pools providers create-oidc github --project="anifusioncanvas" --location="global" --workload-identity-pool="github-actions" --display-name="GitHub provider" --issuer-uri="https://token.actions.githubusercontent.com" --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository"
```

複数行で実行したい場合は、次をそのままコピーする。

```bash
gcloud iam workload-identity-pools providers create-oidc github \
  --project="anifusioncanvas" \
  --location="global" \
  --workload-identity-pool="github-actions" \
  --display-name="GitHub provider" \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository"
```

### 4. それでも同じエラーが出る場合

Pool 自体を作り直す。

#### Pool 一覧を確認

```bash
gcloud iam workload-identity-pools list \
  --project="anifusioncanvas" \
  --location="global"
```

#### `github-actions` pool を削除

```bash
gcloud iam workload-identity-pools delete github-actions \
  --project="anifusioncanvas" \
  --location="global" \
  --quiet
```

#### Pool を再作成

```bash
gcloud iam workload-identity-pools create github-actions \
  --project="anifusioncanvas" \
  --location="global" \
  --display-name="GitHub Actions"
```

#### その後に provider を再作成

```bash
gcloud iam workload-identity-pools providers create-oidc github \
  --project="anifusioncanvas" \
  --location="global" \
  --workload-identity-pool="github-actions" \
  --display-name="GitHub provider" \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository"
```

#### 作成後に `GCP_WORKLOAD_IDENTITY_PROVIDER` を取得するコマンド

作成後は、次のコマンドをそのまま実行する。

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

出力された値を、そのまま GitHub Secret `GCP_WORKLOAD_IDENTITY_PROVIDER` に入れる。

#### Provider 作成後に condition を入れる場合

Provider が作れたあとで repository 制限をかけるなら、次のようにする。

```text
attribute.repository == "Yuji181181/AnifusionCanvas"
```

branch 制限まで入れる場合は、その時点で mapping に `attribute.ref=assertion.ref` を追加する。

### 2-3. 取得する値

Provider 作成後に **Provider resource name** が表示される。

これを GitHub Secret `GCP_WORKLOAD_IDENTITY_PROVIDER` に入れる。

形式例:

```text
projects/123456789/locations/global/workloadIdentityPools/github-actions/providers/github
```

#### どうやって取得するか

Console 上で Provider を開くと、詳細画面に resource name が表示される。
見つからない場合は CLI でも確認できる。

```bash
gcloud iam workload-identity-pools providers describe github \
  --project=anifusioncanvas \
  --location=global \
  --workload-identity-pool=github-actions
```

出力に含まれる `name:` の値が、そのまま `GCP_WORKLOAD_IDENTITY_PROVIDER` になる。

---

## 3. Cloud Run 用 service account と IAM 設定

GitHub Actions が Cloud Run デプロイや Artifact Registry push を行うためには、専用 service account を作って必要な権限を与える必要がある。

### 3-1. service account を作る

#### CLI で作る例

```bash
gcloud iam service-accounts create anifusion-api-deployer \
  --project anifusioncanvas \
  --display-name="Anifusion API Deployer"
```

#### 作成後のメールアドレス例

```text
anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com
```

これを GitHub Secret `GCP_SERVICE_ACCOUNT_EMAIL` に設定する。

### 3-2. 必要な IAM ロールを付与する

最低限、次のいずれかが必要。

#### 推奨最小寄り

- `roles/run.admin`
- `roles/artifactregistry.writer`
- `roles/iam.serviceAccountUser`

#### コマンド例

```bash
gcloud projects add-iam-policy-binding anifusioncanvas \
  --member="serviceAccount:anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding anifusioncanvas \
  --member="serviceAccount:anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding anifusioncanvas \
  --member="serviceAccount:anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"
```

### 3-3. Workload Identity からこの service account を impersonate できるようにする

GitHub OIDC principal に対して service account の利用権限を与える。

#### コマンド例

```bash
gcloud iam service-accounts add-iam-policy-binding \
  anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com \
  --project=anifusioncanvas \
  --role=roles/iam.workloadIdentityUser \
  --member="principalSet://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github-actions/attribute.repository/YOUR_GITHUB_OWNER/YOUR_REPOSITORY"
```

ここで必要なのは次。

- `PROJECT_NUMBER`
- GitHub owner / repository

### 3-4. `PROJECT_NUMBER` の確認方法

これは **project ID ではなく project number**。`anifusioncanvas` とは別物。

確認コマンド:

```bash
gcloud projects describe anifusioncanvas --format='value(projectNumber)'
```

出力例:

```text
976317870900
```

この数値を `PROJECT_NUMBER` に使う。

### 3-5. `YOUR_GITHUB_OWNER/YOUR_REPOSITORY` の確認方法

これは GitHub の URL に含まれている。

例:

```text
https://github.com/haseg/AnifusionCanvas
```

この場合:

- `YOUR_GITHUB_OWNER` = `haseg`
- `YOUR_REPOSITORY` = `AnifusionCanvas`

つまり、binding に使う値は次になる。

```text
haseg/AnifusionCanvas
```

完成形の例:

```bash
gcloud iam service-accounts add-iam-policy-binding \
  anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com \
  --project=anifusioncanvas \
  --role=roles/iam.workloadIdentityUser \
  --member="principalSet://iam.googleapis.com/projects/976317870900/locations/global/workloadIdentityPools/github-actions/attribute.repository/haseg/AnifusionCanvas"
```

---

## 4. Secret Manager への本番値登録

本番では、機密情報を `.env` や GitHub Secrets に長く置くより、**Secret Manager へ移す** ほうがよい。

### 4-1. 登録する secret 名

- `DATABASE_URL`
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET`
- `R2_PUBLIC_BASE_URL`
- `R2_ENDPOINT_URL`
- `REPLICATE_API_TOKEN`

### 4-2. secret を作る

#### 新規作成例

```bash
echo -n 'YOUR_VALUE' | gcloud secrets create DATABASE_URL \
  --project anifusioncanvas \
  --data-file=-
```

#### すでに存在する場合

```bash
echo -n 'YOUR_VALUE' | gcloud secrets versions add DATABASE_URL \
  --project anifusioncanvas \
  --data-file=-
```

### 4-3. Cloud Run service account に secret 読み取り権限を付与する

```bash
gcloud secrets add-iam-policy-binding DATABASE_URL \
  --project anifusioncanvas \
  --member="serviceAccount:anifusion-api-deployer@anifusioncanvas.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

これを各 secret に対して繰り返す。

### 4-4. どの workflow を使うか

- `.env` ベースなら: `deploy-api.yml`
- **Secret Manager ベースなら: `deploy-api-secret-manager.yml`**

本番では後者を推奨する。

---

## 5. Cloud Run の運用強化

PoC ではなく、継続利用や外部公開を見据えるならここも必要。

### 5-1. ログ確認

#### コンソール操作

1. Google Cloud Console を開く
2. **Cloud Run** → `anifusion-api` を開く
3. **Logs** タブを開く
4. デプロイログと runtime error がないか確認する

### 5-2. メトリクス確認

確認したいもの:

- Request count
- Error rate
- Latency
- Instance count
- Memory usage

#### コンソール操作

1. Cloud Run サービスを開く
2. **Metrics** タブを開く
3. 上記メトリクスを見る

### 5-3. アラート作成

最低限のアラート対象:

- 5xx エラー増加
- 高レイテンシ
- インスタンス起動失敗

#### コンソール操作

1. Google Cloud Console → **Monitoring**
2. **Alerting** → **Create Policy**
3. Cloud Run service を対象に条件を追加する

### 5-4. サービス設定の見直し

必要に応じて見直す項目:

- Memory
- CPU
- Timeout
- Max instances
- Concurrency

---

## 6. Cloudflare API Token の作成

GitHub Actions で Pages デプロイするには Cloudflare API Token が必要。

### 6-1. Cloudflare Dashboard を開く

1. Cloudflare にログイン
2. 右上プロフィール → **My Profile**
3. **API Tokens** を開く
4. **Create Token** を押す

### 6-2. 必要権限

Pages デプロイに必要な権限を含む token を作る。

最低限のイメージ:

- Account / Pages / Edit
- Account / Account Settings / Read

もしテンプレートを選ぶ画面が出る場合は、Pages に近い権限セットを選び、最終的に上の権限が含まれていることを確認する。

### 6-3. 発行直後にやること

Cloudflare の API Token は、**発行後にしか完全な値を見られない**ことが多い。
そのため、作成したらすぐにコピーする。

1. token を作成する
2. 表示された token をコピーする
3. すぐ GitHub の `CLOUDFLARE_API_TOKEN` secret に貼り付ける
4. ローカルのメモ帳などには残さない

### 6-4. GitHub Secret へ登録

作成した token を `CLOUDFLARE_API_TOKEN` として GitHub Secrets に登録する。

---

## 6.5 `.env` の値を GitHub Secrets へ転記する方法

`.env` にすでに入っている値は、次のようにそのまま GitHub Secret に転記できる。

対応表:

- `.env` の `DATABASE_URL` → GitHub Secret `DATABASE_URL`
- `.env` の `R2_ACCOUNT_ID` → GitHub Secret `R2_ACCOUNT_ID`
- `.env` の `R2_ACCESS_KEY_ID` → GitHub Secret `R2_ACCESS_KEY_ID`
- `.env` の `R2_SECRET_ACCESS_KEY` → GitHub Secret `R2_SECRET_ACCESS_KEY`
- `.env` の `R2_BUCKET` → GitHub Secret `R2_BUCKET`
- `.env` の `R2_PUBLIC_BASE_URL` → GitHub Secret `R2_PUBLIC_BASE_URL`
- `.env` の `R2_ENDPOINT_URL` → GitHub Secret `R2_ENDPOINT_URL`
- `.env` の `REPLICATE_API_TOKEN` → GitHub Secret `REPLICATE_API_TOKEN`

転記時のポイント:

- `KEY=value` の **`value` 部分だけ** をコピーする
- 余計な引用符を付けない
- 行末の空白を入れない
- 改行を含めない

例:

```env
REPLICATE_API_TOKEN=r8_xxxxxxxxxxxxxxxxxxxx
```

GitHub Secret に入れる値は次だけ。

```text
r8_xxxxxxxxxxxxxxxxxxxx
```

---

## 7. 最後に GitHub Actions を手動実行する

### 7-1. GitHub Actions タブを開く

1. GitHub リポジトリを開く
2. **Actions** タブを開く

### 7-2. 実行順

推奨順:

1. `CI`
2. まず `Deploy API to Cloud Run`
3. その後 `Deploy Web to Cloudflare Pages`
4. Secret Manager と service account 権限を整えたあとで `Deploy API to Cloud Run via Secret Manager`

### 7-3. 成功確認

- API workflow が成功する
- Pages workflow が成功する
- API `/health` が 200 になる
- Pages のURLが 200 で開ける

---

## 8. どこまでが私にできて、どこからがあなたの操作か

### 私ができること

- リポジトリ内の workflow / script / config 作成
- CLI での確認
- デプロイコマンド整備
- 必要な secret 名、IAM 権限、手順の設計

### あなたがやる必要があること

- GitHub Secrets の登録
- GCP Console 上の Workload Identity 作成
- IAM ロール付与
- Secret Manager への本番値登録
- Cloudflare API Token 発行
- Monitoring / Alerting の最終設定

---

## 9. このドキュメントを使った最短ルート

最短で進めるなら次の順。

1. GitHub Secrets を全部登録
2. Workload Identity Pool / Provider を作る
3. service account を作って権限を付ける
4. `GCP_WORKLOAD_IDENTITY_PROVIDER` と `GCP_SERVICE_ACCOUNT_EMAIL` を GitHub Secrets に登録
5. Secret Manager を使うなら secret を全部作る
6. GitHub Actions の `CI` を実行
7. GitHub Actions の `Deploy API to Cloud Run via Secret Manager` を実行
8. GitHub Actions の `Deploy Web to Cloudflare Pages` を実行

ここまで終わったら、環境構築としてはかなり完成度が高い状態です。
