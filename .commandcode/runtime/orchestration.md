# Orchestration

## Default Loop

1. planner がタスクを解釈して分解する
2. generater が最小変更で実装する
3. evaluer が受け入れ判定する
4. fail の場合は planner が差し戻し内容を再構造化する
5. pass までループする

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
