# Storage Design

AnifusionCanvas は生成結果、マスク、手動編集画像、書き出し動画を Cloudflare R2 に保存する。

## Object Key Policy

R2 object key は project 単位で分け、用途ごとの prefix を固定する。

| Purpose | Prefix |
| --- | --- |
| User uploaded keyframes | `projects/{projectId}/inputs/` |
| Generated / edited frames | `projects/{projectId}/frames/` |
| Inpainting masks | `projects/{projectId}/masks/` |
| Exported videos | `projects/{projectId}/exports/` |

Rules:

- Keys use `/` as a logical separator.
- Keys should include stable IDs rather than user-facing names.
- File extensions should match the encoded content type when known.
- Future cleanup jobs can delete by `projects/{projectId}/` prefix.

## URL Policy

`R2_PUBLIC_BASE_URL` controls the URL stored in API results.

- When `R2_PUBLIC_BASE_URL` is set, stored object URLs are `${R2_PUBLIC_BASE_URL}/${escapedKey}`.
- When `R2_PUBLIC_BASE_URL` is empty, stored object URLs use `r2://{bucket}/{key}`.
- `r2://` URLs are internal object references. They are not browser-readable without a future signed URL or proxy endpoint.

The current implementation intentionally does not mint signed URLs. Public asset delivery can be enabled with a Cloudflare custom domain by setting `R2_PUBLIC_BASE_URL`.

## Current Implementation

- `apps/api/internal/infrastructure/storage.R2Store` wraps AWS SDK for Go v2 S3-compatible calls.
- `PutDataURL` decodes a `data:` URL and uploads it to R2.
- `PutBytes`, `GetObject`, and `DeleteObject` are available for generated files and cleanup flows.
- `StudioService.UpdateFrame` stores manual edits in R2 when R2 is fully configured.
- Local development without R2 credentials keeps the previous data URL behavior.
- Unit tests use a fake S3 client, so CI does not require real R2 credentials.
