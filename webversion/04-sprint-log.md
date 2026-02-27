# Sprint Completion Log

This file tracks completed sprints and what was shipped at the end of each sprint.

## Sprint 1
- Status: Completed
- Completion date: February 27, 2026
- Release theme: Writing MVP

### Outcome
- Usable browser writing app delivered (Web Components + TypeScript).
- Single chapter editing flow works end-to-end.
- Local persistence via `localStorage` with autosave behavior.
- Reload restores saved writing state.
- TUI-inspired visual styling applied.

### Included backlog items
- WEB-001
- WEB-002
- WEB-003
- WEB-004
- WEB-008
- WEB-009
- WEB-010
- WEB-012
- WEB-013
- WEB-014
- WEB-017
- WEB-027
- WEB-029A
- WEB-030A

### Validation snapshot
- `npm run lint`: pass
- `npm run typecheck`: pass
- `npm run test`: pass
- `npm run build`: pass

### Notes
- Typewriter Mode padding-based centering experiment was reverted on February 27, 2026.
- Current behavior remains scroll-based typewriter centering logic.

## Sprint 2
- Status: Completed
- Completion date: February 27, 2026
- Release theme: Project and Chapter Management

### Outcome
- Command parser and command bus implemented.
- Command palette is now functional (not placeholder).
- Core commands shipped:
  - `save` (supports save-as style project cloning by name)
  - `open` / `load`
  - `chapter new`
  - `chapter rename`
  - `chapter delete` (with confirmation modal)
  - `wordcount`
- Migration layer added for legacy project/settings data compatibility.
- Modal dialog component added for confirmations and information dialogs.
- Project lifecycle and chapter lifecycle validated via e2e.

### Included backlog items
- WEB-011
- WEB-018
- WEB-019
- WEB-020A
- WEB-021
- WEB-028
- WEB-029B
- WEB-030B

### Validation snapshot
- `npm run lint`: pass
- `npm run typecheck`: pass
- `npm run test`: pass
- `npm run build`: pass
- `npm run test:e2e`: pass (3 tests)

### Notes
- Playwright browsers were installed on February 27, 2026 to enable local e2e execution.

## Sprint 3
- Status: Completed
- Completion date: February 27, 2026
- Release theme: Story Planning (Notes + Wiki + Templates + Help)

### Outcome
- Notes view implemented with per-chapter persistence.
- Wiki view implemented (split layout) with entry selection and content editing.
- Structure templates implemented and command-integrated (`3act`, `hero`, `cat`, `fichtean`, `horror`).
- Planning command set shipped:
  - `notes`
  - `wiki`
  - `wiki new`
  - `wiki rename`
  - `wiki delete`
  - `structure <name>`
- Shortcut expansion shipped:
  - `Ctrl/Cmd+N` toggle notes view
  - `Ctrl/Cmd+W` toggle wiki view
- In-app help command and registry fully integrated, now covering all shipped Sprint 1-3 capabilities.

### Included backlog items
- WEB-015
- WEB-016
- WEB-005
- WEB-020B
- WEB-027 (expanded)
- WEB-029C
- WEB-030C
- WEB-039

### Validation snapshot
- `npm run lint`: pass
- `npm run typecheck`: pass
- `npm run test`: pass (23 tests)
- `npm run build`: pass
- `npm run test:e2e`: pass (7 tests)
