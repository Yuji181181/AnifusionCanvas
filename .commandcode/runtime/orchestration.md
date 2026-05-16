# Orchestration

## Default Loop

1. planner がタスクを解釈して分解する
2. generater が最小変更で実装する
3. evaluer が受け入れ判定する
4. fail の場合は planner が差し戻し内容を再構造化する
5. pass までループする
6. pass した最小単位ごとにコミットする
7. 作業単位が完了したら pull request を作成する
8. pull request の検証が通ったら merge する

## Git And Pull Request Flow

- 変更は要件、機能、修正、設定、文書などの意味のある小さな単位でコミットする
- 1つのコミットには1つの責務だけを含め、無関係な変更を混ぜない
- コミット前に `git status --short --branch` と関連する検証コマンドを確認する
- コミットメッセージは変更内容が追跡できる具体的な文にする
- 作業ブランチを push し、`gh pr create` で pull request を作成する
- pull request には目的、主な変更、検証結果、未解決事項を記載する
- CI、レビュー、必要な検証が通ったら `gh pr merge` で merge する
- merge 後はローカルブランチとリモートブランチの整理を検討する
- `gh` が未認証、CI が失敗、競合がある、または merge 権限がない場合は、状態と次アクションを明示して停止する

## Routing Rules

- 新機能、複数ファイル変更、構成変更、設計変更は planner から開始する
- 単純修正でも evaluer は最低限の整合性確認を行う
- 文書変更のみでも evaluer は用語整合と変更漏れを確認する

## Escalation Rules

- 要件矛盾: planner に戻す
- 実装不整合: generater に戻す
- 受け入れ基準不足: planner に戻す
- テスト不足: generater に追加実装を依頼する

## Required Artifacts

- plan.md または同等のタスク計画
- implementation-report.md または同等の変更記録
- evaluation-report.md または同等の判定記録
