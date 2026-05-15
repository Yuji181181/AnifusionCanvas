# CLI認証ガイド

このドキュメントでは、AnifusionCanvas のデプロイを **CLI ベース** で進めるために必要な、**Google Cloud CLI (`gcloud`)** と **Cloudflare Wrangler (`wrangler`)** の認証手順を、プロジェクト未作成の状態から順番に詳しくまとめる。

このドキュメントは、特に次の状態を前提にしている。

- Google Cloud プロジェクトをまだ作成していない
- `gcloud auth login` に失敗した
- WSL / Windows 環境でブラウザ連携が不安定
- まずは Cloud Run / Cloudflare Pages を CLI で触れる状態にしたい

---

## 0. 現在の状況

この環境で確認できている状態は次のとおり。

- `gcloud`: インストール済み、**未認証**
- `wrangler`: インストール済み、**未認証**
- `terraform`: **未導入**

そのため、まず必要なのは次の3つ。

1. **Google Cloud プロジェクトの作成**
2. **gcloud の認証と初期化**
3. **wrangler の認証**

---

## 1. まず理解しておくこと

### 1-1. `gcloud auth login` と `gcloud init` の違い

この2つは役割が違う。

- `gcloud auth login`
  - Google アカウントで **CLI を認証する** ためのコマンド
  - まだプロジェクトを作っていなくても実行できる
- `gcloud init`
  - 認証済みアカウントを使って、**デフォルトのプロジェクトやリージョンなどを初期設定する** コマンド
  - 認証後にやるとスムーズ

つまり、順番としては普通こうなる。

1. `gcloud auth login`
2. `gcloud init`
3. `gcloud config set project ...`

ただし、Windows / WSL では `gcloud auth login` のブラウザ起動がうまくいかないことがあるので、その場合は **`--no-launch-browser`** を使う。

---

## 2. Google Cloud プロジェクトを作成する

ここは **あなたが Google Cloud Console 上で操作する必要がある**。私はブラウザであなたのアカウントにログインできないため、ここは手動で進めてもらう必要がある。

### 2-1. Google Cloud Console を開く

ブラウザで Google Cloud Console を開く。

### 2-2. 新しいプロジェクトを作る

1. 画面上部のプロジェクト選択メニューを開く
2. **「新しいプロジェクト」** を選ぶ
3. 次の内容を入力する
   - **プロジェクト名**: わかりやすい名前
     - 例: `AnifusionCanvas`
   - **プロジェクトID**: グローバル一意のID
     - 例: `anifusion-canvas-xxxxx`
4. 作成する

### 2-3. 課金を有効化する

Cloud Run や Artifact Registry を使うには、通常は課金アカウントの紐付けが必要。

1. 左メニューから **Billing** を開く
2. 新規課金アカウントを作成するか、既存の課金アカウントを紐付ける
3. 対象プロジェクトに課金を有効化する

### 2-4. 必要な API を有効化する

次の API を有効化する。

- Cloud Run API
- Artifact Registry API
- Cloud Build API

#### 手順

1. 左メニューから **APIs & Services** → **Library** を開く
2. 上記 API を検索して順番に有効化する

---

## 3. `gcloud auth login` に失敗したときの正しい進め方

WSL 環境では、通常の `gcloud auth login` がブラウザ起動に失敗することがある。その場合は、**ブラウザ自動起動を使わない方法**で進める。

### 3-1. まず試すコマンド

```bash
gcloud auth login --no-launch-browser
```

### 3-2. 何が起きるか

このコマンドを実行すると、ターミナルに長いURLが表示される。

あなたがやることは次のとおり。

1. 表示されたURLをコピーする
2. 普段使っているブラウザで開く
3. Google アカウントでログインする
4. 許可画面で承認する
5. 表示された **認証コード** をターミナルに貼り付ける

### 3-3. 成功確認

```bash
gcloud auth list
```

期待する状態:

- Google アカウントが1つ以上表示される
- `ACTIVE` に `*` が付く

---

## 4. `gcloud init` を実行する

認証が通ったら、次に `gcloud init` を実行する。これは **やったほうがよい**。特にプロジェクト未設定の状態では、後のコマンドが安定する。

### 実行コマンド

```bash
gcloud init
```

### `gcloud init` でやること

対話形式でいくつか聞かれる。

#### 聞かれる内容の例

- どのアカウントを使うか
- どのプロジェクトをデフォルトにするか
- デフォルトリージョンやゾーンを設定するか

### 推奨

- アカウント: 今ログインしたものを選ぶ
- プロジェクト: 今作成した AnifusionCanvas 用プロジェクトを選ぶ
- リージョン/ゾーン: Compute Engine をすぐ使わないなら無理に設定しなくてもよい

### 確認コマンド

```bash
gcloud config list
```

特に次が確認できればよい。

- `core/account`
- `core/project`

---

## 5. プロジェクトを明示設定する

`gcloud init` 後でも、明示的に設定しておくと安全。

```bash
gcloud config set project YOUR_PROJECT_ID
```

### 確認

```bash
gcloud config list project
```

期待する表示例:

```text
[core]
project = anifusion-canvas-xxxxx
```

---

## 6. Artifact Registry 用 Docker 認証

Cloud Run にコンテナを載せるには、Docker イメージを Artifact Registry に push できる状態にする必要がある。

### 6-1. リージョンを決める

Cloud Run と Artifact Registry は同じリージョンが扱いやすい。候補としては次が無難。

- `asia-northeast1`（東京）

### 6-2. 認証コマンド

```bash
gcloud auth configure-docker REGION-docker.pkg.dev
```

例:

```bash
gcloud auth configure-docker asia-northeast1-docker.pkg.dev
```

### 6-3. 確認

```bash
cat ~/.docker/config.json
```

`credHelpers` に次のような設定が入っていればよい。

```json
{
  "credHelpers": {
    "asia-northeast1-docker.pkg.dev": "gcloud"
  }
}
```

---

## 7. Artifact Registry を作成する

ここは Console でも CLI でもできる。CLI でやるなら認証後に次を使える。

### 作成コマンド例

```bash
gcloud artifacts repositories create anifusion \
  --repository-format=docker \
  --location=asia-northeast1 \
  --description="AnifusionCanvas API images"
```

### 確認

```bash
gcloud artifacts repositories list --location=asia-northeast1
```

### ここで決める値

- Artifact Registry リポジトリ名
  - 例: `anifusion`
- リージョン
  - 例: `asia-northeast1`

---

## 8. `wrangler` の認証

Cloudflare Pages の設定確認や CLI 操作を進めるには `wrangler` 認証が必要。

### 8-1. 実行コマンド

```bash
wrangler login
```

### 8-2. あなたがやること

1. ブラウザが開く
2. Cloudflare にログインする
3. Wrangler へのアクセス許可を承認する

### 8-3. 成功確認

```bash
wrangler whoami
```

アカウント情報が出れば成功。

---

## 9. 失敗時の切り分け

### 9-1. `gcloud auth login` が失敗する

まずは次を試す。

```bash
gcloud auth login --no-launch-browser
```

それでも失敗する場合は次を確認する。

#### 確認項目

- Google Cloud SDK が最新か
- ブラウザで Google アカウントにログインできるか
- プロキシや VPN が邪魔していないか
- WSL から URL を開くのではなく、URL をコピーして Windows 側ブラウザで開いたか

### 9-2. `gcloud init` がうまくいかない

最小限なら、次の2つが通れば先へ進める。

```bash
gcloud auth list
gcloud config set project YOUR_PROJECT_ID
```

### 9-3. `wrangler login` が失敗する

確認ポイント:

- ブラウザで Cloudflare にログインできるか
- `wrangler whoami` が未認証のままか
- ログイン後にブラウザ側で承認を押したか

---

## 10. Terraform は今必要か

今は **必須ではない**。先に CLI で Cloud Run / Pages を動かすほうが順序として自然。

Terraform が必要になるのは次のようなタイミング。

- インフラをコード化して再現性を上げたい
- 複数環境（dev/staging/prod）を整理したい
- 何度も同じクラウド設定を作り直す可能性がある

今の段階では、まず CLI で一度動かすほうがよい。

---

## 11. 認証完了後に私へ伝えてほしいこと

次の情報がそろえば、私はそのまま CLI ベースで次のデプロイ作業に進める。

- `gcloud auth login` が完了したか
- `gcloud init` が完了したか
- `gcloud config set project ...` に設定した **project ID**
- 使用リージョン
  - 迷うなら `asia-northeast1`
- Artifact Registry リポジトリ名
  - 迷うなら `anifusion`
- `wrangler login` が完了したか
- Cloudflare Pages のプロジェクト名
  - 迷うなら `anifusion-canvas`

---

## 12. 認証後に私ができること

認証が終われば、私は次を CLI で進められる。

- Artifact Registry 作成
- Cloud Run デプロイ
- コンテナイメージの push
- Pages 用設定の確定
- Pages 用のビルド変数整理
- デプロイ手順の最終化

---

## 13. いちばん短い実行手順だけ抜き出すとこれ

### Google Cloud

```bash
gcloud auth login --no-launch-browser
gcloud init
gcloud config set project YOUR_PROJECT_ID
gcloud auth configure-docker asia-northeast1-docker.pkg.dev
```

### Cloudflare

```bash
wrangler login
wrangler whoami
```

---

## 14. 次にやるべきこと

次はこの順で進めるのがよい。

1. Google Cloud Console でプロジェクト作成
2. `gcloud auth login --no-launch-browser`
3. `gcloud init`
4. `gcloud config set project YOUR_PROJECT_ID`
5. `gcloud auth configure-docker asia-northeast1-docker.pkg.dev`
6. `wrangler login`

ここまで終わったら、私に **project ID / region / Artifact Registry 名 / Pages project 名** を伝えてください。そこから私が CLI で続けます。
