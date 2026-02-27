# Implementation Backlog

## Working Agreements
- Architecture style: modular TypeScript + Web Components only.
- No backend calls.
- Every task includes tests unless explicitly marked `Spike`.
- Merge criteria: lint, unit tests, and relevant e2e tests pass.

## Definition of Done
- Feature works in Chrome, Firefox, and Safari latest stable.
- Behavior aligns with current Go app intent for mapped feature.
- State survives refresh via `localStorage`.
- Keyboard interactions are accessible and documented.
- Unit and/or e2e coverage added for changed behavior.

## Priority Backlog

| ID | Priority | Epic | Task | Depends On | Estimate | Acceptance Criteria |
|---|---|---|---|---|---|---|
| WEB-001 | P0 | Foundation | Initialize `webversion/app` with Vite + TS strict + ESLint + Prettier | None | S | `npm run dev`, `npm run build`, `npm run test` succeed in clean clone |
| WEB-002 | P0 | Foundation | Create directory architecture (`domain/app/infra/ui`) and barrel exports policy | WEB-001 | S | Project structure matches architecture doc and compiles |
| WEB-003 | P0 | Foundation | Add CI workflow for lint, typecheck, unit tests, build | WEB-001 | S | CI runs on PR and fails on errors |
| WEB-004 | P0 | Domain | Implement core types (`Chapter`, `WikiEntry`, `Project`, `Settings`, `View`) | WEB-002 | S | Type definitions published and imported by all modules |
| WEB-005 | P0 | Domain | Port structure templates (3act, hero, cat, fichtean, horror) | WEB-004 | S | Template apply returns chapter arrays matching Go template names |
| WEB-006 | P0 | Domain | Port readability calculation (ARI) | WEB-004 | S | Unit tests cover baseline cases from Go tests |
| WEB-007 | P0 | Domain | Port Hemingway analysis tagging | WEB-004 | S | Unit tests assert adverb/passive/long sentence tagging |
| WEB-008 | P0 | App Core | Build app store with reducer/actions/selectors | WEB-004 | M | Store supports chapter/wiki/view/settings updates without direct mutation |
| WEB-009 | P0 | App Core | Implement dirty-state tracking + transaction-safe updates | WEB-008 | M | Dirty flag toggles correctly after edits and clears after save |
| WEB-010 | P0 | Persistence | Implement `localStorage` repository (CRUD project + settings) | WEB-004 | M | Save/load/delete/list work with stable key scheme |
| WEB-011 | P0 | Persistence | Implement schema migration layer for versioned project data | WEB-010 | M | Old schema loads and is upgraded in-memory without crash |
| WEB-012 | P0 | Persistence | Wire autosave (debounced + heartbeat every 60s when dirty) | WEB-009, WEB-010 | M | Editing triggers autosave, refresh restores latest saved state |
| WEB-013 | P0 | UI Shell | Create `<gw-app>` root component and store subscription loop | WEB-008 | M | Root renders active view and reacts to store updates |
| WEB-014 | P0 | UI Shell | Create `<gw-editor>` with chapter text area | WEB-013 | M | Typing edits current chapter content and updates word stats |
| WEB-015 | P0 | UI Shell | Create `<gw-notes>` for chapter notes view | WEB-013 | S | Notes edit path persists per chapter |
| WEB-016 | P0 | UI Shell | Create `<gw-wiki>` list + editor split panel | WEB-013 | M | Wiki entries selectable/editable and persisted |
| WEB-017 | P0 | UI Shell | Create `<gw-status-bar>` with word count and cursor info | WEB-014 | S | Status updates as cursor and text change |
| WEB-018 | P0 | Commands | Implement command parser (`command + args`) | WEB-008 | S | Parser handles commands and quoted args consistently |
| WEB-019 | P0 | Commands | Implement command bus/registry pattern | WEB-018 | M | Commands register independently and dispatch typed actions |
| WEB-020 | P0 | Commands | Port P0 commands: `save`, `open/load`, `export`, `structure`, `chapter`, `wiki`, `notes`, `analyze`, `wordcount` | WEB-019, WEB-010 | L | Each command has acceptance tests and UI feedback |
| WEB-021 | P0 | UI Shell | Build `<gw-command-palette>` with open/close/focus flow | WEB-019, WEB-013 | M | Palette toggles via shortcut, executes commands, restores focus |
| WEB-022 | P0 | Features | Build analysis view component `<gw-analysis>` | WEB-006, WEB-007, WEB-013 | M | Analysis renders highlighting and readability summary |
| WEB-023 | P0 | Features | Implement dictionary loader from bundled `dictionary.txt` | WEB-001 | M | Dictionary loads once, errors handled gracefully |
| WEB-024 | P0 | Features | Port spellcheck command using dictionary module | WEB-023, WEB-019 | M | Unknown words list shown with cap + overflow message |
| WEB-025 | P0 | File I/O | Implement JSON import/export via File API and download blob | WEB-010 | M | User can export project JSON and import valid JSON |
| WEB-026 | P0 | File I/O | Implement TXT import to current/new chapter | WEB-014, WEB-025 | M | Import modes match Go behavior and show result messaging |
| WEB-027 | P0 | UX | Implement global shortcuts (`Ctrl/Cmd+S`, palette, notes/wiki toggle, help) | WEB-013, WEB-021 | M | Shortcut map works on desktop and does not break browser essentials |
| WEB-028 | P0 | UX | Add modal/dialog component for errors, confirms, success | WEB-013 | S | Confirm delete flows block until explicit user action |
| WEB-029 | P0 | Testing | Add unit tests for domain + parser + migrations + storage adapter | WEB-011, WEB-019 | M | Coverage includes failure paths and legacy schema load |
| WEB-030 | P0 | Testing | Add Playwright e2e for chapter/wiki lifecycle + autosave recovery | WEB-027, WEB-028 | L | End-to-end suite verifies authoring, reload persistence, command flows |
| WEB-031 | P1 | UX | Implement theme system (`dark`, `light`, `retro`) with CSS tokens | WEB-013 | M | Theme command updates tokens and persists selected theme |
| WEB-032 | P1 | UX | Implement centered mode and focus mode layout toggles | WEB-014, WEB-015 | M | Modes visually match intent and persist in settings |
| WEB-033 | P1 | Commands | Port `search`, `help`, chapter list modal parity | WEB-021 | M | Commands available and covered by tests |
| WEB-034 | P1 | Commands | Evaluate and implement chapter reorder command parity | WEB-020 | S | Either implemented with tests or explicitly removed/documented |
| WEB-035 | P1 | Docs | Add user docs for browser usage, shortcuts, import/export | WEB-030 | S | `webversion/docs` includes usage and troubleshooting |
| WEB-036 | P1 | Release | Package as static app and document deployment options (GitHub Pages/Netlify) | WEB-030 | S | Deployment guide tested on one hosted environment |
| WEB-037 | P2 | Quality | Add performance budget and typing latency benchmark | WEB-014 | S | Typing remains responsive for large chapter text |
| WEB-038 | P2 | Quality | Add corruption recovery UX for invalid localStorage payloads | WEB-011 | S | App can recover/reset gracefully with user confirmation |

## Command Parity Checklist

| Command | Status Target | Backlog Item |
|---|---|---|
| `save` | P0 | WEB-020 |
| `open` / `load` | P0 | WEB-020 |
| `export` | P0 | WEB-020, WEB-025 |
| `structure <name>` | P0 | WEB-005, WEB-020 |
| `chapter new/delete/rename` | P0 | WEB-020 |
| `wiki new/delete/rename` | P0 | WEB-020 |
| `notes` | P0 | WEB-020 |
| `analyze` | P0 | WEB-022 |
| `spellcheck` | P0 | WEB-024 |
| `wordcount` | P0 | WEB-020 |
| `import` / `import new` | P0 | WEB-026 |
| `search` | P1 | WEB-033 |
| `help` | P1 | WEB-033 |
| `theme` | P1 | WEB-031 |

## Execution Order (Updated)
1. Execute by `Vertical Sprint Mapping` below (this is the primary sequencing).
2. Use task dependencies in the backlog table to order work within each sprint.

## Vertical Sprint Mapping

| Sprint | Release Theme | Backlog IDs |
|---|---|---|
| Sprint 1 | Writing MVP | WEB-001, WEB-002, WEB-003, WEB-004, WEB-008, WEB-009, WEB-010, WEB-012, WEB-013, WEB-014, WEB-017, WEB-027, WEB-029A, WEB-030A |
| Sprint 2 | Project + Chapter Management | WEB-011, WEB-018, WEB-019, WEB-020A, WEB-021, WEB-028, WEB-029B, WEB-030B |
| Sprint 3 | Notes + Wiki + Templates | WEB-015, WEB-016, WEB-005, WEB-020B, WEB-027 (expanded), WEB-029C, WEB-030C |
| Sprint 4 | Analysis + Spellcheck | WEB-006, WEB-007, WEB-022, WEB-023, WEB-024, WEB-020C, WEB-029C, WEB-030D |
| Sprint 5 | Import/Export + Secondary Commands | WEB-025, WEB-026, WEB-033, WEB-034, WEB-020D, WEB-020E, WEB-029D, WEB-030E |
| Sprint 6 | Polish + Release | WEB-031, WEB-032, WEB-035, WEB-036, WEB-037, WEB-038 |

## Child Tasks: WEB-020 (Commands)

| ID | Parent | Sprint | Task | Depends On | Estimate | Acceptance Criteria |
|---|---|---|---|---|---|---|
| WEB-020A | WEB-020 | Sprint 2 | Implement core command handlers: `save`, `open/load`, `chapter new/delete/rename`, `wordcount` | WEB-019, WEB-010, WEB-021, WEB-028 | M | Commands execute via registry, update store correctly, and display success/error feedback |
| WEB-020B | WEB-020 | Sprint 3 | Implement planning command handlers: `wiki new/delete/rename`, `notes`, `structure <name>` | WEB-020A, WEB-015, WEB-016, WEB-005 | M | Commands mutate wiki/view/template state correctly and preserve focus behavior |
| WEB-020C | WEB-020 | Sprint 4 | Implement `analyze` command integration to open analysis view with current text | WEB-022, WEB-020A | S | `analyze` switches view, renders metrics, and returns to editing mode cleanly |
| WEB-020D | WEB-020 | Sprint 5 | Implement command integration for import/export workflows (`export`, `import`, `import new`) | WEB-025, WEB-026, WEB-020A | M | File commands trigger browser file APIs and user-visible completion/error messages |
| WEB-020E | WEB-020 | Sprint 5 | Implement secondary commands: `search`, `help`, chapter list modal entry points | WEB-033, WEB-020A | S | Secondary commands are discoverable in palette and behave consistently with docs |
| WEB-020F | WEB-020 | Sprint 6 | Command parity cleanup and deprecation handling for any intentionally unsupported commands | WEB-020A, WEB-020B, WEB-020C, WEB-020D, WEB-020E | S | Command catalog documented, unsupported commands return actionable guidance, all command tests pass |

## Child Tasks: WEB-029 (Unit/Integration Testing)

| ID | Parent | Sprint | Task | Depends On | Estimate | Acceptance Criteria |
|---|---|---|---|---|---|---|
| WEB-029A | WEB-029 | Sprint 1 | Add unit tests for models, reducers/selectors, and localStorage adapter happy paths | WEB-004, WEB-008, WEB-010 | M | Core editing and persistence logic has deterministic tests and no flaky cases |
| WEB-029B | WEB-029 | Sprint 2 | Add parser/command/migration tests for core command set and schema upgrade paths | WEB-011, WEB-018, WEB-019, WEB-020A | M | Parser and migration edge cases covered; command handler tests validate store side-effects |
| WEB-029C | WEB-029 | Sprint 3-4 | Add tests for wiki/notes/templates plus analysis/spellcheck modules | WEB-020B, WEB-022, WEB-024 | M | Planning and intelligence features have unit tests for normal and edge inputs |
| WEB-029D | WEB-029 | Sprint 5 | Add file I/O integration tests and command regression suite for import/export | WEB-020D, WEB-025, WEB-026 | M | Import/export workflows validated against fixture files and failure conditions |

## Child Tasks: WEB-030 (Playwright E2E)

| ID | Parent | Sprint | Task | Depends On | Estimate | Acceptance Criteria |
|---|---|---|---|---|---|---|
| WEB-030A | WEB-030 | Sprint 1 | E2E baseline: launch app, edit text, autosave, refresh recovery | WEB-014, WEB-012 | M | Test proves user text survives reload and app remains interactive |
| WEB-030B | WEB-030 | Sprint 2 | E2E chapter/project lifecycle: create/rename/delete chapter, open/load project | WEB-020A, WEB-021, WEB-028 | M | Lifecycle flows succeed with confirmation dialog behavior validated |
| WEB-030C | WEB-030 | Sprint 3 | E2E planning lifecycle: notes editing, wiki CRUD, structure apply | WEB-020B, WEB-015, WEB-016 | M | Planning features validated through command palette and direct UI interactions |
| WEB-030D | WEB-030 | Sprint 4 | E2E intelligence flows: analyze view and spellcheck command results | WEB-020C, WEB-024 | M | Analysis/spellcheck UX and outputs are verified end-to-end |
| WEB-030E | WEB-030 | Sprint 5 | E2E portability flows: JSON/TXT import/export and cross-session recovery smoke | WEB-020D, WEB-025, WEB-026 | L | Import/export files are usable and app remains stable after repeated round trips |
