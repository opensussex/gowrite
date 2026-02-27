# Session Handoff Template

Use this template to restart work cleanly in a new session.

## 1. Pre-Session Checks
- Open:
  - `webversion/04-sprint-log.md`
  - `webversion/03-delivery-plan.md`
  - `webversion/02-implementation-backlog.md`
- Confirm current sprint status from sprint log.
- Confirm next backlog items from vertical sprint mapping.

## 2. Local Validation (Before New Work)
Run:

```bash
cd webversion/app
npm run lint && npm run typecheck && npm run test && npm run build
```

If any command fails, fix that first before new feature work.

## 3. Copy/Paste Prompt For New Session

```text
Continue work on gowrite web port.

Context:
- Repo root: /Users/graham.holden/Development/programming_exploration/go/gowrite
- Planning docs:
  - webversion/02-implementation-backlog.md
  - webversion/03-delivery-plan.md
  - webversion/04-sprint-log.md
- Current status: Sprint [N] [in progress/completed].
- Last completed backlog IDs: [list]
- Next target backlog IDs: [list]

Instructions:
1) Implement the next backlog IDs for the current sprint as vertical slices.
2) Keep architecture modular TypeScript + Web Components (no React, no backend).
3) Run lint, typecheck, tests, and build after changes.
4) Update webversion/04-sprint-log.md with progress and completion notes.
5) Summarize changed files and validation output.
```

## 4. End-Of-Session Checklist
- Update `webversion/04-sprint-log.md`:
  - what was completed,
  - exact backlog IDs,
  - validation results,
  - open risks/blockers.
- If sprint scope changed, update:
  - `webversion/03-delivery-plan.md`
  - `webversion/02-implementation-backlog.md`
- Ensure local checks pass.
- Provide a short “next session starts with …” note.

