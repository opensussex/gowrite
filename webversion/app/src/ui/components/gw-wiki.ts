import type { AppState } from "@/domain/models";
import { getCurrentWikiEntry } from "@/app/selectors";

export interface WikiEvents {
  onSelectEntry: (wikiId: string) => void;
  onChangeContent: (value: string) => void;
  onCursorChange: (row: number, column: number) => void;
}

export class GwWiki extends HTMLElement {
  private readonly root = this.attachShadow({ mode: "open" });
  private listEl!: HTMLElement;
  private textarea!: HTMLTextAreaElement;
  private titleEl!: HTMLElement;
  private events: WikiEvents = {
    onSelectEntry: () => {
      return;
    },
    onChangeContent: () => {
      return;
    },
    onCursorChange: () => {
      return;
    }
  };

  connectedCallback(): void {
    this.root.innerHTML = `
      <style>
        :host {
          display: block;
          font-family: var(--tui-font);
        }

        .layout {
          border: 1px solid var(--tui-border-strong);
          background: linear-gradient(180deg, rgb(8 17 11 / 96%) 0%, rgb(5 10 7 / 96%) 100%);
          display: grid;
          grid-template-columns: 260px minmax(0, 1fr);
          min-height: 66vh;
        }

        .list {
          border-right: 1px solid var(--tui-border);
          padding: 8px;
          overflow-y: auto;
        }

        .list-title {
          color: var(--tui-accent);
          text-transform: uppercase;
          font-size: 12px;
          margin-bottom: 8px;
          letter-spacing: 0.05em;
        }

        .entry {
          width: 100%;
          text-align: left;
          border: 1px solid var(--tui-border);
          background: var(--tui-bg-0);
          color: var(--tui-fg-soft);
          font: 13px/1.4 var(--tui-font);
          padding: 7px;
          margin-bottom: 6px;
          cursor: pointer;
        }

        .entry.active {
          border-color: var(--tui-border-strong);
          color: var(--tui-accent);
        }

        .editor {
          padding: 10px;
          display: grid;
          grid-template-rows: auto minmax(0, 1fr);
          gap: 8px;
        }

        .editor-title {
          color: var(--tui-accent);
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }

        textarea {
          width: 100%;
          min-height: 55vh;
          resize: vertical;
          border: 1px solid var(--tui-border);
          border-radius: 0;
          background: var(--tui-bg-0);
          color: var(--tui-fg-main);
          caret-color: var(--tui-accent);
          font: 15px/1.55 var(--tui-font);
          padding: 14px;
          box-sizing: border-box;
          outline: none;
        }

        textarea:focus {
          border-color: var(--tui-border-strong);
          box-shadow: inset 0 0 0 1px rgb(125 255 168 / 25%);
        }
      </style>
      <section class="layout">
        <aside class="list">
          <div class="list-title">Wiki Entries</div>
          <div id="entries"></div>
        </aside>
        <section class="editor">
          <div class="editor-title" id="entryTitle">WIKI : GENERAL NOTES</div>
          <textarea aria-label="Wiki entry editor" placeholder="Character notes, locations, lore..."></textarea>
        </section>
      </section>
    `;

    this.listEl = this.root.querySelector("#entries") as HTMLElement;
    this.titleEl = this.root.querySelector("#entryTitle") as HTMLElement;
    this.textarea = this.root.querySelector("textarea") as HTMLTextAreaElement;

    this.textarea.addEventListener("input", () => {
      this.events.onChangeContent(this.textarea.value);
    });

    const cursorHandler = () => {
      const { row, column } = this.computeCursorPosition(this.textarea.value, this.textarea.selectionStart);
      this.events.onCursorChange(row, column);
    };

    this.textarea.addEventListener("click", cursorHandler);
    this.textarea.addEventListener("keyup", cursorHandler);
  }

  setEvents(events: WikiEvents): void {
    this.events = events;
  }

  render(state: AppState): void {
    if (!this.listEl || !this.textarea) {
      return;
    }

    this.listEl.innerHTML = "";
    state.project.wiki.forEach((entry) => {
      const btn = document.createElement("button");
      btn.className = `entry${entry.id === state.currentWikiId ? " active" : ""}`;
      btn.type = "button";
      btn.textContent = entry.title;
      btn.addEventListener("click", () => {
        this.events.onSelectEntry(entry.id);
      });
      this.listEl.appendChild(btn);
    });

    const current = getCurrentWikiEntry(state);
    this.titleEl.textContent = `WIKI : ${(current?.title ?? "General Notes").toUpperCase()}`;

    const nextValue = current?.content ?? "";
    if (this.textarea.value !== nextValue) {
      this.textarea.value = nextValue;
    }
  }

  focusWiki(): void {
    this.textarea?.focus();
  }

  private computeCursorPosition(text: string, selectionStart: number): { row: number; column: number } {
    const textToCursor = text.slice(0, selectionStart);
    const lines = textToCursor.split("\n");
    return {
      row: lines.length,
      column: lines[lines.length - 1].length + 1
    };
  }
}

customElements.define("gw-wiki", GwWiki);
