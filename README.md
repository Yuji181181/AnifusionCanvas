# AnifusionCanvas

アニメ制作現場のDXをテーマにした、AI中割り支援デモアプリケーションです。

原画2枚から ToonCrafter で中割りフレームを生成し、破綻した箇所だけを SDXL Inpainting で部分修正します。制作現場で人間の判断を残しながら、AIで作業負荷を下げる Human-in-the-Loop 型のワークフローを検証するために開発しました。

## 制作意図

このアプリケーションは、AIエンジニアとしての就職活動に向けたポートフォリオです。

私は、ソニーやサイバーエージェントのようにアニメ・映像コンテンツを持つ企業で、アニメ制作現場のDX化を推進したいと考えています。中割り、修正、確認といった制作工程には、熟練したクリエイターの判断が必要な一方で、反復的で負荷の高い作業も多く存在します。

AnifusionCanvas は、その課題に対して「AIで完全自動化する」のではなく、「クリエイターが主導権を持ったままAIを使う」体験をWebアプリとして形にしたものです。

なお、これは私が実際に技術検証を行うために開発したデモであり、そのまま実用できるプロダクトではありません。目的は、現時点の生成AIでアニメ制作ワークフローのどこまでを支援できるかを自分の手で確かめ、課題と可能性を具体的に理解することです。

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
┌────────────────────────────────────────────────────────────┐
│ Frontend: Cloudflare Pages                                 │
│                                                            │
│  React SPA / Vite / TanStack Router                         │
│  TanStack Query: job polling                                │
│  Zustand: timeline and selected frame state                 │
│  Fabric.js: inpainting mask canvas                          │
└──────────────────────────────┬─────────────────────────────┘
                               │ HTTPS / JSON
                               ▼
┌────────────────────────────────────────────────────────────┐
│ Backend: Google Cloud Run                                  │
│                                                            │
│  Go + Echo                                                  │
│  HTTP Handler                                               │
│      ↓                                                     │
│  Usecase: generation / inpainting / frame / job             │
│      ↓                                                     │
│  Infrastructure                                             │
└───────────────┬──────────────┬──────────────┬──────────────┘
                │              │              │
                ▼              ▼              ▼
        ┌────────────┐ ┌──────────────┐ ┌──────────────┐
        │ Replicate  │ │ Cloudflare R2│ │ TiDB         │
        │ AI推論     │ │ 生成物保存   │ │ 状態永続化   │
        └─────┬──────┘ └──────────────┘ └──────────────┘
              │
              ▼
        ┌────────────┐
        │ FFmpeg     │
        │ MP4分割    │
        └────────────┘
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

実際に推論を行い、AI中割りを生成してみた結果、現時点の ToonCrafter の精度では、そのままアニメ制作現場で使える品質には達していないと感じました。生成結果には破綻や意図しない補間が残り、アニメーターが実務で安心して採用できる水準ではありません。

一方で、ワークフローとしては可能性があると感じました。もし中割り生成モデルの精度を十分に高めることができれば、生成結果をアニメーターが確認し、必要な部分だけを修正する形で、実際のアニメ制作ワークフローに組み込むことは可能だと考えています。

## 今後の発展

- 中割り生成モデルの開発
  - ToonCrafter は 2024 年以降大きく開発が止まっているため、近年の拡散モデルなどをベースにした、より高精度なアニメ向け中割り生成モデルが必要だと考えています。
- 作品・制作会社ごとの事後学習
  - 作品ごとのキャラクター設定、線の癖、色、演出ルールに合った推論を行うには、そのアニメ会社や作品に適した事後学習が必要です。
- 編集ソフトへの機能組み込み
  - 最適な Human-in-the-Loop を実現するには、アニメーターが作画を行う既存ソフトや制作環境にAI機能を組み込む必要があります。
  - 生成した中割り画像をアニメーターがそのまま確認・修正・採用できる編集体験が重要です。
- 制作進行・レビュー工程との連携
  - カット単位でのレビュー、修正履歴、承認フロー、制作管理ツールとの連携まで含めることで、現場導入に近づけられると考えています。

## ライセンス

Private
