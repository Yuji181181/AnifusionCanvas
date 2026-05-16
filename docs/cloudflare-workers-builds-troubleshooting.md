# Cloudflare Workers Builds の確認手順

## 現状

PR の外部チェック `Workers Builds: anifusioncanvas` は Cloudflare Pages ではなく Cloudflare Workers Builds のチェックです。

このリポジトリはフロントエンド配信を Cloudflare Pages 前提で用意していますが、Cloudflare 側に `anifusioncanvas` という Worker/Workers Builds 連携も存在しています。Workers Builds はリポジトリルートの `wrangler.toml` を参照するため、ルートに Worker Static Assets 用の設定を追加しています。

GitHub の check-run API には Cloudflare の詳細ログ本文が返っておらず、失敗理由は Cloudflare Dashboard の Build log で確認する必要があります。

## Codex が対応済みのこと

- ルートに `wrangler.toml` を追加
- Root directory が `apps/web` の場合にも対応できるよう `apps/web/wrangler.toml` を追加
- Worker 名を Cloudflare チェック名に合わせて `anifusioncanvas` に設定
- `apps/web/dist` を Workers Static Assets の配信ディレクトリに設定
- SPA の直接URLアクセスに対応するため `not_found_handling = "single-page-application"` を設定

## あなたが確認すること

Cloudflare Dashboard で次を確認してください。

1. Cloudflare Dashboard にログインする
2. `Workers & Pages` を開く
3. `anifusioncanvas` Worker を開く
4. `Settings` → `Builds` を開く
5. Git repository が `Yuji181181/AnifusionCanvas` になっていることを確認する
6. Build command が次になっていることを確認する

```bash
bun install && bun run build:web
```

7. Deploy command が次になっていることを確認する

```bash
npx wrangler deploy
```

8. Root directory を確認する

Root directory が空、または `/` の場合:

- リポジトリルートの `wrangler.toml` が使われます
- Build command は `bun install && bun run build:web`
- Deploy command は `npx wrangler deploy`

9. Environment variables に次が入っていることを確認する

```text
VITE_API_BASE_URL=https://anifusion-api-976317870900.asia-northeast1.run.app
```

10. Build log を開き、エラー本文を確認する

よくある原因は次です。

- Root directory と `wrangler.toml` の場所が一致していない
- Build command が Root directory と合っていない
- Deploy command が `npx wrangler deploy` ではない
- Cloudflare 側の GitHub App / repository access が切れている
- `anifusioncanvas` Worker の Builds 連携が不要なのに有効なままになっている

Root directory が `apps/web` の場合:

- `apps/web/wrangler.toml` が使われます
- Build command は `bun install && bun run build`
- Deploy command は `npx wrangler deploy`

## まだ失敗する場合

Cloudflare 側に Bun が入らない、または Bun バージョンが合わない場合は、Build command を次に変更してください。

```bash
npm install -g bun && bun install && bun run build:web
```

もし `Workers Builds: anifusioncanvas` を使わず Cloudflare Pages だけで運用する場合は、Cloudflare Dashboard 側で `anifusioncanvas` Worker の Git 連携を無効化してください。その場合、PR の必須チェック設定から `Workers Builds: anifusioncanvas` も外す必要があります。

## 参考

- Cloudflare Workers Static Assets: https://developers.cloudflare.com/workers/static-assets/
- Workers Builds configuration: https://developers.cloudflare.com/workers/ci-cd/builds/configuration/
- Migrate from Pages to Workers: https://developers.cloudflare.com/workers/static-assets/migrate-from-pages/
