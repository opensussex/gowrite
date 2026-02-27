import type { AppState } from "@/domain/models";
import { getCurrentChapter } from "@/app/selectors";

export interface NotesEvents {
  onChange: (value: string) => void;
  onCursorChange: (row: number, column: number) => void;
}

export class GwNotes extends HTMLElement {
  private readonly root = this.attachShadow({ mode: "open" });
  private textarea!: HTMLTextAreaElement;
  private notesTitle!: HTMLElement;
  private events: NotesEvents = {
    onChange: () => {
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
          height: 100%;
          font-family: var(--tui-font);
        }

        .panel {
          border: 1px solid var(--tui-border-strong);
          background: linear-gradient(180deg, rgb(8 17 11 / 96%) 0%, rgb(5 10 7 / 96%) 100%);
        }

        .panel-head {
          display: flex;
          justify-content: space-between;
          gap: 8px;
          padding: 8px 10px;
          border-bottom: 1px solid var(--tui-border);
          color: var(--tui-fg-soft);
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }

        .panel-head .label {
          color: var(--tui-accent);
        }

        .panel-body {
          padding: 10px;
        }

        textarea {
          width: 100%;
          min-height: 66vh;
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

        textarea::placeholder {
          color: var(--tui-fg-dim);
        }

        textarea:focus {
          border-color: var(--tui-border-strong);
          box-shadow: inset 0 0 0 1px rgb(125 255 168 / 25%);
        }
      </style>
      <section class="panel">
        <header class="panel-head">
          <span class="label" id="notesTitle">NOTES : THE BEGINNING</span>
          <span>MODE: NOTES</span>
        </header>
        <div class="panel-body">
          <textarea aria-label="Chapter notes editor" placeholder="Scene notes, reminders, and structure ideas..."></textarea>
        </div>
      </section>
    `;

    this.textarea = this.root.querySelector("textarea") as HTMLTextAreaElement;
    this.notesTitle = this.root.querySelector("#notesTitle") as HTMLElement;

    this.textarea.addEventListener("input", () => {
      this.events.onChange(this.textarea.value);
    });

    const cursorHandler = () => {
      const { row, column } = this.computeCursorPosition(this.textarea.value, this.textarea.selectionStart);
      this.events.onCursorChange(row, column);
    };

    this.textarea.addEventListener("click", cursorHandler);
    this.textarea.addEventListener("keyup", cursorHandler);
  }

  setEvents(events: NotesEvents): void {
    this.events = events;
  }

  render(state: AppState): void {
    if (!this.textarea) {
      return;
    }

    const chapter = getCurrentChapter(state);
    const nextValue = chapter?.notes ?? "";
    if (this.textarea.value !== nextValue) {
      this.textarea.value = nextValue;
    }

    if (chapter) {
      this.notesTitle.textContent = `NOTES : ${chapter.title.toUpperCase()}`;
    }
  }

  focusNotes(): void {
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

customElements.define("gw-notes", GwNotes);
