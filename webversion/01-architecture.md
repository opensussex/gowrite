# Architecture Blueprint (TypeScript + Web Components)

## Proposed Stack
- Build: Vite
- Language: TypeScript (strict mode)
- Testing: Vitest (unit), Playwright (e2e)
- UI: Native Web Components + Shadow DOM
- Storage: Browser `localStorage`

## Module Boundaries
```text
src/
  domain/
    models.ts
    templates.ts
    readability.ts
    hemingway.ts
    spellcheck.ts
    validation.ts
  app/
    store.ts
    actions.ts
    reducers.ts
    selectors.ts
    command-bus.ts
    command-parser.ts
    shortcuts.ts
    autosave.ts
  infra/
    local-storage-repo.ts
    file-io.ts
    dictionary-loader.ts
    migrations.ts
  ui/
    components/
      gw-app.ts
      gw-editor.ts
      gw-notes.ts
      gw-wiki.ts
      gw-command-palette.ts
      gw-analysis.ts
      gw-status-bar.ts
      gw-modal.ts
    styles/
      tokens.css
      app.css
  main.ts
```

## Data Contracts
- `Chapter`: `{ id, title, content, notes, target }`
- `WikiEntry`: `{ id, title, content }`
- `Project`: `{ id, name, chapters, wiki, createdAt, updatedAt, schemaVersion }`
- `AppSettings`: `{ theme, centered, focusMode, lastProjectId }`
- `View`: `main | notes | analyze | wiki`

## Persistence Strategy
- Keys:
  - `gowrite.settings`
  - `gowrite.currentProjectId`
  - `gowrite.project.<id>`
  - `gowrite.projectIndex`
- Write path:
  - Immediate in-memory update
  - Debounced localStorage write (e.g. 800ms)
  - Periodic autosave heartbeat (60s) when dirty
- Migration path:
  - Read schema version
  - Upgrade older structures
  - Backfill missing wiki/chapter defaults

## UI Composition
- `gw-app` owns store subscription and route/view state.
- Child components emit typed custom events (`chapter-created`, `command-submitted`).
- Store actions are the only mutation path.
- Components receive state snapshots and render idempotently.

## Command Model
- Parser splits command + args.
- Registry maps command names to handlers.
- Handlers call use-case functions, never directly mutate DOM.

## Key Quality Gates
- No circular imports across `domain`, `app`, `infra`, `ui`.
- Domain logic pure and unit tested.
- Local storage failures handled gracefully.
- Keyboard navigation and focus management verified via e2e.

