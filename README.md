# AnifusionCanvas

アニメーション制作の「中割り」工程を AI で支援する Human-in-the-Loop 型ツール。
クリエイターが介在して品質を担保しながら、ToonCrafter によるフレーム補間と
SDXL Inpainting による部分修正を行い、最終的に MP4 動画として書き出すことができます。

## ワークフロー

```
Step 1:          Step 2:          Step 3:          Step 4:
原画アップロード  AI で部分修正    手動編集          動画に書き出し
      ↓              ↓              ↓              ↓
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│ ToonCrafter │→│SDXL Inpaint│→│ Fabric.js  │→│  FFmpeg  │
│ フレーム補間 │  │ マスク修正 │  │ 直接編集  │  │ MP4 出力 │
└──────────┘  └──────────┘  └──────────┘  └──────────┘
```

## アーキテクチャ

```
┌───────────── Cloudflare Pages ───────────────┐
│  React SPA (Vite + TanStack Router)          │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌────────┐     │
│  │Step 1│ │Step 2│ │Step 3│ │Export  │     │
│  └──────┘ └──────┘ └──────┘ └────────┘     │
│  Zustand / TanStack Query / Fabric.js       │
└──────────────────┬──────────────────────────┘
                   │ HTTPS
┌──────────────────▼──────────────────────────┐
│  Google Cloud Run (Go + Echo)               │
│  ┌─────────────────────────────────────┐    │
│  │ Inference / Frame / Export / Job    │    │
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
| 動画処理 | FFmpeg (os/exec) |
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
| `DATABASE_URL` | DB 使用時 | TiDB/MySQL DSN |
| `STUDIO_STORE` | DB 使用時 | `memory`（デフォルト）または `database` |
| `R2_BUCKET` ほか | 画像保存時 | Cloudflare R2 設定 |
| `REPLICATE_TOONCRAFTER_VERSION` | 実推論時 | ToonCrafter モデルバージョン |
| `REPLICATE_SDXL_INPAINTING_VERSION` | 実推論時 | SDXL Inpainting モデルバージョン |

詳細は `apps/api/.env.example` を参照してください。未設定時はデモモードで動作します。

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
- レイヤー管理、画像フィルターは基本機能のみ実装済みです
- モバイル向けの最適化は限定的です

## ライセンス

Private
