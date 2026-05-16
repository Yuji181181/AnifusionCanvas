# Codex Project Instructions

このリポジトリでは `.commandcode/` の AI agent 設定を Codex の開発ルールとして扱う。
作業前にこのファイルを起点として、必要な `.commandcode` ファイルを読むこと。

## Language And Communication

- ユーザーとのやり取り、計画、実装報告、レビューは日本語で行う。
- 不明点は前提・仮説・未確定事項として明示する。
- 変更前に既存ファイルを読み、読んでいないファイルは編集しない。

## Required Workflow

`.commandcode/runtime/orchestration.md` の方針に従い、原則として次のループで作業する。

1. `planner`: 要件を解釈し、影響範囲、非変更条件、タスク、受け入れ基準、検証計画を決める。
2. `generater`: 既存規約に沿って必要最小限の実装・編集・テスト追加・設定変更を行う。
3. `evaluer`: 要件充足、設計整合、テスト妥当性、運用リスク、変更漏れを評価する。

複数ファイル変更、新機能、構成変更、設計変更は必ず planner から開始する。
単純修正でも、完了前に evaluer 観点で自己点検する。

## Source Files For Agent Rules

- Role definitions:
  - `.commandcode/agents/planner.md`
  - `.commandcode/agents/generater.md`
  - `.commandcode/agents/evaluer.md`
- Runtime rules:
  - `.commandcode/runtime/orchestration.md`
  - `.commandcode/runtime/quality-gates.md`
  - `.commandcode/runtime/task-contract.md`
  - `.commandcode/runtime/usage-guide.md`
  - `.commandcode/runtime/sub-agents.yaml`
- Reusable procedures:
  - `.commandcode/skills/requirements-planning.md`
  - `.commandcode/skills/frontend-spa-implementation.md`
  - `.commandcode/skills/backend-go-implementation.md`
  - `.commandcode/skills/documentation-rewrite.md`
  - `.commandcode/skills/evaluation-review.md`
- User preferences:
  - `.commandcode/taste/taste.md`
- Templates:
  - `.commandcode/templates/plan-template.md`
  - `.commandcode/templates/implementation-template.md`
  - `.commandcode/templates/evaluation-template.md`

## Project Constraints

- Frontend: Vite + React SPA を前提にする。SSR 前提の実装は持ち込まない。
- Frontend libraries: TanStack Router, TanStack Query, Zustand, TailwindCSS, shadcn/ui, React Hook Form + Zod, Fabric.js の前提を不要に崩さない。
- Backend: Go + Echo を前提にする。
- Backend architecture: HTTP ハンドラとユースケースを分離し、外部 I/O は infrastructure に閉じ込める。
- Infrastructure: Cloud Run 前提を壊さず、CLI と自動化を優先する。
- Documentation: 旧技術名の取り残し、本文・表・図の用語不一致を必ず確認する。

## Quality Gates

完了前に `.commandcode/runtime/quality-gates.md` を基準として確認する。

- 非変更条件を破っていない。
- 依頼範囲に対して変更漏れがない。
- 追加した技術選定が現行アーキテクチャと整合している。
- ドキュメントとコードが矛盾していない。
- 実行可能な検証を行い、実行できない場合は理由を明示している。

## Pull Requests

- GitHub CLI は `gh` を使用する。
- Git protocol は SSH を優先する。
- 開発中は、要件、機能、修正、設定、文書などの意味のある小さな単位でコミットする。
- 1つのコミットには1つの責務だけを含め、無関係な変更を混ぜない。
- PR 作成前に `git status --short --branch` と関連する検証コマンドを確認する。
- 作業ブランチを push し、`gh pr create` で pull request を作成する。
- pull request には目的、主な変更、検証結果、未解決事項を記載する。
- CI、レビュー、必要な検証が通ったら `gh pr merge` で merge する。
- `gh` が未認証、CI が失敗、競合がある、または merge 権限がない場合は、状態と次アクションを明示して停止する。
