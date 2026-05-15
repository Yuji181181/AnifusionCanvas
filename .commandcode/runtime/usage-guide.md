# Usage Guide

## Recommended Operating Pattern

### 1. planner

入力:
- 依頼内容
- 対象ドキュメント/コード
- 制約条件

出力:
- タスク分解
- 変更範囲
- 非変更条件
- 検証計画

### 2. generater

入力:
- planner のタスク分解
- 対象ファイル
- 実装制約

出力:
- 変更済みファイル
- 実装メモ
- 実行した検証

### 3. evaluer

入力:
- planner の受け入れ基準
- generater の変更結果

出力:
- pass / fail / conditional-pass
- 指摘事項
- 次アクション

## Autonomous Task Loop

- planner は常に最小実装単位までタスクを落とす
- generater は1つの実装単位ごとに検証する
- evaluer は明確な根拠を持って判定する
- fail の場合は planner が差分の学習を行い、再計画する

## For This Project

- フロントエンドは Vite+ SPA 前提
- バックエンドは Go + Echo 前提
- AIモデル選定と3ステップ編集ワークフローは固定条件
- ドキュメント更新では旧技術名の取り残しを必ず監査する
