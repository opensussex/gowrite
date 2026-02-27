import type { AppState } from "@/domain/models";
import { getCurrentChapter } from "@/app/selectors";

export interface EditorEvents {
  onChange: (value: string) => void;
  onCursorChange: (row: number, column: number) => void;
}

export class GwEditor extends HTMLElement {
  private readonly root = this.attachShadow({ mode: "open" });
  private textarea!: HTMLTextAreaElement;
  private chapterTitle!: HTMLElement;
  private typewriterMode = true;
  private frameRequested = false;
  private readonly resizeListener = () => {
    this.scheduleTypewriterCenter();
  };
  private events: EditorEvents = {
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
          <span class="label" id="chapterTitle">CHAPTER 1 : THE BEGINNING</span>
          <span>MODE: INSERT</span>
        </header>
        <div class="panel-body">
          <textarea aria-label="Chapter editor" placeholder="Start writing your masterpiece..."></textarea>
        </div>
      </section>
    `;

    this.textarea = this.root.querySelector("textarea") as HTMLTextAreaElement;
    this.chapterTitle = this.root.querySelector("#chapterTitle") as HTMLElement;

    this.textarea.addEventListener("input", () => {
      this.events.onChange(this.textarea.value);
      this.scheduleTypewriterCenter();
    });

    const cursorHandler = () => {
      const { row, column } = this.computeCursorPosition(this.textarea.value, this.textarea.selectionStart);
      this.events.onCursorChange(row, column);
      this.scheduleTypewriterCenter();
    };

    this.textarea.addEventListener("click", cursorHandler);
    this.textarea.addEventListener("keyup", cursorHandler);
    window.addEventListener("resize", this.resizeListener);
    this.scheduleTypewriterCenter();
  }

  disconnectedCallback(): void {
    window.removeEventListener("resize", this.resizeListener);
  }

  setEvents(events: EditorEvents): void {
    this.events = events;
  }

  setTypewriterMode(enabled: boolean): void {
    this.typewriterMode = enabled;
    this.scheduleTypewriterCenter();
  }

  render(state: AppState): void {
    if (!this.textarea) {
      return;
    }

    const chapter = getCurrentChapter(state);
    const nextValue = chapter?.content ?? "";
    if (this.textarea.value !== nextValue) {
      this.textarea.value = nextValue;
    }

    if (chapter) {
      this.chapterTitle.textContent = `CHAPTER : ${chapter.title.toUpperCase()}`;
    }

    this.setTypewriterMode(state.settings.typewriterMode);
  }

  focusEditor(): void {
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

  private scheduleTypewriterCenter(): void {
    if (!this.typewriterMode || this.frameRequested) {
      return;
    }
    this.frameRequested = true;
    window.requestAnimationFrame(() => {
      this.frameRequested = false;
      this.centerCaretLine();
    });
  }

  private centerCaretLine(): void {
    if (!this.typewriterMode || !this.textarea) {
      return;
    }

    if (this.textarea.clientHeight === 0 || this.textarea.clientWidth === 0) {
      return;
    }

    const styles = window.getComputedStyle(this.textarea);
    const mirror = document.createElement("div");
    const textBeforeCaret = this.textarea.value.slice(0, this.textarea.selectionStart);
    const marker = document.createElement("span");
    marker.textContent = this.textarea.value.slice(this.textarea.selectionStart, this.textarea.selectionStart + 1) || "\u200b";

    const mirroredProperties = [
      "boxSizing",
      "width",
      "height",
      "overflowX",
      "overflowY",
      "borderTopWidth",
      "borderRightWidth",
      "borderBottomWidth",
      "borderLeftWidth",
      "paddingTop",
      "paddingRight",
      "paddingBottom",
      "paddingLeft",
      "fontStyle",
      "fontVariant",
      "fontWeight",
      "fontStretch",
      "fontSize",
      "lineHeight",
      "fontFamily",
      "textAlign",
      "textTransform",
      "textIndent",
      "letterSpacing"
    ] as const;

    mirror.style.position = "absolute";
    mirror.style.visibility = "hidden";
    mirror.style.whiteSpace = "pre-wrap";
    mirror.style.wordBreak = "break-word";
    mirror.style.overflowWrap = "break-word";
    mirror.style.top = "-9999px";
    mirror.style.left = "-9999px";
    mirror.style.width = `${this.textarea.clientWidth}px`;

    mirroredProperties.forEach((property) => {
      mirror.style[property] = styles[property];
    });

    mirror.textContent = textBeforeCaret;
    mirror.appendChild(marker);
    document.body.appendChild(mirror);

    const borderTop = parseFloat(styles.borderTopWidth) || 0;
    const paddingTop = parseFloat(styles.paddingTop) || 0;
    const lineHeight = parseFloat(styles.lineHeight) || (parseFloat(styles.fontSize) || 15) * 1.55;
    const caretTop = Math.max(0, marker.offsetTop - borderTop - paddingTop);

    document.body.removeChild(mirror);

    const desired = caretTop - this.textarea.clientHeight / 2 + lineHeight / 2;
    const maxScrollTop = Math.max(0, this.textarea.scrollHeight - this.textarea.clientHeight);
    const nextScrollTop = Math.min(Math.max(desired, 0), maxScrollTop);
    this.textarea.scrollTop = nextScrollTop;
  }
}

customElements.define("gw-editor", GwEditor);
