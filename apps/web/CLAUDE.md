# Frontend Conventions (apps/web)

## Architecture

Pure client-side application. **No Server Actions, no Route Handlers, no `fetch` from Next.js server.**
All API calls go directly to FastAPI via `lib/api.ts`.

## Component Conventions

- **shadcn/ui** components in `components/ui/` — never edit directly; customize via `components.json`
- Feature components in `components/editor/`, `components/timeline/`, `components/forms/`
- Page routes (`app/step1/`, etc.) should be thin wrappers delegating to components

## State Management (Zustand)

- One store per domain in `stores/`: `frame-store.ts`, `job-store.ts`, `editor-store.ts`
- Keep stores flat; prefer multiple small stores over one large store
- Async operations belong in the API layer + useSWR hooks, NOT in stores

## Data Fetching (useSWR)

- useSWR for all reads from FastAPI
- SWR `mutate` for cache invalidation after mutations
- Polling pattern for job status:
  ```ts
  useSWR(() => jobId ? `/api/jobs/${jobId}` : null, fetcher, { refreshInterval: 2000 })
  ```

## Forms (React Hook Form + Zod)

- Define Zod schemas in `types/` alongside TypeScript types
- Use `zodResolver` from `@hookform/resolvers/zod`

## Canvas (Fabric.js)

- **Step 2 (Inpainting):** Drawing mode ONLY. PencilBrush for mask painting. No object manipulation, no filters.
- **Step 3 (Manual Edit):** Full Fabric.js — free drawing, shapes, text, filters, undo/redo, layers
- Never mix AI inference logic into Fabric.js code. Canvas is purely for visual editing.

## Testing

- **Vitest** for: hooks, stores, lib utilities, Zod schemas, non-trivial components
- Mock `lib/api.ts` for any API-dependent hook tests
- Skip pure layout components (no logic = no unit test needed)
- **Playwright** E2E: only the 4-step critical path
- Test file structure mirrors source: `tests/unit/stores/`, `tests/unit/hooks/`, etc.
