# Delivery Plan (Vertical Slices)

## Assumptions
- Team: 1-2 engineers.
- Sprint length: 1 week.
- Each sprint ends with a usable release candidate.

## Release Strategy
- Every sprint must include:
  - at least one end-user visible capability,
  - persistence and basic test coverage for new behavior,
  - a deployable static build.

## Sprint 1: Writing MVP Release
User value:
- A writer can open the app, write chapter content, refresh the page, and keep their work.

Backlog items:
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
- WEB-029

Exit criteria:
- Single-project writing workflow works end-to-end.
- Autosave and reload recovery verified.
- Basic shortcuts include save + command palette trigger placeholder.

## Sprint 2: Project and Chapter Management Release
User value:
- A writer can manage chapters and projects, and safely edit with confirmations.

Backlog items:
- WEB-011
- WEB-018
- WEB-019
- WEB-020 (subset: `save`, `open/load`, `chapter new/delete/rename`, `wordcount`)
- WEB-021
- WEB-028
- WEB-030 (subset for chapter/project lifecycle)

Exit criteria:
- Command palette supports core project/chapter workflows.
- Deletion confirms are in place.
- e2e covers chapter lifecycle.

## Sprint 3: Story Planning Release (Notes + Wiki + Templates)
User value:
- A writer can plan scenes and world-building directly in the app.

Backlog items:
- WEB-015
- WEB-016
- WEB-005
- WEB-020 (subset: `wiki new/delete/rename`, `notes`, `structure`)
- WEB-027 (expand shortcuts for notes/wiki flows)
- WEB-029 (add wiki/template unit coverage)
- WEB-030 (add notes/wiki lifecycle coverage)

Exit criteria:
- Notes and wiki are fully usable and persisted.
- Structure templates can be applied from command palette.

## Sprint 4: Editing Intelligence Release
User value:
- A writer gets readability and spellcheck feedback while drafting.

Backlog items:
- WEB-006
- WEB-007
- WEB-022
- WEB-023
- WEB-024
- WEB-020 (subset: `analyze`)
- WEB-030 (add analysis/spellcheck e2e)

Exit criteria:
- Analysis view shows readability + highlighted writing issues.
- Spellcheck returns candidate misspellings reliably.

## Sprint 5: Portability and Workflow Release
User value:
- A writer can move work in/out of the app and use broader command parity.

Backlog items:
- WEB-025
- WEB-026
- WEB-033
- WEB-034
- WEB-020 (subset completion: `export` parity confirmation)
- WEB-029 (file I/O and parser edge-case tests)
- WEB-030 (import/export e2e)

Exit criteria:
- JSON and TXT import/export are production-ready.
- Secondary commands (`search`, `help`, chapter list/reorder decision) shipped.

## Sprint 6: Polish and Public Release
User value:
- A stable, pleasant, documented web app ready for regular use.

Backlog items:
- WEB-031
- WEB-032
- WEB-035
- WEB-036
- WEB-037
- WEB-038

Exit criteria:
- Theme + layout polish complete.
- Deployment docs complete and validated.
- Recovery paths and performance checks in place.

