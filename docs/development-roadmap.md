# AnifusionCanvas 現在進捗と開発ロードマップ

最終更新: 2026-05-17 JST

## 目的

このドキュメントは、`docs/アプリケーション要件定義書.md`、現在の実装、GitHub の Issue / PR / Actions 状態を突き合わせ、AnifusionCanvas の現在地と今後の開発順序を整理する。

ロードマップのチェックは GitHub Issue の open / close だけではなく、merge 済み PR とコード実態を優先して判定する。2026-05-17 時点では、Phase 0 から Phase 9 の Issue #20 から #29 はすべて open のままだが、多くの実装 PR は merge 済みである。

## 調査サマリー

### GitHub / リポジトリ状態

- [x] `main` と `origin/main` は同期している
- [x] 調査開始時点の作業ツリーはクリーン
- [x] GitHub CLI は `Yuji181181` アカウントで認証済み
- [x] Git protocol は SSH
- [x] default branch は `main`
- [x] PR #1 から #55 はすべて merge 済み
- [x] 直近 CI は成功
- [x] Phase 管理用 Issue #20 から #29 は存在する
- [ ] Issue #20 から #29 は実装済み項目があっても open のまま
- [ ] `main` の branch protection は未設定

### 実装の現在地

| 領域 | 状態 | メモ |
| --- | --- | --- |
| フロントエンド基盤 | 概ね実装済み | Vite + React SPA、TanStack Router、TanStack Query、Zustand、Fabric.js を使用 |
| フォーム検証 | 実装済み | React Hook Form + Zod を Step 1、Step 2、Export に導入済み |
| UI コンポーネント体系 | 未完了 | shadcn/ui / Radix UI の依存と `components/ui` 体系は未導入 |
| Step 1 中割り生成 | 実装済み | Replicate 設定時は ToonCrafter、未設定時は demo 生成 |
| Step 2 Inpainting | 実装済み | Replicate 設定時は SDXL Inpainting、未設定時は demo 差し替え |
| Step 3 手動編集 | 一部実装済み | 画像読み込み、ペン、図形、テキスト、削除、undo / redo、基本フィルター、保存に対応 |
| Timeline | 一部実装済み | frame kind 表示、選択、API 同期、drag reorder に対応 |
| Export | 実装済み | FFmpeg で MP4 を生成し、R2 設定時は R2 に保存 |
| API 基盤 | 実装済み | Go + Echo、handler / usecase / infrastructure の基本分離あり |
| DB 永続化 | 実装済み | `STUDIO_STORE=database` で TiDB / MySQL backed store に切り替え |
| R2 ストレージ | 実装済み | 入力、生成、マスク、編集、書き出し動画を object として保存可能 |
| Replicate 推論 | 実装済み | client、adapter、polling、download、mock tests あり |
| FFmpeg | 実装済み | MP4 分割と PNG 群からの MP4 エンコードあり |
| ジョブ制御 | 一部実装済み | queued / running / succeeded / failed、version による楽観的ロック、失敗時 cleanup あり |
| E2E | 初期実装済み | Playwright の smoke / navigation / validation / timeline / editor tests あり |
| CI | 一部実装済み | web lint / build / unit test / Playwright E2E、api test を実行 |
| 本番運用 | 一部整備済み | deploy docs と operations guide あり。Secret / 権限 / smoke / rollback は追加整理が必要 |

## 完了定義と現在の充足状況

- [x] 2枚のキーフレームから ToonCrafter 推論を開始できる
- [x] ToonCrafter の MP4 出力をダウンロードし、FFmpeg で PNG フレームへ分割できる
- [x] 分割フレームをタイムラインへ保存できる
- [x] 選択フレームにマスクを描き、SDXL Inpainting を開始できる
- [x] Inpainting 結果を対象フレームへ反映できる
- [x] 手動エディタで基本的な編集、保存、undo / redo ができる
- [x] 完成フレーム列を FFmpeg で MP4 に書き出せる
- [x] R2 設定時に画像、マスク、生成動画、書き出し動画を保存できる
- [x] ジョブ状態とフレームメタデータを TiDB / MySQL に永続化できる
- [ ] 実 Replicate 環境で ToonCrafter と SDXL Inpainting の end-to-end 品質確認が完了している
- [ ] Cloud Run と Cloudflare Pages の本番 URL で主要フローが完走している
- [x] CI で Playwright E2E まで自動実行されている
- [ ] shadcn/ui / Radix UI の導入方針と実装体系が確定している
- [ ] testcontainers-go による DB 結合テストがある
- [ ] Secret Manager、権限、rollback、コスト上限、smoke test が本番運用手順として固定されている

## フェーズ別進捗

### Phase 0: 開発基盤の固定

Issue: #20
関連 PR: #30, #53, #55

- [x] `docs/development-roadmap.md` を進捗管理ドキュメントとして採用
- [x] Phase 0 から Phase 9 を Issue #20 から #29 として登録
- [x] PR 単位で小さく開発する運用が始まっている
- [x] `apps/web/dist` は `.gitignore` で除外し、ビルド成果物は CI / deploy workflow で生成する方針
- [x] `.env.example` 系の整理が進んでいる
- [x] `/health/dependencies` の契約は `docs/api-contract.md` に記載済み
- [x] README はポートフォリオ向けに改訂済み
- [x] `docs/operations-guide.md` が追加済み
- [ ] `main` の branch protection と必須 CI チェックは未設定
- [ ] Phase Issue は進捗に合わせた close / split / update が未完了

次の作業:

- GitHub の branch protection で `CI / web` と `CI / api` を必須化する
- Issue #20 から #29 の本文を現在の実装状況に合わせて更新し、完了済み Phase は close する
- README、operations guide、deployment guide の手順を実際に再実行して差分を潰す

### Phase 1: API 契約とデータモデルの完成

Issue: #21
関連 PR: #17, #18, #19, #31, #32, #33, #38, #39, #44

- [x] Project の作成、取得、更新 API を追加
- [x] Frame の削除、並び替え、kind / note 更新 API を追加
- [x] Job の type 別 result schema を Go / TypeScript 双方で定義
- [x] `packages/contracts` に Project、storage object、export artifact、dependency check の型を追加
- [x] Web API client で Zod schema を使って response を検証
- [x] API error response を標準化
- [x] `studio_jobs.version` を追加し、DB store の `UpdateJob` で楽観的ロックを実装
- [ ] DB migration の down 手順と本番バックアップ注意点の記述がまだ薄い
- [ ] API contract、Go domain、TypeScript contracts、web client の差分検出は手動

次の作業:

- `docs/database-setup.md` に migration rollback、バックアップ、既存データ影響の章を追加する
- API 契約変更時のチェックリストを PR template または docs に追加する
- OpenAPI または contract test の導入可否を判断する

### Phase 2: R2 ストレージ実装

Issue: #22
関連 PR: #34, #35, #36, #37, #42, #43, #45

- [x] R2 client を `apps/api/internal/infrastructure/storage` に追加
- [x] data URL 画像を decode して R2 にアップロードする処理を追加
- [x] object key は `projects/{projectId}/inputs/`、`frames/`、`masks/`、`generated/`、`exports/` 系で整理
- [x] `R2_PUBLIC_BASE_URL` があれば public URL、なければ `r2://` URI を返す方針
- [x] 生成、Inpainting、手動編集、書き出し動画で R2 保存に対応
- [x] R2 upload / delete 周辺のテストを追加
- [x] 失敗時の job failure と cleanup を追加
- [ ] R2 bucket の CORS、public access、lifecycle rule の本番設定は未確定
- [ ] R2 の実 bucket に対する smoke test 結果は未記録

次の作業:

- `docs/storage.md` と `docs/deployment-guide.md` に CORS、public access、lifecycle rule、検証コマンドを追記する
- 本番 bucket で object 作成、公開 URL 取得、削除の smoke test を実施する
- signed URL が必要なケースを整理する

### Phase 3: Replicate 推論実装

Issue: #23
関連 PR: #40

- [x] Replicate client を追加
- [x] ToonCrafter `fofr/tooncrafter` の request / response adapter を実装
- [x] SDXL Inpainting `lucataco/sdxl-inpainting` の request / response adapter を実装
- [x] create prediction、polling、terminal status、timeout、download を実装
- [x] Replicate 出力 URL の download 処理を追加
- [x] 取得した画像 / 動画を R2 に保存する経路を追加
- [x] mock client による成功 / 失敗 / timeout テストを追加
- [x] Replicate token 未設定時は demo mode で動作する
- [ ] 実 Replicate API で ToonCrafter 1回、SDXL Inpainting 1回の opt-in 検証が未記録
- [ ] モデル version 指定の最新性、料金、品質の再評価が未完了

次の作業:

- `REPLICATE_API_TOKEN` を設定した opt-in integration 手順を `docs/operations-guide.md` に追加する
- ToonCrafter / SDXL Inpainting の実レスポンス例と失敗例を docs に保存する
- 料金変動に備え、README のコスト目安に確認日と取得元を明記する

### Phase 4: FFmpeg フレーム処理と動画書き出し

Issue: #24
関連 PR: #41, #42, #43

- [x] `apps/api/internal/infrastructure/media/ffmpeg.go` を追加
- [x] ToonCrafter の MP4 出力を一時ディレクトリへ保存して分割
- [x] FFmpeg で PNG フレームに分割
- [x] `frameCount` と分割枚数に差がある場合、足りない分は最後のフレームを再利用
- [x] PNG フレームを R2 にアップロードし、DB に `generated` frame として保存
- [x] 完成フレームを取得して FFmpeg 入力に変換
- [x] FPS 指定で MP4 をエンコード
- [x] 書き出し MP4 を R2 にアップロードし、`ExportVideoResult.videoUrl` に反映
- [x] FFmpeg 不在、入力不足、壊れた入力、context cancel のテストを追加
- [ ] Cloud Run の `/tmp` 容量、timeout、CPU / memory の制約整理が未完了
- [ ] Docker image 内での `ffmpeg -version` 検証結果がリリース判定に組み込まれていない

次の作業:

- `docs/deployment-guide.md` に Cloud Run の `/tmp`、timeout、memory、CPU、concurrency の推奨値を書く
- 本番相当 Docker image で `ffmpeg -version` と sample encode を実行する
- FFmpeg stderr の詳細を job error / logs にどこまで出すか決める

### Phase 5: 永続ジョブ制御と Cloud Run 対応

Issue: #25
関連 PR: #44, #45

- [x] job 状態遷移を `queued -> running -> succeeded / failed` に制限
- [x] `studio_jobs.version` による楽観的ロックを追加
- [x] context timeout を Replicate polling、download、FFmpeg に伝播
- [x] 途中失敗時の一部 R2 object cleanup を追加
- [x] job progress / message を user-facing に整理
- [x] 長時間 job の基本的な構造化ログを追加
- [ ] Cloud Run スケールアウト時に同一 job が複数 worker で走らない設計は十分に固定されていない
- [ ] request 終了後の background goroutine と Cloud Run インスタンス停止の扱いがリスクとして残る
- [ ] retry / cancel API は未実装
- [ ] Cloud Tasks、Cloud Run Jobs、Pub/Sub への移行判断は未記録

次の作業:

- Cloud Run の request lifecycle と background goroutine のリスクを設計メモにする
- 長時間処理を Cloud Tasks / Pub/Sub に切り出すか、現行構成でデモ用途に限定するかを決める
- job retry、cancel、resume の要否を Phase 8 の運用要件と合わせて決める

### Phase 6: フロントエンド機能完成

Issue: #26
関連 PR: #46, #47, #48, #49, #50

- [x] React Hook Form + Zod で Step 1、Step 2、Export の入力検証を実装
- [x] API error を form / job status に表示する基礎を追加
- [x] Step 1 にアップロード画像、生成枚数、生成中 disable を追加
- [x] Step 2 にマスク描画、strength、clear、実行中 disable を追加
- [x] Step 3 に Fabric image 読み込みを追加
- [x] Step 3 に brightness / contrast / saturation / blur の基本フィルターを追加
- [x] Step 3 に undo / redo、削除、図形、テキスト、保存を追加
- [x] Timeline に frame kind、選択状態、drag reorder を追加
- [x] Export に動画再生、download、再書き出し、書き出し中 disable を追加
- [ ] shadcn/ui / Radix UI の導入方針は未確定
- [ ] Step 2 のブラシサイズ調整、マスク表示切替、undo、対象フレーム切替は未完了
- [ ] Step 3 のレイヤー管理、重ね順、表示 / 非表示、複製、多角形、変形 UI は未完了
- [ ] API failure / job failure からの復帰 UX は不足
- [ ] モバイル / タブレット / デスクトップの網羅確認は未記録

次の作業:

- UI 方針を「shadcn/ui を導入する」か「既存 Tailwind custom UI を正式採用する」かで決める
- Step 2 のブラシ UX と失敗復帰を優先して仕上げる
- Step 3 はポートフォリオ用途で必要な編集機能に絞り、レイヤー管理を入れるかを判断する
- Playwright screenshot または手動チェックで主要 viewport の崩れを記録する

### Phase 7: テスト戦略の完成

Issue: #27
関連 PR: #51, #52, #54

- [x] Web API client と frame store の unit test がある
- [x] Playwright E2E の初期テストを追加
- [x] Go handler / usecase / infrastructure のテストを拡張
- [x] memory store の単体テストを追加
- [x] FFmpeg は実行可能なら実 FFmpeg を使い、なければ skip するテストを追加
- [x] Replicate は mock client によるテストを追加
- [ ] MSW などの API mock 方針は未決定
- [ ] Playwright は happy path の一部に留まり、Inpainting、保存、export の完全フローは未完了
- [ ] API failure / job failure の recovery path E2E は未実装
- [ ] testcontainers-go で TiDB 互換または MySQL 結合テストは未実装
- [x] CI に Playwright E2E を統合
- [ ] flaky test 対策と CI 所要時間の記録は未整備

次の作業:

- Playwright の CI 実行時間を見ながら、ブラウザ install の cache 方針を必要に応じて見直す
- E2E を「demo mode happy path」「API failure」「job failure」「export」の順に増やす
- DB store は testcontainers-go または Docker Compose のどちらで検証するか決める
- 外部サービス検証は protected environment または手動 opt-in に分離する

### Phase 8: デプロイと本番運用準備

Issue: #28
関連 PR: #5, #6, #7, #55

- [x] Cloud Run deploy workflow と Cloudflare Pages deploy workflow がある
- [x] Secret Manager を使う deploy script がある
- [x] database store を Cloud Run で有効化する設定がある
- [x] R2 public base URL を optional として扱う修正済み
- [x] `docs/operations-guide.md` に動作確認手順を追加
- [ ] Secret Manager の secret 名、取得元、更新手順、権限境界の整理はまだ不足
- [ ] `VITE_API_BASE_URL` と `FRONTEND_ORIGIN` の本番値の確認結果が未記録
- [ ] Cloud Run の memory / CPU / timeout / concurrency が実ジョブ基準で未調整
- [ ] Artifact Registry、Cloud Run、Cloudflare Pages、R2、TiDB の最小権限整理が未完了
- [ ] R2 bucket の CORS、public access、lifecycle rule が未確定
- [ ] Replicate token の保管、ローテーション、コスト上限運用が未確定
- [ ] 本番 smoke test と rollback 手順が未完成

次の作業:

- `docs/deployment-guide.md` に secret 一覧、取得元、更新手順、検証コマンドを追記する
- Cloud Run `/health` と `/health/dependencies` の本番確認結果を記録する
- Pages 公開 URL から Cloud Run API へ CORS エラーなく通信できることを確認する
- rollback 手順を deploy workflow 単位で書く

### Phase 9: ポートフォリオ完成度の仕上げ

Issue: #29
関連 PR: #53

- [x] README にアプリの狙い、技術スタック、アーキテクチャ、ローカル実行、テスト、既知制約を追加
- [x] README に推論コスト目安を追加
- [x] 技術的な見どころを README から読み取れる状態に整理
- [x] 実装済み機能と既知制約を分けて記載
- [ ] デモ用のサンプルキーフレームは未整備
- [ ] スクリーンショットまたは短いデモ動画は未整備
- [ ] 失敗時に高額な再試行が連発しない UI / API 制限は未実装
- [ ] README の `React 19` 記述と実依存 `react@18.3.1` に不一致がある
- [ ] README の `shadcn/ui, Radix UI` 記述は要件上の前提だが、実依存には未導入
- [ ] 最終レビューで旧技術名、本文 / 表 / 図の不一致を再確認する必要がある

次の作業:

- README の技術スタックを実依存に合わせて修正する
- デモ素材、スクリーンショット、短い動画を追加する
- 推論コストの取得日と再確認手順を明記する
- 採用 / 評価者が 5 分で価値を把握できる demo path を用意する

## 推奨される今後の優先順位

### 1. 進捗管理の現実合わせ

目的: GitHub Issue と docs のズレをなくし、次に何をやるかを迷わない状態にする。

- Issue #20 から #29 の本文をこのドキュメントに合わせて更新
- 完了済みの Phase または完了済み subtask を close / split
- branch protection を設定
- README の技術スタック不一致を修正

完了条件:

- GitHub Issue を見れば未完了作業だけが残っている
- `main` への PR で CI が必須になる
- README と実依存の矛盾がない

### 2. 本番 smoke test の固定

目的: すでに実装済みの R2、Replicate、FFmpeg、DB が本番相当で動くことを証明する。

- Cloud Run `/health`、`/health/dependencies` を確認
- Cloudflare Pages から Cloud Run API へ通信確認
- R2 object 作成、公開 URL、削除確認
- TiDB migration 適用と DB-backed store の疎通確認
- Replicate ToonCrafter / SDXL Inpainting を各 1 回 opt-in で実行
- Export MP4 を R2 経由で再生 / download 確認

完了条件:

- `docs/operations-guide.md` に実行日、環境、結果、失敗時の対処が記録されている
- 本番 URL で主要フローが完走する

### 3. ジョブ基盤の運用判断

目的: Cloud Run の background goroutine 前提を、デモとして許容するか、より堅牢な非同期基盤に切り替えるか決める。

- 現行 goroutine 方式のリスクを docs に明文化
- request timeout、instance shutdown、scale-out、retry、二重実行の扱いを整理
- Cloud Tasks / Pub/Sub / Cloud Run Jobs の採用可否を判断
- 採用しない場合は「デモ用途の制約」として README と operations guide に明記

完了条件:

- 長時間ジョブの運用リスクがドキュメント化されている
- 必要なら実装 Issue が作られている

### 4. E2E と CI の強化

目的: 現在の実装済みワークフローを継続的に壊さないようにする。

- Playwright E2E を CI で安定運用する
- demo mode の生成 -> Inpainting -> 手動編集 -> Export happy path を追加
- API failure / job failure recovery path を追加
- testcontainers-go または Docker Compose による DB store 結合テストを追加
- 外部 API integration は手動 opt-in に分離

完了条件:

- 通常 CI は外部サービス未設定でも成功する
- E2E が主要画面の破損を検知できる
- DB store の基本 CRUD / job update が実 DB で検証される

### 5. フロントエンド UX の仕上げ

目的: ポートフォリオとして触ったときの完成度を上げる。

- Step 2 のブラシサイズ、mask undo、mask clear、対象フレーム切替を改善
- Step 3 のレイヤー管理または最低限の重ね順操作を追加
- 失敗時 toast / retry / disabled state を整理
- モバイル、タブレット、デスクトップで主要画面を確認
- shadcn/ui / Radix UI を導入するか、現行 custom UI を正式方針にするか決める

完了条件:

- 要件定義書の Step 1 から Step 4 を UI から自然に完走できる
- 失敗してもユーザーが復帰できる
- 実装方針と docs の UI 技術説明が一致する

### 6. ポートフォリオ素材の完成

目的: 第三者に価値が伝わる状態にする。

- デモ用キーフレームを用意
- README にスクリーンショットまたは短いデモ動画を追加
- 実行コストと制約を最新情報で更新
- 技術的見どころを「React」「Go」「Replicate」「R2」「TiDB」「Cloud Run」「FFmpeg」に分けて整理

完了条件:

- README だけでアプリの価値、構成、起動、検証、制約が分かる
- 採用 / 評価者が短時間で主要フローを理解できる

## GitHub PR の実装履歴

| 範囲 | PR | 内容 |
| --- | --- | --- |
| Vertical slice | #1 | demo animation workflow |
| Health / DB / deploy | #2-#7 | dependency health、database store、migration、Cloud Run deploy 修正 |
| 初期テスト / CI | #8-#16 | API tests、web tests、CI、validation、timeline sync |
| API contract | #17-#19 | error response、validation、API contract |
| Phase 0 / 1 | #30-#33, #38, #39 | roadmap、Project API、Frame API、typed job、contract、Zod response |
| Phase 2 | #34-#37 | R2 client、edited / generated / inpainted image storage |
| Phase 3 | #40 | Replicate inference |
| Phase 4 | #41-#43 | FFmpeg infrastructure、frame splitting、MP4 export |
| Phase 5 | #44-#45 | job state machine、optimistic lock、logging、cleanup |
| Phase 6 | #46-#50 | form validation、editor undo / redo、filters、timeline reorder、export playback |
| Phase 7 | #51-#52, #54 | Playwright 初期 E2E、memory store tests、handler tests |
| Phase 9 / docs | #53, #55 | README 改訂、operations guide |

## 現在残っている主なリスク

- GitHub Issue が open のままで、実装済み / 未実装の境界が Issue 上では分かりにくい
- branch protection がなく、CI 必須化がされていない
- Cloud Run の background goroutine 方式は長時間 AI / FFmpeg 処理で停止リスクがある
- 実 Replicate / R2 / TiDB / Cloud Run / Pages を組み合わせた本番 smoke test が未記録
- Playwright E2E の CI 所要時間と flaky test 対策が未整理
- testcontainers-go による DB 結合テストがない
- shadcn/ui / Radix UI は要件に残っているが実装体系としては未導入
- README に React version と UI 技術スタックの不一致がある
- 推論コストは変動するため、料金目安の取得日と再確認手順が必要

## 調査で確認した主な根拠

- `apps/api/internal/usecase/studio_service.go`: Replicate / demo 切り替え、R2 保存、FFmpeg export、job state 更新
- `apps/api/internal/infrastructure/replicate/`: Replicate client と ToonCrafter / SDXL Inpainting adapter
- `apps/api/internal/infrastructure/media/ffmpeg.go`: MP4 分割と MP4 エンコード
- `apps/api/internal/infrastructure/storage/r2_store.go`: R2 object storage
- `apps/api/internal/infrastructure/db/studio_store.go`: DB-backed store、frame / job / project 永続化、job version lock
- `apps/api/internal/domain/types.go`: job status、result、transition rule
- `apps/web/src/features/frame-generation/components/generation-panel.tsx`: Step 1 UI、React Hook Form + Zod、polling
- `apps/web/src/features/inpainting/components/inpainting-panel.tsx`: Step 2 mask UI、strength、polling
- `apps/web/src/features/editor/components/editor-panel.tsx`: Step 3 editor、undo / redo、filter、save
- `apps/web/src/features/timeline/components/timeline.tsx`: frame kind 表示、drag reorder
- `apps/web/src/features/export/components/export-panel.tsx`: preview、MP4 export、download、video playback
- `apps/web/tests/e2e/workflow.spec.ts`: Playwright 初期 E2E
- `.github/workflows/ci.yml`: web lint / build / test、api test
- `docs/api-contract.md`: HTTP 契約
- `docs/operations-guide.md`: ローカル / 本番動作確認手順
- GitHub PR #1 から #55: すべて merge 済み
- GitHub Actions: 直近 CI は success

## 今回実行した調査コマンド

- `git status --short --branch`
- `git log --oneline --decorate -20`
- `gh auth status`
- `gh repo view --json nameWithOwner,defaultBranchRef,url,sshUrl`
- `gh pr list --state all --limit 80 --json number,title,state,mergedAt,closedAt,headRefName,baseRefName,url`
- `gh issue list --state all --limit 80 --json number,title,state,closedAt,labels,url`
- `gh issue view 20` から `gh issue view 29`
- `gh run list --limit 20 --json databaseId,workflowName,displayTitle,status,conclusion,createdAt,url`
- `gh api repos/Yuji181181/AnifusionCanvas/branches/main/protection`
- `rg --files`
- 主要な API / web / docs ファイルの `sed` 読み取り

## 今回の文書更新に対する評価

- 要件定義書の技術前提は維持した
- 旧ロードマップで未反映だった PR #40 から #55 の実装状況を反映した
- Issue が open のまま残っている事実と、コード上は進んでいる事実を分けて記載した
- 本番検証、CI E2E、DB 結合テスト、UI 技術不一致などの未完了事項を明示した
- コード変更は行わず、ドキュメントだけを更新した
