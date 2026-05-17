# AnifusionCanvas

アニメ制作現場のDXをテーマにした、AI中割り支援デモアプリケーションです。

原画2枚から ToonCrafter で中割りフレームを生成し、破綻した箇所だけを SDXL Inpainting で部分修正します。制作現場で人間の判断を残しながら、AIで作業負荷を下げる Human-in-the-Loop 型のワークフローを検証するために開発しました。

## 制作意図

このアプリケーションは、AIエンジニアとしての就職活動に向けたポートフォリオです。

私は、ソニーやサイバーエージェントのようにアニメ・映像コンテンツを持つ企業で、アニメ制作現場のDX化を推進したいと考えています。中割り、修正、確認といった制作工程には、熟練したクリエイターの判断が必要な一方で、反復的で負荷の高い作業も多く存在します。

AnifusionCanvas は、その課題に対して「AIで完全自動化する」のではなく、「クリエイターが主導権を持ったままAIを使う」体験をWebアプリとして形にしたものです。

## デモURL

- Frontend: `https://anifusion-canvas.pages.dev`
- API: `https://anifusion-api-976317870900.asia-northeast1.run.app`

2026-05-18 に、本番環境で AI中割り 1回、Inpainting 1回の実推論 smoke test を完了しています。

## 主な機能

- 2枚の原画をアップロードして、AI中割りフレームを生成
- 生成されたフレームをタイムラインで確認
- 修正したいフレームを選び、黒いブラシでマスクを描画
- 自然言語プロンプトでマスク部分だけを Inpainting
- ジョブ状態の表示、失敗時の再試行導線
- 画像サイズ・形式の基本バリデーション

現在の公開UIは「AI中割り」と「Inpainting」に絞っています。

## ワークフロー

```text
Step 1: AI中割り
原画1 + 原画2 + 動きの指示
        ↓
ToonCrafter 推論
        ↓
生成フレームをタイムラインへ保存

Step 2: Inpainting
対象フレーム選択 + マスク描画 + 修正プロンプト
        ↓
SDXL Inpainting 推論
        ↓
修正済みフレームをタイムラインへ反映
```

## 技術的な見どころ

- React SPA と Go API を分離した実運用寄りの構成
- TanStack Query によるジョブポーリングと非同期状態管理
- Zustand によるタイムライン・選択フレーム状態の管理
- Fabric.js を使った Inpainting 用マスク描画
- Go + Echo による API / usecase / infrastructure の分離
- Replicate API を使った ToonCrafter / SDXL Inpainting 推論
- FFmpeg による ToonCrafter 出力動画のフレーム分割
- Cloudflare R2 への生成物保存
- TiDB Serverless によるジョブ・フレーム永続化
- Cloud Run + Cloudflare Pages への本番デプロイ
- GitHub Actions による build / test / deploy

## アーキテクチャ

```text
Cloudflare Pages
  React + Vite + TanStack Router
  Zustand / TanStack Query / Fabric.js
          |
          | HTTPS
          v
Google Cloud Run
  Go + Echo
  Handler -> Usecase -> Infrastructure
          |
          +-- Replicate
          +-- Cloudflare R2
          +-- TiDB Serverless
          +-- FFmpeg
```

## 技術スタック

| 領域 | 技術 |
|------|------|
| Frontend | React 18, Vite, TanStack Router, TanStack Query |
| State | Zustand |
| Form | React Hook Form, Zod |
| Canvas | Fabric.js |
| UI | TailwindCSS, custom components, lucide-react |
| Backend | Go 1.22, Echo |
| AI Inference | Replicate API, ToonCrafter, SDXL Inpainting |
| Media | FFmpeg |
| Storage | Cloudflare R2 |
| Database | TiDB Serverless |
| Deploy | Cloudflare Pages, Google Cloud Run |
| CI/CD | GitHub Actions |

## ローカル起動

```bash
bun install
cd apps/api && go mod tidy && cd ../..
```

API:

```bash
cd apps/api
go run ./cmd/server
```

Web:

```bash
cd apps/web
bun run dev
```

デフォルトではデモモードで動作します。実推論を使う場合は Replicate / R2 / DB 関連の環境変数を設定します。

## 検証

```bash
bun run build:web
bun run lint:web
bun run test:web
cd apps/api && go test ./...
```

本番環境では以下を確認済みです。

- `/health`: ok
- `/health/dependencies`: database / replicate / r2 / ffmpeg すべて ok
- AI中割り実推論: succeeded
- Inpainting実推論: succeeded
- R2 公開画像URL: HTTP 200

## 今後の発展

- 制作会社ごとの作画ルールやキャラクター設定を反映したプロンプト支援
- カット単位でのレビュー・承認フロー
- 修正履歴の比較表示
- スタジオ内アセット管理や制作進行ツールとの連携
- 現場で使いやすい権限管理、監査ログ、コスト管理

## ライセンス

Private
