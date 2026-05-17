# 動作確認手順書

このドキュメントでは AnifusionCanvas のローカルおよび本番環境での動作確認手順を説明する。

## 前提条件

### ローカル開発環境

| ツール | 最低バージョン | 確認コマンド |
|--------|------------|------------|
| Go | 1.22+ | `go version` |
| Bun | 1.0+ | `bun --version` |
| FFmpeg | 4.0+ | `ffmpeg -version` |
| Git | 2.0+ | `git --version` |

### 本番環境 (Cloud Run + Cloudflare Pages)

| サービス | 説明 |
|----------|------|
| Google Cloud Run | API サーバーの実行環境 |
| Cloudflare Pages | フロントエンドの配信 |
| TiDB Serverless | データベース |
| Cloudflare R2 | 画像・動画ストレージ |
| Replicate | AI 推論サービス |

## 1. リポジトリのセットアップ

```bash
git clone git@github.com:Yuji181181/AnifusionCanvas.git
cd AnifusionCanvas
bun install
cd apps/api && go mod tidy && cd ../..
```

## 2. 環境変数の設定

ルートディレクトリに `.env` ファイルを作成する。

### デモモード用（最小構成）

```bash
# .env
APP_ENV=development
API_PORT=8080
FRONTEND_ORIGIN=http://localhost:3000
VITE_API_BASE_URL=http://localhost:8080
STUDIO_STORE=memory
```

この設定では AI 推論はダミー画像（SVG）で動作し、データはメモリ上に保持される。

### 実 AI 推論モード用

```bash
# .env (デモモードの設定に加えて)
REPLICATE_API_TOKEN=r8_xxxxxx
REPLICATE_TOONCRAFTER_VERSION=fofr/tooncrafter
REPLICATE_SDXL_INPAINTING_VERSION=lucataco/sdxl-inpainting
```

### 画像保存 + データベースモード用

```bash
# .env (上記に加えて)
STUDIO_STORE=database
DATABASE_URL=mysql://user:password@host:4000/anifusion?tls=true

R2_ACCOUNT_ID=xxxxxx
R2_ACCESS_KEY_ID=xxxxxx
R2_SECRET_ACCESS_KEY=xxxxxx
R2_BUCKET=anifusion-canvas
R2_ENDPOINT_URL=https://xxxxxx.r2.cloudflarestorage.com
R2_PUBLIC_BASE_URL=https://assets.example.test
R2_REGION=auto
```

## 3. API サーバーの起動

```bash
cd apps/api
go run ./cmd/server
```

起動成功時の出力:

```
   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v4.12.0
High performance, minimalist Go web framework
https://echo.labstack.com
____________________________________O/_______
                                    O\
⇨ http server started on [::]:8080
```

## 4. フロントエンドの起動

別のターミナルで:

```bash
cd apps/web
bun run dev
```

起動成功時の出力:

```
  VITE v5.4.21  ready in xxx ms

  ➜  Local:   http://localhost:3000/
  ➜  Network: use --host to expose
```

## 5. ヘルスチェック

### API ヘルスチェック

```bash
curl http://localhost:8080/health
```

期待されるレスポンス:

```json
{"status":"ok"}
```

### 依存サービスのヘルスチェック

```bash
curl http://localhost:8080/health/dependencies
```

期待されるレスポンス（デモモード）:

```json
{
  "status": "ok",
  "results": [
    {"name": "database", "status": "skipped", "message": "DATABASE_URL is not set"},
    {"name": "replicate", "status": "skipped", "message": "REPLICATE_API_TOKEN is not set"},
    {"name": "r2", "status": "skipped", "message": "missing config: [R2_ACCOUNT_ID ...]"},
    {"name": "ffmpeg", "status": "ok", "message": "ffmpeg found at /usr/bin/ffmpeg"}
  ]
}
```

実環境設定時の期待レスポンス:

```json
{
  "status": "ok",
  "results": [
    {"name": "database", "status": "ok", "message": "TiDB/MySQL connection is reachable"},
    {"name": "replicate", "status": "ok", "message": "Replicate API token is valid"},
    {"name": "r2", "status": "ok", "message": "R2 bucket is reachable"},
    {"name": "ffmpeg", "status": "ok", "message": "ffmpeg found at /usr/bin/ffmpeg"}
  ]
}
```

## 5.1 ジョブ重複開始の確認

同一プロジェクトで `generation`、`inpainting`、`export` の同種ジョブが `queued` または `running` の間は、同じ種類のジョブを追加開始しない。画面上で短時間に連続実行した場合は、後続のリクエストが失敗として返り、完了後の再試行を促す。

## 6. 画面表示の確認

ブラウザで `http://localhost:3000` を開き、以下の画面が表示されることを確認する。

### 6.1 Step 1: 中割り生成画面

| URL | `http://localhost:3000/step1/generate` |
|-----|--------------------------------------|

**確認項目**:
- [ ] 「AIで中割りを生成」の見出しが表示される
- [ ] 動きの指示（テキストエリア）が表示される
- [ ] 生成枚数（数値入力）が表示される（デフォルト 6）
- [ ] 原画 1 / 原画 2 の画像アップロード領域が表示される
- [ ] デモ用の原画が自動表示される
- [ ] 「生成を開始」ボタンが表示される
- [ ] 生成を開始するとジョブ状態パネルが表示される
- [ ] 生成完了後、タイムラインにフレームが表示される
- [ ] 生成開始 API またはジョブが失敗した場合、失敗内容、再試行、失敗表示を閉じる操作が表示される

### 6.2 Step 2: Inpainting 修正画面

| URL | `http://localhost:3000/step2/inpaint` |
|-----|--------------------------------------|

**確認項目**:
- [ ] 左側にフレーム画像を背景としたキャンバスが表示される
- [ ] 対象フレーム一覧が表示され、クリックで修正対象を切り替えられる
- [ ] キャンバス上で黒いブラシでマスクが描ける
- [ ] ブラシサイズ調整、マスクの元に戻す、消去、表示切替ができる
- [ ] 「修正プロンプト」テキストエリアが表示される
- [ ] 変化量（strength）スライダーが表示される
- [ ] 「修正を実行」ボタンが表示される
- [ ] 修正実行後、ジョブ状態が表示される
- [ ] 修正開始 API またはジョブが失敗した場合、失敗内容、再試行、失敗表示を閉じる操作が表示される

### 6.3 Step 3: 手動編集画面

| URL | `http://localhost:3000/step3/edit` |
|-----|----------------------------------|

**確認項目**:
- [ ] ツールバーが表示される
  - [ ] 選択 / ペン / 四角 / 円 / 多角形 / テキスト ツール
  - [ ] 元に戻す / やり直す / 削除 ボタン
  - [ ] 複製 / 背面へ / 前面へ / 最前面へ ボタン
  - [ ] 色選択 / ブラシサイズ
  - [ ] 追加 / 保存 ボタン
- [ ] レイヤー一覧が表示され、追加した編集レイヤーを選択できる
- [ ] レイヤー一覧から表示 / 非表示を切り替えられる
- [ ] 選択中の編集レイヤーの X / Y / 拡大率 / 回転を数値で調整できる
- [ ] キャンバスが表示される
- [ ] フィルターツールバーが表示される
  - [ ] 明度 / コントラスト / 彩度 / ブラー スライダー
- [ ] 選択フレームの画像がキャンバスに読み込まれる
- [ ] 図形やテキストの追加ができる
- [ ] 保存でフレームが更新される

### 6.4 Step 4: 動画書き出し画面

| URL | `http://localhost:3000/export` |
|-----|-------------------------------|

**確認項目**:
- [ ] 「完成フレームを動画に書き出し」の見出しが表示される
- [ ] FPS スライダーが表示される（デフォルト 8）
- [ ] 書き出し前にフレーム切り替えプレビューが表示される
- [ ] 「MP4を書き出す」ボタンが表示される
- [ ] 書き出し完了後、video 要素で動画再生ができる
- [ ] 「MP4をダウンロード」ボタンが表示される
- [ ] 「再書き出し」ボタンでリセットできる
- [ ] 書き出し開始 API またはジョブが失敗した場合、失敗内容、再試行、失敗表示を閉じる操作が表示される

### 6.5 タイムライン

**確認項目**:
- [ ] 左サイドバーにフレーム一覧が表示される
- [ ] 各フレームに種類タグが表示される（生成/修正/編集）
- [ ] フレームクリックで選択状態が切り替わる
- [ ] ドラッグ & ドロップでフレームの並び替えができる

### 6.6 ナビゲーション

**確認項目**:
- [ ] 上部にステップナビゲーションが表示される
- [ ] Step 1 / Step 2 / Step 3 / Step 4 の各リンクで画面遷移する
- [ ] URL が各ステップに応じて変化する

## 7. API 動作確認

### 7.1 プロジェクト作成

```bash
curl -s -X POST http://localhost:8080/projects \
  -H 'Content-Type: application/json' \
  -d '{"id":"demo-project","name":"動作確認プロジェクト"}' | jq
```

### 7.2 フレーム生成（デモモード）

```bash
curl -s -X POST http://localhost:8080/inference/generate \
  -H 'Content-Type: application/json' \
  -d '{
    "projectId": "demo-project",
    "prompt": "キャラクターが振り向く",
    "frameCount": 4,
    "startImageDataUrl": "data:image/png;base64,iVBORw0KGgo=",
    "endImageDataUrl": "data:image/png;base64,iVBORw0KGgo="
  }' | jq
```

レスポンス例:

```json
{
  "job": {
    "id": "job-1715872800123456789",
    "projectId": "demo-project",
    "type": "generation",
    "status": "queued",
    "progress": 0,
    "message": "中割り生成を受け付けました",
    "version": 1,
    "createdAt": "2026-05-16T12:00:00Z",
    "updatedAt": "2026-05-16T12:00:00Z"
  }
}
```

### 7.3 ジョブ状態のポーリング

```bash
# JOB_ID を実際の値に置き換える
curl -s http://localhost:8080/jobs/JOB_ID | jq
```

完了時のレスポンス例:

```json
{
  "job": {
    "id": "job-1715872800123456789",
    "type": "generation",
    "status": "succeeded",
    "progress": 100,
    "message": "完了しました",
    "result": {
      "frames": [
        {"id": "frame-0-...", "index": 0, "kind": "key", "imageUrl": "data:image/svg+xml;base64,..."},
        {"id": "frame-1-...", "index": 1, "kind": "generated", "imageUrl": "data:image/svg+xml;base64,..."}
      ]
    },
    "version": 2,
    "createdAt": "2026-05-16T12:00:00Z",
    "updatedAt": "2026-05-16T12:00:01Z"
  }
}
```

### 7.4 フレーム一覧の取得

```bash
curl -s http://localhost:8080/projects/demo-project/frames | jq
```

### 7.5 フレーム更新（手動編集保存）

```bash
# FRAME_ID を実際の値に置き換える
curl -s -X PUT http://localhost:8080/projects/demo-project/frames/FRAME_ID \
  -H 'Content-Type: application/json' \
  -d '{
    "projectId": "demo-project",
    "frameId": "FRAME_ID",
    "imageDataUrl": "data:image/png;base64,iVBORw0KGgo=",
    "note": "手動編集による修正"
  }' | jq
```

### 7.6 動画書き出し

```bash
curl -s -X POST http://localhost:8080/export/video \
  -H 'Content-Type: application/json' \
  -d '{"projectId":"demo-project","fps":12}' | jq
```

## 8. テスト実行

### 8.1 API テスト

```bash
cd apps/api
go test ./... -v
```

期待される結果: 全パッケージ `ok`、全テスト `PASS`

### 8.2 フロントエンド単体テスト

```bash
bun run test:web
```

期待される結果: 全テスト `PASS`

### 8.3 E2E テスト

```bash
# 事前にフロントエンドを起動しておく
cd apps/web
bun run test:e2e
```

E2E は通常 CI でも Chromium で実行される。CI では workflow が Vite dev server を起動し、外部 API 依存はテスト内の route mock で分離する。

## 9. 本番環境の動作確認

本番環境は以下の URL でアクセスできる。

| 環境 | URL |
|------|-----|
| フロントエンド | `https://anifusion-canvas.pages.dev` |
| API サーバー | `https://anifusion-api-976317870900.asia-northeast1.run.app` |

### 9.1 本番 API ヘルスチェック

```bash
curl https://anifusion-api-976317870900.asia-northeast1.run.app/health
```

### 9.2 本番依存チェック

```bash
curl https://anifusion-api-976317870900.asia-northeast1.run.app/health/dependencies
```

### 9.3 本番フロントエンド

ブラウザで `https://anifusion-canvas.pages.dev` を開き、ローカルと同様の手順で動作確認を行う。

## 10. エラー発生時の確認ポイント

### API が起動しない

1. ポート 8080 が使用中でないか確認: `lsof -i :8080`
2. Go モジュールが最新か確認: `cd apps/api && go mod tidy`
3. 環境変数が正しいか確認: `.env` ファイルの内容を確認

### フロントエンドが起動しない

1. ポート 3000 が使用中でないか確認: `lsof -i :3000`
2. 依存関係が最新か確認: `bun install`
3. `VITE_API_BASE_URL` が正しいか確認

### フロントエンドから API に接続できない

1. API サーバーが起動しているか確認: `curl http://localhost:8080/health`
2. CORS 設定を確認: `FRONTEND_ORIGIN` 環境変数
3. ブラウザの DevTools → Network タブでエラーを確認

### 画像生成が動作しない

1. デモモードの場合は SVG が自動生成される
2. 実推論モードの場合は `REPLICATE_API_TOKEN` が有効か確認
3. `/health/dependencies` で `replicate` が `ok` か確認

### FFmpeg エラー

1. FFmpeg がインストールされているか確認: `ffmpeg -version`
2. `/health/dependencies` で `ffmpeg` が `ok` か確認
3. 十分なディスク空き容量があるか確認（`/tmp` 領域を使用）

### データベース接続エラー

1. `DATABASE_URL` が正しいか確認
2. TiDB クラスタが起動しているか確認
3. TLS 設定を確認（TiDB Serverless は TLS 必須）
4. `/health/dependencies` で `database` のステータスを確認
