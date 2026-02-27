import type { AppState } from "@/domain/models";
import { AppStore } from "@/app/store";
import { appReducer } from "@/app/reducer";
import { createInitialState } from "@/app/initial-state";
import { LocalStorageRepository } from "@/infra/local-storage-repo";
import { createAutosaveController, type AutosaveController } from "@/app/autosave";
import { GwEditor } from "./gw-editor";
import { GwStatusBar } from "./gw-status-bar";
import { CommandBus, type CommandContext } from "@/app/command-bus";
import { registerCoreCommands } from "@/app/commands/core";
import { GwCommandPalette, type CommandSubmitDetail } from "./gw-command-palette";
import { GwModal } from "./gw-modal";
import { getCurrentChapterIndex, getCurrentWikiIndex } from "@/app/selectors";
import { GwNotes } from "./gw-notes";
import { GwWiki } from "./gw-wiki";

export class GwApp extends HTMLElement {
  private readonly root = this.attachShadow({ mode: "open" });
  private readonly repo = new LocalStorageRepository();
  private readonly store = new AppStore(createInitialState(), appReducer);
  private readonly commandBus = new CommandBus();
  private autosave: AutosaveController | null = null;
  private unsubscribe: (() => void) | null = null;

  private readonly keydownHandler = (event: KeyboardEvent) => {
    const state = this.store.getState();

    if (event.key === "Escape" && state.commandPaletteOpen) {
      event.preventDefault();
      this.store.dispatch({ type: "SET_COMMAND_PALETTE_OPEN", open: false });
      this.focusActivePane();
      return;
    }

    const isMeta = event.ctrlKey || event.metaKey;
    if (!isMeta) {
      return;
    }

    const key = event.key.toLowerCase();

    if (key === "s") {
      event.preventDefault();
      this.persistCurrentProject(true);
      return;
    }

    if (key === "k") {
      event.preventDefault();
      const nextOpen = !this.store.getState().commandPaletteOpen;
      this.store.dispatch({ type: "SET_COMMAND_PALETTE_OPEN", open: nextOpen });
      if (!nextOpen) {
        this.focusActivePane();
      }
      return;
    }

    if (key === "n") {
      event.preventDefault();
      const nextView = state.view === "notes" ? "main" : "notes";
      this.store.dispatch({ type: "SET_VIEW", view: nextView });
      this.store.dispatch({ type: "SET_STATUS", message: nextView === "notes" ? "Notes view opened" : "Main editor opened" });
      this.focusActivePane();
      return;
    }

    if (key === "w") {
      event.preventDefault();
      const nextView = state.view === "wiki" ? "main" : "wiki";
      this.store.dispatch({ type: "SET_VIEW", view: nextView });
      this.store.dispatch({ type: "SET_STATUS", message: nextView === "wiki" ? "Wiki view opened" : "Main editor opened" });
      this.focusActivePane();
    }
  };

  connectedCallback(): void {
    this.renderShell();

    registerCoreCommands(this.commandBus);
    this.loadFromStorage();

    const editor = this.getEditor();
    editor.setEvents({
      onChange: (value) => {
        const chapterId = this.store.getState().currentChapterId;
        this.store.dispatch({ type: "SET_CHAPTER_CONTENT", chapterId, content: value });
      },
      onCursorChange: (row, column) => {
        this.store.dispatch({ type: "SET_CURSOR", cursor: { row, column } });
      }
    });

    const notes = this.getNotes();
    notes.setEvents({
      onChange: (value) => {
        const chapterId = this.store.getState().currentChapterId;
        this.store.dispatch({ type: "SET_CHAPTER_NOTES", chapterId, notes: value });
      },
      onCursorChange: (row, column) => {
        this.store.dispatch({ type: "SET_CURSOR", cursor: { row, column } });
      }
    });

    const wiki = this.getWiki();
    wiki.setEvents({
      onSelectEntry: (wikiId) => {
        this.store.dispatch({ type: "SET_CURRENT_WIKI", wikiId });
      },
      onChangeContent: (value) => {
        const wikiId = this.store.getState().currentWikiId;
        this.store.dispatch({ type: "SET_WIKI_CONTENT", wikiId, content: value });
      },
      onCursorChange: (row, column) => {
        this.store.dispatch({ type: "SET_CURSOR", cursor: { row, column } });
      }
    });

    const palette = this.getCommandPalette();
    palette.addEventListener("command-submit", (event) => {
      void this.handleCommandSubmit((event as CustomEvent<CommandSubmitDetail>).detail.command);
    });

    palette.addEventListener("command-cancel", () => {
      this.store.dispatch({ type: "SET_COMMAND_PALETTE_OPEN", open: false });
      this.focusActivePane();
    });

    this.unsubscribe = this.store.subscribe((state) => {
      this.renderState(state);
    });

    this.autosave = createAutosaveController({
      store: this.store,
      repo: this.repo
    });

    window.addEventListener("keydown", this.keydownHandler);
    this.renderState(this.store.getState());
    this.focusActivePane();
  }

  disconnectedCallback(): void {
    this.autosave?.stop();
    this.unsubscribe?.();
    window.removeEventListener("keydown", this.keydownHandler);
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

        .context {
          border: 1px solid var(--tui-border);
          background: rgb(7 13 9 / 78%);
          color: var(--tui-fg-soft);
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.04em;
          padding: 6px 10px;
        }

        .context strong {
          color: var(--tui-accent);
          font-weight: 600;
        }

        .pane {
          display: none;
        }

        .pane.active {
          display: block;
        }
      </style>
      <main class="container">
        <section class="terminal-head">
          <div class="title-row">
            <h1 class="title">gowrite :: terminal mode</h1>
            <span class="meta">LOCAL STORAGE SESSION</span>
          </div>
          <p class="hint">[<strong>CTRL/CMD+K</strong>] command palette [<strong>CTRL/CMD+S</strong>] save [<strong>CTRL/CMD+N</strong>] notes [<strong>CTRL/CMD+W</strong>] wiki [<strong>help</strong>] guide</p>
        </section>

        <section class="context" id="contextRow"></section>

        <section id="paneMain" class="pane">
          <gw-editor></gw-editor>
        </section>
        <section id="paneNotes" class="pane">
          <gw-notes></gw-notes>
        </section>
        <section id="paneWiki" class="pane">
          <gw-wiki></gw-wiki>
        </section>

        <gw-command-palette></gw-command-palette>
        <gw-status-bar></gw-status-bar>
      </main>
      <gw-modal></gw-modal>
    `;
  }

  private renderState(state: AppState): void {
    const editor = this.getEditor();
    const notes = this.getNotes();
    const wiki = this.getWiki();
    const statusBar = this.getStatusBar();
    const palette = this.getCommandPalette();
    const contextRow = this.root.querySelector("#contextRow") as HTMLElement;

    const chapterIndex = getCurrentChapterIndex(state) + 1;
    const chapterCount = state.project.chapters.length;
    const wikiIndex = getCurrentWikiIndex(state) + 1;
    const wikiCount = state.project.wiki.length;

    contextRow.innerHTML = `<strong>PROJECT</strong> ${state.project.name}  |  <strong>VIEW</strong> ${state.view.toUpperCase()}  |  <strong>CHAPTER</strong> ${chapterIndex}/${chapterCount}  |  <strong>WIKI</strong> ${wikiIndex}/${wikiCount}`;

    editor.setTypewriterMode(state.settings.typewriterMode);
    editor.render(state);
    notes.render(state);
    wiki.render(state);
    statusBar.render(state);
    palette.setOpen(state.commandPaletteOpen);

    this.togglePane("paneMain", state.view === "main");
    this.togglePane("paneNotes", state.view === "notes");
    this.togglePane("paneWiki", state.view === "wiki");
  }

  private loadFromStorage(): void {
    const initial = this.store.getState();
    const project = this.repo.loadCurrentProject();

    if (!project) {
      this.repo.saveProject(initial.project);
      this.repo.saveSettings(initial.settings);
      return;
    }

    const settings = this.repo.loadSettings(initial.settings) ?? initial.settings;
    this.store.dispatch({ type: "INIT_APP", project, settings });
  }

  private persistCurrentProject(force = false): void {
    const state = this.store.getState();
    if (!force && !state.dirty) {
      return;
    }

    this.repo.saveProject(state.project);
    this.repo.saveSettings(state.settings);
    this.store.dispatch({ type: "SAVE_SUCCEEDED", timestamp: Date.now() });
  }

  private async handleCommandSubmit(command: string): Promise<void> {
    this.store.dispatch({ type: "SET_COMMAND_PALETTE_OPEN", open: false });

    if (!command.trim()) {
      this.focusActivePane();
      return;
    }

    const result = await this.commandBus.execute(command, this.createCommandContext());
    if (!result.ok) {
      await this.getModal().showInfo("Command Error", result.error ?? "Command failed");
      this.store.dispatch({ type: "SET_STATUS", message: result.error ?? "Command failed" });
      this.focusActivePane();
      return;
    }

    if (result.message) {
      this.store.dispatch({ type: "SET_STATUS", message: result.message });
    }

    this.focusActivePane();
  }

  private createCommandContext(): CommandContext {
    return {
      getState: () => this.store.getState(),
      dispatch: (action) => this.store.dispatch(action),
      repo: this.repo,
      persistCurrentProject: (force = false) => this.persistCurrentProject(force),
      showModal: async (title, message) => this.getModal().showInfo(title, message),
      confirm: async (title, message) => this.getModal().confirm(title, message)
    };
  }

  private togglePane(id: string, active: boolean): void {
    const pane = this.root.querySelector(`#${id}`) as HTMLElement | null;
    if (!pane) {
      return;
    }
    pane.classList.toggle("active", active);
  }

  private getEditor(): GwEditor {
    return this.root.querySelector("gw-editor") as GwEditor;
  }

  private getNotes(): GwNotes {
    return this.root.querySelector("gw-notes") as GwNotes;
  }

  private getWiki(): GwWiki {
    return this.root.querySelector("gw-wiki") as GwWiki;
  }

  private getStatusBar(): GwStatusBar {
    return this.root.querySelector("gw-status-bar") as GwStatusBar;
  }

  private getCommandPalette(): GwCommandPalette {
    return this.root.querySelector("gw-command-palette") as GwCommandPalette;
  }

  private getModal(): GwModal {
    return this.root.querySelector("gw-modal") as GwModal;
  }

  private focusActivePane(): void {
    const view = this.store.getState().view;
    if (view === "notes") {
      this.getNotes().focusNotes();
      return;
    }

    if (view === "wiki") {
      this.getWiki().focusWiki();
      return;
    }

    this.getEditor().focusEditor();
  }
}

customElements.define("gw-app", GwApp);
