# Database Setup

## 現状

`STUDIO_STORE=database` で API を起動すると、フレームとジョブを TiDB/MySQL に保存する構成になります。

今回の実DB確認では、接続自体は成功しましたが、既存DBには別用途の `projects / frames / jobs` テーブルが存在していました。

そのため、既存テーブルを壊さないようにアプリ用テーブルは `studio_projects / studio_frames / studio_jobs` として作成します。

2026-05-16 時点で、開発DBには migration version `1` を適用済みです。

確認できたエラー:

```text
Error 1054 (42S22): Unknown column 'image_url' in 'field list'
```

## Codex が対応済みのこと

- `STUDIO_STORE=memory` をデフォルトにし、既存デモ動作を壊さないようにした
- `STUDIO_STORE=database` の場合に DB-backed store を使えるようにした
- `/health/dependencies` で必須テーブル/カラムの不足を検出するようにした
- 初期マイグレーションを追加した

対象ファイル:

- `apps/api/internal/infrastructure/db/migrations/000001_create_core_tables.up.sql`
- `apps/api/internal/infrastructure/db/migrations/000001_create_core_tables.down.sql`

## 適用済みのこと

Codex が以下を実行済みです。

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -path apps/api/internal/infrastructure/db/migrations -database "$MIGRATE_DATABASE_URL" up
```

TiDB では `SERIALIZABLE` 分離レベルが拒否されるため、migration 実行時の DSN には `tidb_skip_isolation_level_check=1` を付与しています。

確認済み:

- `migrate version`: `1`
- `STUDIO_STORE=database` の `/health/dependencies`: `ok`
- `POST /inference/generate`: succeeded
- `GET /jobs/:jobId`: succeeded
- `GET /projects/db-demo-project/frames`: succeeded

## 今後あなたが行う可能性があること

DB スキーマを変更する操作は実データに影響する可能性があるため、適用前に対象 TiDB database が開発用であることを確認してください。

### 1. 現在の接続先を確認する

```bash
cd apps/api
echo "$DATABASE_URL"
```

接続先が本番データではなく、開発/検証用DBであることを確認します。

### 2. migration tool を用意する

`migrate` CLI が未インストールの場合はインストールします。

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 3. マイグレーションを適用する

```bash
cd apps/api
migrate -path internal/infrastructure/db/migrations -database "$MIGRATE_DATABASE_URL" up
```

`DATABASE_URL` が Go MySQL Driver 形式の場合、`golang-migrate` 用に先頭へ `mysql://` を付けます。

例:

```bash
MIGRATE_DATABASE_URL="mysql://$DATABASE_URL&tidb_skip_isolation_level_check=1"
```

### 4. schema health を確認する

```bash
STUDIO_STORE=database go run ./cmd/server
curl http://127.0.0.1:8080/health/dependencies
```

期待値:

```json
{
  "status": "ok"
}
```

### 5. database store を有効化する

ローカルまたは Cloud Run の環境変数に次を設定します。

```text
STUDIO_STORE=database
```

未適用のまま `STUDIO_STORE=database` にすると、API は `database schema mismatch` または `Unknown column` のようなエラーを返します。
