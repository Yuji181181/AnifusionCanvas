# 環境変数セットアップガイド

このドキュメントでは、AnifusionCanvas の開発・本番運用に必要な環境変数の意味、設定値の入手方法、設定時の注意点をまとめる。

## 現在の前提

- フロントエンド: Vite+ / React SPA
- フロントエンド配信: Cloudflare Pages
- バックエンド: Go + Echo
- バックエンド実行: Google Cloud Run
- DB: TiDB Serverless
- オブジェクトストレージ: Cloudflare R2
- 外部AI API: Replicate

## 必要な環境変数一覧

### 共通

| 変数名 | 必須 | 用途 |
|---|---|---|
| `APP_ENV` | 任意 | 実行環境名。`development` / `staging` / `production` など |

### フロントエンド

| 変数名 | 必須 | 用途 |
|---|---|---|
| `VITE_API_BASE_URL` | 必須 | フロントエンドから接続するGo APIのベースURL |

### バックエンド

| 変数名 | 必須 | 用途 |
|---|---|---|
| `API_PORT` | 任意 | Go APIの待受ポート。ローカルでは `8080` を推奨 |
| `FRONTEND_ORIGIN` | 必須 | CORS許可対象のフロントエンドURL |
| `DATABASE_URL` | 必須 | Go MySQLドライバ互換のTiDB接続文字列 |
| `R2_ACCOUNT_ID` | 必須 | Cloudflare R2アカウント識別子 |
| `R2_ACCESS_KEY_ID` | 必須 | Cloudflare R2 APIアクセスキーID |
| `R2_SECRET_ACCESS_KEY` | 必須 | Cloudflare R2 APIシークレット |
| `R2_BUCKET` | 必須 | 利用するR2バケット名 |
| `R2_PUBLIC_BASE_URL` | 任意 | 公開配信する場合のベースURL |
| `R2_ENDPOINT_URL` | 必須 | R2のS3互換エンドポイント |
| `R2_REGION` | 任意 | 通常は `auto` |
| `REPLICATE_API_TOKEN` | 必須 | Replicate APIトークン |

## すでに埋めた値

`.env` には以下を追記済み。

```dotenv
APP_ENV=development
API_PORT=8080
FRONTEND_ORIGIN=http://localhost:3000
VITE_API_BASE_URL=http://localhost:8080
```

これらはローカル開発用の妥当なデフォルト値としてそのまま使える。

## あなたにしか取得できない、または確認が必要な値

以下は私が外部サービスへログインして取得できないため、あなたが確認・設定する必要がある。

- `REPLICATE_API_TOKEN`
- `R2_PUBLIC_BASE_URL`（公開URL運用する場合のみ）
- 本番用 `FRONTEND_ORIGIN`
- 本番用 `VITE_API_BASE_URL`
- **Go用に変換した `DATABASE_URL`**

---

## 1. Replicate API トークンの取得方法

### 手順

1. Replicate にログインする
2. 右上のプロフィールメニューから API Tokens 画面へ移動する
3. 新しいトークンを作成する
4. 発行されたトークンを `REPLICATE_API_TOKEN` に設定する

### 設定例

```dotenv
REPLICATE_API_TOKEN=r8_xxxxxxxxxxxxxxxxxxxxx
```

### 注意

- トークンは再表示できない場合があるので、発行直後に安全な場所へ保管する
- Git にコミットしない
- Cloud Run 本番では Secret Manager か環境変数管理で注入する

---

## 2. Cloudflare Pages 用 URL の確認方法

### `FRONTEND_ORIGIN` の本番値

Cloudflare Pages にデプロイしたときのフロントエンドURLを設定する。

### 手順

1. Cloudflare Dashboard にログインする
2. Workers & Pages から対象の Pages プロジェクトを開く
3. `*.pages.dev` のURL、または独自ドメインURLを確認する
4. そのURLを `FRONTEND_ORIGIN` に設定する

### 設定例

```dotenv
FRONTEND_ORIGIN=https://anifusion-canvas.pages.dev
```

---

## 3. フロントエンド用 API URL の確認方法

### `VITE_API_BASE_URL` の本番値

Cloud Run にデプロイしたAPIの公開URLを設定する。

### 手順

1. Google Cloud Console にログインする
2. Cloud Run から対象サービスを開く
3. サービスURLを確認する
4. そのURLを `VITE_API_BASE_URL` に設定する

### 設定例

```dotenv
VITE_API_BASE_URL=https://anifusion-api-xxxxx-an.a.run.app
```

### 注意

- フロントエンドに露出してよいのは公開APIのURLのみ
- 秘密鍵やシークレットを `VITE_` プレフィックス付きで置かない

---

## 4. Cloudflare R2 公開URLの取得方法

### `R2_PUBLIC_BASE_URL` が必要なケース

生成物を署名付きURLではなく公開URLで配信する場合に使う。

### 取得方法

1. Cloudflare Dashboard にログインする
2. R2 の対象バケットを開く
3. Public bucket または Custom domain の設定を確認する
4. 公開URLまたはカスタムドメインURLを `R2_PUBLIC_BASE_URL` に設定する

### 設定例

```dotenv
R2_PUBLIC_BASE_URL=https://pub-xxxxxxxx.r2.dev
```

または

```dotenv
R2_PUBLIC_BASE_URL=https://assets.example.com
```

### 空のままでよいケース

- すべてバックエンド経由の配信にする
- 署名付きURLのみを使う

---

## 5. TiDB の `DATABASE_URL` を Go 用に変換する方法

今の `.env` には Python 向けの接続文字列が入っている。

```dotenv
DATABASE_URL=mysql+pymysql://USER:PASSWORD@HOST:PORT/DB_NAME?ssl_verify_cert=true
```

これは **Go の `go-sql-driver/mysql` ではそのまま使えない**。Go では次のような MySQL DSN 形式に変換する必要がある。

```dotenv
DATABASE_URL=USER:PASSWORD@tcp(HOST:PORT)/DB_NAME?tls=true
```

### あなたの `.env` にある値を使った変換の考え方

元の形式:

```text
mysql+pymysql://USER:PASSWORD@HOST:PORT/DB_NAME?ssl_verify_cert=true
```

Go用の形式:

```text
USER:PASSWORD@tcp(HOST:PORT)/DB_NAME?tls=true
```

### 変換手順

1. `mysql+pymysql://` を削除する
2. `@HOST:PORT` の部分を `@tcp(HOST:PORT)` に置き換える
3. DB名はそのまま引き継ぐ
4. `?ssl_verify_cert=true` は、まずは `?tls=true` に置き換える
5. 接続できない場合は TiDB 側の TLS 要件に応じて追加調整する

### 変換例

```dotenv
DATABASE_URL=USER:PASSWORD@tcp(HOST:4000)/test?tls=true
```

### 注意

- パスワードに特殊文字が含まれる場合はURLエンコードやDSNの解釈差異に注意する
- TiDB Cloud の推奨接続パラメータがある場合はそちらを優先する
- 必要なら将来的に `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` の分割管理へ変更したほうが安全

---

## 6. ローカル開発用の推奨値

```dotenv
APP_ENV=development
API_PORT=8080
FRONTEND_ORIGIN=http://localhost:3000
VITE_API_BASE_URL=http://localhost:8080
R2_REGION=auto
```

---

## 7. 本番環境での推奨運用

- `.env` をそのまま本番へ持ち込まない
- Cloud Run には Secret Manager または環境変数設定を使う
- Cloudflare Pages には公開して問題ない変数のみ設定する
- `VITE_` 付き変数には秘密情報を絶対に入れない

---

## 8. 次にあなたがやること

1. `REPLICATE_API_TOKEN` を埋める
2. `DATABASE_URL` を Go 用DSNに変換する
3. 公開配信するなら `R2_PUBLIC_BASE_URL` を設定する
4. 本番デプロイ後に `FRONTEND_ORIGIN` と `VITE_API_BASE_URL` を本番URLへ差し替える
