# Gowrite Web Port Overview

## Goal
Port the current Go TUI writing app to a browser-only app with:
- TypeScript
- Web Components (Custom Elements, no React)
- `localStorage` persistence
- No backend

## Scope
- Feature parity for core authoring workflows:
  - Chapters (create, rename, delete, reorder if added)
  - Per-chapter notes
  - Story wiki entries
  - Command palette with mapped commands
  - Structure templates
  - Hemingway analysis + readability
  - Spellcheck using `dictionary.txt`
  - Project save/load in browser storage
  - JSON import/export and TXT import/export
  - Autosave behavior

## Non-Goals (Initial Release)
- Multi-user collaboration
- Cloud sync
- Server APIs
- Rich text formatting
- React/Vue/Svelte framework adoption

## Constraints
- No backend services.
- Use modular TypeScript.
- Use Web Components for UI composition.

