# AnifusionCanvas 開発完了ロードマップ

最終更新: 2026-05-17 JST

## 目的

このドキュメントは、`docs/アプリケーション要件定義書.md`、現在の実装、GitHub の PR / Actions 状態を突き合わせ、AnifusionCanvas を開発完了まで進めるためのチェックリストを整理する。

## 現在の開発状況

### 全体

- [x] リポジトリは `main` と `origin/main` が同期している
- [x] 作業ツリーは調査開始時点でクリーン
- [x] GitHub CLI は認証済み
- [x] Git protocol は SSH 設定
- [x] GitHub の既存 PR #1 から #19 はすべて merge 済み
- [x] GitHub Issue は未登録
- [x] 直近の GitHub Actions `CI` は成功
- [x] ローカル検証で `bun run lint:web` が成功
- [x] ローカル検証で `bun run test:web` が成功
- [x] ローカル検証で `go test ./...` が成功

### フロントエンド

- [x] Vite + React SPA の基本構成がある
- [x] TanStack Router による `/step1/generate`、`/step2/inpaint`、`/step3/edit`、`/export` 相当の画面遷移がある
- [x] TanStack Query によるジョブポーリングと API 呼び出しがある
- [x] Zustand によるフレーム、選択フレーム、ジョブ状態、エディタ状態の管理がある
- [x] ステップ1の原画アップロード、生成枚数、プロンプト入力、生成開始 UI がある
- [x] ステップ2の Fabric.js マスク描画、プロンプト入力、strength 指定、Inpainting 実行 UI がある
- [x] ステップ3の Fabric.js エディタに、選択、ペン、四角、円、テキスト、色、ブラシサイズ、保存の最小機能がある
- [x] ステップ4の簡易プレビュー、FPS 指定、書き出しジョブ開始 UI がある
- [x] API client と frame store の単体テストがある
- [ ] shadcn/ui / Radix UI コンポーネントとしての体系化は未完了
- [ ] React Hook Form + Zod によるフォーム実装は未導入
- [ ] Step 3 の画像フィルター、レイヤー管理、アンドゥ / リドゥ、多角形、変形 UI は未完了
- [ ] Playwright E2E テストは設定のみで、主要ユーザージャーニーのテストは未実装
- [ ] 実運用を想定したアクセシビリティ、レスポンシブ、失敗時 UI の検証は不足

### バックエンド

- [x] Go + Echo の API サーバー構成がある
- [x] HTTP handler、router、usecase、domain、infrastructure の基本分離がある
- [x] `/health`、`/health/dependencies` がある
- [x] `/projects/:projectId/frames`、`/inference/generate`、`/inference/inpaint`、`/export/video`、`/jobs/:jobId` がある
- [x] API エラー応答形式が標準化されている
- [x] API バリデーションとルーターテストがある
- [x] memory store と database-backed store がある
- [x] TiDB / MySQL 用 migration がある
- [x] `/health/dependencies` で database、Replicate、R2、FFmpeg の到達性確認ができる
- [x] Dockerfile に FFmpeg が含まれている
- [ ] `GenerateFrames` は ToonCrafter 呼び出しではなく demo image 生成
- [ ] `InpaintFrame` は SDXL Inpainting 呼び出しではなく demo image 差し替え
- [ ] `ExportVideo` は FFmpeg 実行ではなく data URL の模擬結果
- [ ] R2 は health check のみで、画像 / 動画アップロード処理は未実装
- [ ] Replicate は account health check のみで、推論 client / polling / result mapping は未実装
- [ ] Cloud Run のスケールアウトを前提にした永続ジョブワーカー、排他、再開処理は未完了
- [ ] testcontainers-go を使う DB 結合テストは未実装

### ドキュメント / インフラ

- [x] 要件定義書がある
- [x] API contract がある
- [x] DB setup 手順がある
- [x] Cloud Run / Cloudflare Pages のデプロイ手順がある
- [x] Cloudflare Workers Builds のトラブルシュート文書がある
- [x] CI workflow が web lint / build / test と api test を実行している
- [x] Cloud Run / Cloudflare Pages の deploy workflow がある
- [ ] 本番環境変数と Secret Manager の取得元、更新手順、権限境界の整理は不足
- [ ] リリース判定用の運用チェックリストは未整備
- [ ] PRD、API contract、実装の差分を継続的に検出する仕組みは未整備

## 完了定義

- [ ] 2枚のキーフレームから ToonCrafter で中割り生成し、分割したフレームをタイムラインに表示できる
- [ ] 選択フレームにマスクを描き、SDXL Inpainting で部分修正し、結果を対象フレームへ反映できる
- [ ] 手動エディタで実用上必要な編集、保存、やり直しができる
- [ ] 完成フレーム列を FFmpeg で MP4 として書き出し、R2 経由で取得できる
- [ ] ジョブ状態とフレームメタデータが TiDB に永続化される
- [ ] 画像、マスク、生成結果、書き出し動画が R2 に保存される
- [ ] Cloud Run と Cloudflare Pages の本番デプロイで主要フローが動作する
- [ ] CI で web lint、web build、web unit、api test、主要 E2E が通る
- [ ] ドキュメントが実装、API契約、運用手順と矛盾しない

## フェーズ別ロードマップ

### Phase 0: 開発基盤の固定

- [x] `docs/development-roadmap.md` を開発の進捗管理ドキュメントとして採用する
- [x] 未登録の作業単位を GitHub Issue または project に登録する
- [x] 各 Phase を小さな PR 単位に分割する
- [ ] CI 必須チェックを branch protection に設定する
- [x] `apps/web/dist` をリポジトリ管理から外すか、生成物をコミットする方針を明文化する
- [x] `.env.example`、`apps/api/.env.example`、`apps/web/.env.example` の必須値と任意値を最新化する
- [x] `docs/api-contract.md` に `/health/dependencies` のレスポンス型を追加する

補足:

- 2026-05-17 時点で `main` は branch protection 未設定。リポジトリ設定変更を伴うため、Phase 0 の残作業として扱う。
- `apps/web/dist` は `.gitignore` で除外済み。ビルド成果物はコミットせず、CI / deploy workflow で生成する。

受け入れ基準:

- [x] すべての残作業が Issue / PR で追跡できる
- [ ] 新規参加者が README と docs からローカル起動、テスト、デプロイ準備を再現できる

検証:

- [x] `git status --short --branch`
- [x] `bun run lint:web`
- [x] `bun run test:web`
- [x] `bun run build:web`
- [x] `cd apps/api && go test ./...`

### Phase 1: API 契約とデータモデルの完成

- [x] Project の作成、取得、更新 API を追加する
- [ ] Frame の削除、並び替え、kind / note 更新 API を追加する
- [ ] Job の type 別 result schema を Go / TypeScript 双方で厳密化する
- [ ] `packages/contracts` に `/health/dependencies`、Project、storage object、export artifact の型を追加する
- [ ] API request / response の Zod schema を web 側で利用できる形に整理する
- [ ] DB migration に projects / frames / jobs の不足メタデータを追加する
- [ ] migration の down 手順とバックアップ注意点を更新する
- [ ] API contract の変更ルールに従い、Go domain、handler、TypeScript contracts、web client、tests を同一 PR で更新する

受け入れ基準:

- [ ] `docs/api-contract.md` と Go / TypeScript の JSON field が一致する
- [ ] API バリデーションの成功 / 失敗テストが route ごとにある
- [ ] 既存 demo vertical slice が壊れていない

検証:

- [ ] `bun run test:web`
- [ ] `cd apps/api && go test ./...`
- [ ] database store 有効時の `/health/dependencies`

### Phase 2: R2 ストレージ実装

- [ ] R2 client を `apps/api/internal/infrastructure/storage` に追加する
- [ ] data URL 画像を decode して R2 にアップロードする処理を追加する
- [ ] R2 object key 設計を決める
- [ ] 例: `projects/{projectId}/inputs/`、`projects/{projectId}/frames/`、`projects/{projectId}/masks/`、`projects/{projectId}/exports/`
- [ ] public URL または signed URL の採用方針を決める
- [ ] `R2_PUBLIC_BASE_URL` 未設定時の取得方法を決める
- [ ] frame の `imageUrl` / `thumbnailUrl` を R2 URL として保存する
- [ ] 手動編集保存時も data URL ではなく R2 object として保存する
- [ ] R2 upload / download / delete の単体テストを追加する
- [ ] R2 失敗時の user-facing error を整備する

受け入れ基準:

- [ ] 生成、Inpainting、手動編集の画像が R2 に保存される
- [ ] DB には永続的に参照可能な URL または object key が保存される
- [ ] local memory store と database store の両方で同じ API 契約が保たれる

検証:

- [ ] `/health/dependencies` の r2 が `ok`
- [ ] R2 bucket 上の object 作成確認
- [ ] `cd apps/api && go test ./...`

### Phase 3: Replicate 推論実装

- [ ] Replicate client を `apps/api/internal/infrastructure/replicate` に追加する
- [ ] ToonCrafter `fofr/tooncrafter` の request / response adapter を実装する
- [ ] SDXL Inpainting `lucataco/sdxl-inpainting` の request / response adapter を実装する
- [ ] 推論開始、polling、成功、失敗、timeout、cancel の状態遷移を usecase から呼び出せるようにする
- [ ] Replicate 出力 URL の download 処理を追加する
- [ ] 取得した画像 / 動画を R2 に保存する
- [ ] Replicate API の失敗詳細を API error / job error に安全に反映する
- [ ] 料金が発生するため、外部 API integration test は明示的な環境変数で opt-in にする
- [ ] mock client を使った成功 / 失敗 / timeout テストを追加する

受け入れ基準:

- [ ] ステップ1が demo image ではなく ToonCrafter 出力を使う
- [ ] ステップ2が demo image ではなく Inpainting 出力を使う
- [ ] Replicate token 未設定時は health check が skipped になり、推論 API は明確に失敗する
- [ ] 外部 API の長時間処理中も API は即時 job ID を返す

検証:

- [ ] mock client で `cd apps/api && go test ./...`
- [ ] opt-in 環境で ToonCrafter 1回の実推論
- [ ] opt-in 環境で Inpainting 1回の実推論

### Phase 4: FFmpeg フレーム処理と動画書き出し

- [ ] `apps/api/internal/infrastructure/video/ffmpeg.go` を追加する
- [ ] ToonCrafter の MP4 出力を一時ディレクトリへ download する
- [ ] FFmpeg で PNG フレームに分割する
- [ ] 分割枚数と `frameCount` の差異を吸収するルールを決める
- [ ] PNG フレームを R2 にアップロードし、DB に `generated` frame として保存する
- [ ] 完成フレームを FFmpeg 入力用に取得する
- [ ] FPS 指定で MP4 をエンコードする
- [ ] 書き出し MP4 を R2 にアップロードし、`ExportVideoResult.videoUrl` に反映する
- [ ] FFmpeg がない環境、入力フレーム不足、壊れた画像、timeout のテストを追加する
- [ ] Cloud Run の `/tmp` 容量と処理時間の制約を docs に追記する

受け入れ基準:

- [ ] ToonCrafter 出力からフレーム列を作成できる
- [ ] ユーザー操作時だけ MP4 書き出しが走る
- [ ] 書き出し結果が browser から再生またはダウンロードできる
- [ ] FFmpeg 失敗時に job が `failed` になり、エラー原因が確認できる

検証:

- [ ] sample MP4 を使ったフレーム分割テスト
- [ ] sample PNG 群を使った MP4 書き出しテスト
- [ ] Docker image 内で `ffmpeg -version`

### Phase 5: 永続ジョブ制御と Cloud Run 対応

- [ ] job 状態遷移を `queued -> running -> succeeded / failed` に限定する
- [ ] 二重実行を防ぐ DB lock または optimistic update を追加する
- [ ] Cloud Run のスケールアウト時に同一 job が複数 worker で走らない設計にする
- [ ] context timeout / cancellation を Replicate polling、R2、FFmpeg に伝播する
- [ ] job の再試行方針を決める
- [ ] 途中失敗時の一時ファイル削除と R2 object cleanup を実装する
- [ ] job の progress / message を user-facing に整理する
- [ ] 長時間ジョブの観測ログを構造化する
- [ ] Cloud Run request timeout と background goroutine の扱いを再評価する
- [ ] 必要なら Cloud Tasks、Cloud Run Jobs、または Pub/Sub への移行判断を記録する

受け入れ基準:

- [ ] database store 有効時、ジョブ状態が Cloud Run 再起動後も参照できる
- [ ] 同じ job が二重に成功結果を書き込まない
- [ ] 失敗、timeout、cancel の扱いが API contract と UI に反映される

検証:

- [ ] usecase の状態遷移テスト
- [ ] DB-backed store の結合テスト
- [ ] Cloud Run 上で長時間 job の疎通確認

### Phase 6: フロントエンド機能完成

- [ ] shadcn/ui / Radix UI の導入方針を決め、既存 UI を段階的に置き換える
- [ ] React Hook Form + Zod で step1 / step2 / export の入力検証を実装する
- [ ] API error をフォーム、toast、job status に適切に表示する
- [ ] Step 1 にアップロード画像の validation、削除、差し替え、生成中 disable を追加する
- [ ] Step 2 にマスクブラシサイズ、マスク表示切替、undo、clear、対象フレーム切替を追加する
- [ ] Step 3 に画像読み込みを Fabric object として実装し直す
- [ ] Step 3 に画像フィルター、レイヤー管理、重ね順、表示 / 非表示を追加する
- [ ] Step 3 にアンドゥ / リドゥを追加する
- [ ] Step 3 に変形、選択削除、オブジェクト複製を追加する
- [ ] Timeline に frame kind、更新状態、選択状態、drag reorder を追加する
- [ ] Export に動画再生、download、再書き出し、書き出し中 disable を追加する
- [ ] モバイル / タブレット / デスクトップで主要画面のレイアウトを確認する

受け入れ基準:

- [ ] 要件定義書のステップ1からステップ4の操作が UI から完結する
- [ ] 生成失敗、Inpainting 失敗、保存失敗、書き出し失敗からユーザーが復帰できる
- [ ] フレーム選択、編集、保存、再取得で UI と DB が同期する

検証:

- [ ] `bun run lint:web`
- [ ] `bun run test:web`
- [ ] 主要コンポーネントの unit test
- [ ] Playwright による主要フロー E2E

### Phase 7: テスト戦略の完成

- [ ] web component / hook / store / API client のテストを主要機能へ拡張する
- [ ] MSW などの API mock 方針を決める
- [ ] Playwright で「生成、Inpainting、手動編集、書き出し」の happy path を追加する
- [ ] Playwright で API failure / job failure の recovery path を追加する
- [ ] Go handler / usecase / infrastructure の単体テストを追加する
- [ ] testcontainers-go で TiDB 互換または MySQL 結合テストを追加する
- [ ] R2 / Replicate / FFmpeg は mock と opt-in integration を分ける
- [ ] CI に E2E を追加する
- [ ] CI の所要時間と flaky test 対策を記録する

受け入れ基準:

- [ ] CI で品質ゲートが自動確認される
- [ ] 外部サービス未設定でも通常 CI は成功する
- [ ] 外部サービスを使う検証は手動または protected environment で実行できる

検証:

- [ ] `bun run lint:web`
- [ ] `bun run test:web`
- [ ] `bunx playwright test`
- [ ] `cd apps/api && go test ./...`

### Phase 8: デプロイと本番運用準備

- [ ] Cloud Run 用 Secret Manager の secret 名、取得元、更新手順を docs に追記する
- [ ] Cloudflare Pages の `VITE_API_BASE_URL` と Cloud Run の `FRONTEND_ORIGIN` を確認する
- [ ] Cloud Run の memory / CPU / timeout / concurrency を実ジョブに合わせて調整する
- [ ] Artifact Registry、Cloud Run、Cloudflare Pages の権限を最小化する
- [ ] R2 bucket の CORS、public access、lifecycle rule を設定する
- [ ] TiDB の接続数、TLS、接続 timeout を確認する
- [ ] Replicate token の保管、ローテーション、コスト上限の運用を決める
- [ ] 本番 smoke test 手順を作成する
- [ ] rollback 手順を作成する
- [ ] GitHub Actions の手動 deploy workflow を本番手順として確定する

受け入れ基準:

- [ ] Cloud Run `/health` が `ok`
- [ ] Cloud Run `/health/dependencies` の database、replicate、r2、ffmpeg が `ok`
- [ ] Cloudflare Pages から Cloud Run API へ CORS エラーなく通信できる
- [ ] 本番 URL で主要フローが完走する

検証:

- [ ] `gh run list`
- [ ] `curl https://<cloud-run-url>/health`
- [ ] `curl https://<cloud-run-url>/health/dependencies`
- [ ] Cloudflare Pages の公開 URL で手動 smoke test

### Phase 9: ポートフォリオ完成度の仕上げ

- [ ] デモ用のサンプルキーフレームを用意する
- [ ] セッションあたりの推論コストを README または docs に明記する
- [ ] 失敗時に高額な再試行が連発しない UI / API 制限を入れる
- [ ] README にアプリの狙い、技術スタック、アーキテクチャ、動作デモ手順を追加する
- [ ] スクリーンショットまたは短いデモ動画を用意する
- [ ] 技術的な見どころを整理する
- [ ] 既知の制約を明記する
- [ ] 最終レビューで旧技術名、本文 / 表 / 図の不一致を確認する

受け入れ基準:

- [ ] 第三者が README だけでアプリの価値と構成を理解できる
- [ ] 採用 / 評価者向けに、AI、Go、React、Cloud Run、R2、TiDB の実装力が伝わる
- [ ] 実装済み機能と未実装機能の境界が曖昧になっていない

検証:

- [ ] docs 全体の用語検索
- [ ] README の手順再実行
- [ ] 本番 smoke test

## 推奨 PR 分割

- [x] PR 1: Roadmap と Issue 整備 (#20)
- [ ] PR 2: API contract / Project API / contracts 型補強 (#21)
- [ ] PR 3: R2 storage client と画像保存 (#22)
- [ ] PR 4: Replicate client と mock 推論テスト (#23)
- [ ] PR 5: ToonCrafter 生成フロー (#23)
- [ ] PR 6: SDXL Inpainting フロー (#23)
- [ ] PR 7: FFmpeg フレーム分割 (#24)
- [ ] PR 8: FFmpeg MP4 書き出し (#24)
- [ ] PR 9: 永続ジョブ制御と排他 (#25)
- [ ] PR 10: Step 3 エディタ機能拡張 (#26)
- [ ] PR 11: Playwright E2E (#27)
- [ ] PR 12: デプロイ / Secret / smoke test 整備 (#28)
- [ ] PR 13: README とポートフォリオ仕上げ (#29)

## リスクと未確定事項

- [ ] ToonCrafter の Replicate API 入出力仕様は実 API で確認が必要
- [ ] SDXL Inpainting のモデル選定は品質、速度、料金の再評価が必要
- [ ] Cloud Run の background goroutine はリクエスト終了後やインスタンス停止時の扱いに注意が必要
- [ ] R2 の public URL / signed URL 方針は UX とセキュリティの両面で決定が必要
- [ ] TiDB 上の既存テーブルと `studio_*` テーブルの共存方針は継続確認が必要
- [ ] 生成 AI の料金が発生するため、E2E / integration test の実行条件を明確に分ける必要がある
- [ ] `apps/web/dist` の取り扱いが未確定の場合、PR 差分が大きくなりやすい

## 調査で確認した主な根拠

- `apps/api/internal/usecase/studio_service.go`: 現在の生成、Inpainting、書き出しは demo / 模擬処理
- `apps/api/internal/infrastructure/dependency/checker.go`: database、Replicate、R2、FFmpeg の health check は実装済み
- `apps/api/internal/bootstrap/app.go`: `STUDIO_STORE=database` で database-backed store に切り替わる
- `apps/web/src/features/frame-generation/components/generation-panel.tsx`: Step 1 UI と polling が実装済み
- `apps/web/src/features/inpainting/components/inpainting-panel.tsx`: Step 2 の Fabric.js マスク UI が実装済み
- `apps/web/src/features/editor/components/editor-panel.tsx`: Step 3 の最小編集 UI が実装済み
- `apps/web/src/features/export/components/export-panel.tsx`: Step 4 の preview / export UI が実装済み
- `docs/api-contract.md`: 現行 HTTP 契約が整理済み
- `docs/database-setup.md`: database-backed store と migration の適用状況が整理済み
- GitHub PR #1 から #19: demo vertical slice、DB、validation、CI、API contract まで merge 済み
- GitHub Actions: 直近 `CI` は success

## 今回実行した検証

- [x] `git status --short --branch`
- [x] `gh auth status`
- [x] `gh repo view --json nameWithOwner,defaultBranchRef,url`
- [x] `gh pr list --state all --limit 20`
- [x] `gh issue list --state all --limit 30`
- [x] `gh run list --limit 15`
- [x] `bun run lint:web`
- [x] `bun run test:web`
- [x] `cd apps/api && go test ./...`
