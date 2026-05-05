# Replicate API トークン設定手順

## 前提

このアプリケーションは AI推論に [Replicate](https://replicate.com/) のホスト型APIを使用しています。
推論機能（中割り生成・Inpainting）を利用するには、ReplicateのAPIトークンが必要です。

## 手順

### 1. Replicateアカウントの作成

1. [https://replicate.com](https://replicate.com) にアクセス
2. 「Sign up」からGitHubアカウントまたはGoogleアカウントで登録

### 2. APIトークンの取得

1. ログイン後、[https://replicate.com/account/api-tokens](https://replicate.com/account/api-tokens) にアクセス
2. 「API tokens」ページに表示されているトークンをコピー
   - トークンの形式: `r8_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX`
   - 既存のトークンがない場合は「Create token」をクリックして新規作成

### 3. `.env` ファイルにトークンを設定

```bash
# apps/api/.env を開く
```

以下の行を編集:

```env
# 変更前
REPLICATE_API_TOKEN=

# 変更後（取得したトークンを貼り付け）
REPLICATE_API_TOKEN=r8_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

### 4. 動作確認

FastAPIサーバーを起動し、トークンが読み込まれているか確認:

```bash
cd apps/api
uv run uvicorn app.main:app --reload
```

別ターミナルで確認:

```bash
cd apps/api
uv run python -c "from app.config import settings; print('Token configured:', bool(settings.replicate_api_token))"
# → Token configured: True と表示されればOK
```

## 料金について

- Replicateは従量課金制（使った分だけ課金）
- **ToonCrafter**: $0.0014/秒（A100 GPU）、1回約$0.064
- **SDXL Inpainting**: $0.000975/秒（L40S GPU）、1回約$0.003
- 1セッションあたり概算: 約$0.07（約7円）
- 詳細: [https://replicate.com/pricing](https://replicate.com/pricing)

## 注意事項

- APIトークンは **絶対にGitにコミットしないこと**（`.env` は `.gitignore` で除外済み）
- トークンが漏洩した場合は、Replicateのダッシュボードから即座に再生成すること
