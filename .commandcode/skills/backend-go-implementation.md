# Skill: backend-go-implementation

## Goal

Go + Echo, sqlc, golang-migrate, AWS SDK for Go v2, FFmpeg, Goroutine/Channel 前提でバックエンドを実装する。

## Rules

- HTTPハンドラとユースケースを分離する
- SQLは明示的に管理し、sqlc で型生成する
- 外部I/Oは infrastructure に閉じ込める
- 非同期ジョブはメモリ状態に依存しすぎない
- Cloud Run のスケール特性を前提に設計する

## Checklist

- API入出力定義
- DBスキーマ変更
- マイグレーション
- ジョブ状態遷移
- タイムアウト / リトライ
- 結合テスト
