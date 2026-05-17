# AnifusionCanvas

アニメーション制作の「中割り」工程を AI で支援する Human-in-the-Loop 型ツール。
クリエイターが介在して品質を担保しながら、ToonCrafter によるフレーム補間と
SDXL Inpainting による部分修正を行います。

現在の公開 UI は「AI中割り」と「Inpainting」に絞っています。手動編集と書き出しの既存コードや API は残していますが、通常のフロントエンド導線からは外しています。

## ワークフロー

```
Step 1:          Step 2:
原画アップロード  AI で部分修正
      ↓              ↓
┌──────────┐  ┌──────────┐
│ ToonCrafter │→│SDXL Inpaint│
│ フレーム補間 │  │ マスク修正 │
└──────────┘  └──────────┘
```

## アーキテクチャ

```
┌───────────── Cloudflare Pages ───────────────┐
│  React SPA (Vite + TanStack Router)          │
│  ┌──────┐ ┌──────┐                         │
│  │Step 1│ │Step 2│                         │
│  └──────┘ └──────┘                         │
│  Zustand / TanStack Query / Fabric.js       │
└──────────────────┬──────────────────────────┘
                   │ HTTPS
┌──────────────────▼──────────────────────────┐
│  Google Cloud Run (Go + Echo)               │
│  ┌─────────────────────────────────────┐    │
│  │ Inference / Frame / Job             │    │
│  │ Usecase → Repository → Infrastructure│   │
│  └─────────────────────────────────────┘    │
│  Replicate / R2 / FFmpeg / TiDB             │
└──────────────────┬──────────────────────────┘
         ┌─────────┼──────────┐
         ▼         ▼          ▼
    ┌────────┐ ┌──────┐ ┌──────────┐
    │  TiDB  │ │  R2  │ │Replicate │
    │Serverless│ │Storage│ │   API   │
    └────────┘ └──────┘ └──────────┘
```

## 技術スタック

| 領域 | 技術 |
|------|------|
| フロントエンド | React 18, Vite, TanStack Router/Query, Zustand |
| Canvas 編集 | Fabric.js 7 |
| フォーム / バリデーション | React Hook Form + Zod |
| UI | TailwindCSS, custom components, lucide-react |
| スタイル | TailwindCSS |
| バックエンド | Go 1.22 + Echo |
| データベース | TiDB Serverless (MySQL 互換) |
| ストレージ | Cloudflare R2 (S3 互換) |
| AI 推論 | Replicate API (ToonCrafter + SDXL Inpainting) |
| 動画処理 | FFmpeg (ToonCrafter 出力のフレーム分割) |
| テスト | Vitest, Playwright, Go testing |
| CI/CD | GitHub Actions |
| デプロイ | Cloudflare Pages + Google Cloud Run |

## ローカル開発

```bash
# 依存関係のインストール
bun install
cd apps/api && go mod tidy && cd ../..

# API サーバー起動 (http://localhost:8080)
cd apps/api && go run ./cmd/server

# フロントエンド起動 (http://localhost:3000)
cd apps/web && bun run dev
```

FFmpeg がシステムにインストールされている必要があります。

### 環境変数

| 変数 | 必須 | 説明 |
|------|------|------|
| `REPLICATE_API_TOKEN` | 実推論時 | Replicate API トークン |
| `REPLICATE_MODE` | 実推論時 | `demo`（デフォルト）または `replicate` |
| `DATABASE_URL` | DB 使用時 | TiDB/MySQL DSN |
| `STUDIO_STORE` | DB 使用時 | `memory`（デフォルト）または `database` |
| `R2_BUCKET` ほか | 画像保存時 | Cloudflare R2 設定 |
| `REPLICATE_TOONCRAFTER_VERSION` | 実推論時 | ToonCrafter の Replicate version ID |
| `REPLICATE_SDXL_INPAINTING_VERSION` | 実推論時 | SDXL Inpainting の Replicate version ID |
| `R2_PUBLIC_BASE_URL` | 実推論時 | 生成物をブラウザと Replicate から参照するための公開 R2 URL |

詳細は `.env.example` を参照してください。`REPLICATE_MODE=demo` または未設定時は、Replicate token が存在してもデモモードで動作します。

### テスト

```bash
# API テスト
cd apps/api && go test ./...

# フロントエンド単体テスト
bun run test:web

# E2E テスト (Playwright)
cd apps/web && bun run test:e2e
```

## 推論コスト

| モデル | 1回あたりのコスト目安 |
|--------|----------------------|
| ToonCrafter (中割り生成) | 約 $0.064 |
| SDXL Inpainting (部分修正) | 約 $0.003 |

1セッションあたり合計約 $0.07（約 10 円）。

## 既知の制約

- ToonCrafter の出力 MP4 はフレーム分割後に再構成するため、元の動画とフレーム数が一致しない場合があります
- Cloud Run のリクエストタイムアウト（デフォルト 300 秒）内に処理が完了する前提です
- Replicate の料金は変動する可能性があります。本番利用前に最新の料金を確認してください
- 手動編集と書き出しは通常のフロントエンド導線から外しています
- モバイル向けの最適化は限定的です

## ライセンス

Private
