# Quality Gates

## Global Gates

- 非変更条件を破っていない
- 依頼された範囲に対して変更漏れがない
- 追加した技術選定が現行アーキテクチャに整合している
- ドキュメントとコードの記述が矛盾していない
- 変更が意味のある小さな単位でコミットされている、または未コミット理由が明示されている
- pull request 作成、検証、merge の状態が明示されている
- merge できない場合は、権限、認証、CI、競合などの阻害要因と次アクションが明示されている

## Frontend Gates

- Vite+ SPA 前提を破っていない
- TanStack Router / TanStack Query 前提に整合している
- Zustand, TailwindCSS, shadcn/ui, React Hook Form + Zod, Fabric.js を不要に崩していない

## Backend Gates

- Go + Echo 前提を破っていない
- Cloud Run, sqlc, golang-migrate, AWS SDK for Go v2, FFmpeg, Goroutine/Channel 前提に整合している
- Python / FastAPI 依存に逆戻りしていない

## Documentation Gates

- 用語が統一されている
- 旧技術名が残っていない
- 図、表、本文が同じ前提を共有している
