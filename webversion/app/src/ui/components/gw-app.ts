import type { AppState } from "@/domain/models";
import { AppStore } from "@/app/store";
import { appReducer } from "@/app/reducer";
import { createInitialState } from "@/app/initial-state";
import { LocalStorageRepository } from "@/infra/local-storage-repo";
import { createAutosaveController, type AutosaveController } from "@/app/autosave";
import { GwEditor } from "./gw-editor";
import { GwStatusBar } from "./gw-status-bar";

export class GwApp extends HTMLElement {
  private readonly root = this.attachShadow({ mode: "open" });
  private readonly repo = new LocalStorageRepository();
  private readonly store = new AppStore(createInitialState(), appReducer);
  private autosave: AutosaveController | null = null;
  private unsubscribe: (() => void) | null = null;

  connectedCallback(): void {
    this.renderShell();

    this.loadFromStorage();

    const editor = this.root.querySelector("gw-editor") as GwEditor;
    editor.setEvents({
      onChange: (value) => {
        const chapterId = this.store.getState().currentChapterId;
        this.store.dispatch({ type: "SET_CHAPTER_CONTENT", chapterId, content: value });
      },
      onCursorChange: (row, column) => {
        this.store.dispatch({ type: "SET_CURSOR", cursor: { row, column } });
      }
    });

    this.unsubscribe = this.store.subscribe((state) => {
      this.renderState(state);
    });

    this.autosave = createAutosaveController({
      store: this.store,
      repo: this.repo
    });

    this.registerShortcuts();
    this.renderState(this.store.getState());
    editor.focusEditor();
  }

  disconnectedCallback(): void {
    this.autosave?.stop();
    this.unsubscribe?.();
  }

  private renderShell(): void {
    this.root.innerHTML = `
      <style>
        :host {
          display: block;
          min-height: 100vh;
          color: var(--tui-fg-main);
          padding: 18px;
          box-sizing: border-box;
          font-family: var(--tui-font);
        }

        .container {
          max-width: 1060px;
          margin: 0 auto;
          display: grid;
          gap: 10px;
        }

        .terminal-head {
          border: 1px solid var(--tui-border-strong);
          background: linear-gradient(180deg, rgb(11 23 16 / 95%) 0%, rgb(8 17 11 / 95%) 100%);
          padding: 10px 12px;
        }

        .title-row {
          display: flex;
          justify-content: space-between;
          align-items: center;
          gap: 8px;
          color: var(--tui-fg-main);
          font-size: 13px;
        }

        .title {
          margin: 0;
          font-size: 14px;
          text-transform: uppercase;
          letter-spacing: 0.08em;
        }

        .meta {
          color: var(--tui-fg-soft);
          font-size: 12px;
          white-space: nowrap;
        }

        .hint {
          margin: 8px 0 0;
          color: var(--tui-fg-dim);
          font-size: 12px;
        }

        .hint strong {
          color: var(--tui-warning);
          font-weight: 500;
        }
      </style>
      <main class="container">
        <section class="terminal-head">
          <div class="title-row">
            <h1 class="title">gowrite :: terminal mode</h1>
            <span class="meta">LOCAL STORAGE SESSION</span>
          </div>
          <p class="hint">[<strong>CTRL/CMD+S</strong>] save now  [<strong>CTRL/CMD+K</strong>] command palette (sprint 2)</p>
        </section>
        <gw-editor></gw-editor>
        <gw-status-bar></gw-status-bar>
      </main>
    `;
  }

  private renderState(state: AppState): void {
    const editor = this.root.querySelector("gw-editor") as GwEditor;
    const statusBar = this.root.querySelector("gw-status-bar") as GwStatusBar;
    editor.setTypewriterMode(state.settings.typewriterMode);
    editor.render(state);
    statusBar.render(state);
  }

  private loadFromStorage(): void {
    const project = this.repo.loadCurrentProject();
    if (!project) {
      const initial = this.store.getState();
      this.repo.saveProject(initial.project);
      this.repo.saveSettings(initial.settings);
      return;
    }

    const settings = this.repo.loadSettings();
    const currentSettings = this.store.getState().settings;
    const normalizedSettings = {
      ...currentSettings,
      ...settings
    };

    if (!settings) {
      this.store.dispatch({
        type: "INIT_APP",
        project,
        settings: currentSettings
      });
      return;
    }

    this.store.dispatch({ type: "INIT_APP", project, settings: normalizedSettings });
  }

  private registerShortcuts(): void {
    window.addEventListener("keydown", (event) => {
      const isMeta = event.ctrlKey || event.metaKey;
      if (!isMeta) {
        return;
      }

      if (event.key.toLowerCase() === "s") {
        event.preventDefault();
        this.autosave?.saveNow();
      }

      if (event.key.toLowerCase() === "k") {
        event.preventDefault();
        this.store.dispatch({ type: "TOGGLE_COMMAND_PALETTE" });
      }
    });
  }
}

customElements.define("gw-app", GwApp);
