# Skill: frontend-spa-implementation

## Goal

Vite+ / React SPA, TanStack Router, TanStack Query 前提でフロントエンドを実装する。

## Rules

- SSR前提の実装を持ち込まない
- 画面責務は routes に、機能責務は features に置く
- API呼び出しは query / mutation として整理する
- 状態の永続化や共有が必要なものだけ Zustand に置く
- フォームは React Hook Form + Zod を優先する

## Checklist

- ルート設計
- Query key 設計
- エラーハンドリング
- ローディング状態
- 楽観更新の要否
- テスト方針
