import type { AppState } from "@/domain/models";
import { getCurrentChapter, getWordCount } from "@/app/selectors";

export class GwStatusBar extends HTMLElement {
  private readonly root = this.attachShadow({ mode: "open" });

  connectedCallback(): void {
    this.root.innerHTML = `
      <style>
        :host {
          display: block;
          font-family: var(--tui-font);
        }

        .bar {
          margin-top: 8px;
          border: 1px solid var(--tui-border-strong);
          background: linear-gradient(180deg, rgb(9 18 12 / 95%) 0%, rgb(7 13 10 / 95%) 100%);
          color: var(--tui-fg-main);
          padding: 8px 10px;
          display: flex;
          flex-wrap: wrap;
          gap: 10px;
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.04em;
        }

        .segment {
          color: var(--tui-fg-soft);
        }

        .segment strong {
          color: var(--tui-accent);
          font-weight: 600;
        }

        .state-dirty {
          color: var(--tui-warning);
        }
      </style>
      <div class="bar" role="status" aria-live="polite">Ready</div>
    `;
  }

  render(state: AppState): void {
    const bar = this.root.querySelector(".bar");
    if (!bar) {
      return;
    }

    const chapter = getCurrentChapter(state);
    const text = chapter?.content ?? "";
    const words = getWordCount(text);
    const saveState = state.dirty ? "DIRTY" : "CLEAN";
    const saveClass = state.dirty ? "segment state-dirty" : "segment";
    const typewriterState = state.settings.typewriterMode ? "ON" : "OFF";

    bar.innerHTML = `
      <span class="segment"><strong>WORDS</strong> ${words}</span>
      <span class="segment"><strong>ROW</strong> ${state.cursor.row}</span>
      <span class="segment"><strong>COL</strong> ${state.cursor.column}</span>
      <span class="segment"><strong>TYPEWRITER</strong> ${typewriterState}</span>
      <span class="${saveClass}"><strong>STATE</strong> ${saveState}</span>
      <span class="segment"><strong>MSG</strong> ${state.statusMessage.toUpperCase()}</span>
    `;
  }
}

customElements.define("gw-status-bar", GwStatusBar);
