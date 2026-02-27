export interface CommandSubmitDetail {
  command: string;
}

export class GwCommandPalette extends HTMLElement {
  private readonly root = this.attachShadow({ mode: "open" });
  private input!: HTMLInputElement;
  private panel!: HTMLElement;

  connectedCallback(): void {
    this.root.innerHTML = `
      <style>
        :host {
          display: block;
          position: relative;
          font-family: var(--tui-font);
        }

        .panel {
          border: 1px solid var(--tui-border-strong);
          background: linear-gradient(180deg, rgb(10 20 14 / 98%) 0%, rgb(7 14 10 / 98%) 100%);
          padding: 8px;
          display: none;
        }

        .panel.open {
          display: block;
        }

        .label {
          color: var(--tui-fg-soft);
          font-size: 12px;
          margin-bottom: 6px;
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }

        input {
          width: 100%;
          border: 1px solid var(--tui-border);
          background: var(--tui-bg-0);
          color: var(--tui-fg-main);
          padding: 8px;
          font: 14px/1.3 var(--tui-font);
          outline: none;
        }

        input:focus {
          border-color: var(--tui-border-strong);
          box-shadow: inset 0 0 0 1px rgb(125 255 168 / 20%);
        }
      </style>
      <section class="panel" aria-hidden="true">
        <div class="label">Command Palette</div>
        <input aria-label="Command input" placeholder="Type a command and press Enter" />
      </section>
    `;

    this.panel = this.root.querySelector(".panel") as HTMLElement;
    this.input = this.root.querySelector("input") as HTMLInputElement;

    this.input.addEventListener("keydown", (event) => {
      if (event.key === "Enter") {
        event.preventDefault();
        const command = this.input.value.trim();
        this.dispatchEvent(
          new CustomEvent<CommandSubmitDetail>("command-submit", {
            detail: { command },
            bubbles: true,
            composed: true
          })
        );
      }

      if (event.key === "Escape") {
        event.preventDefault();
        this.dispatchEvent(
          new CustomEvent("command-cancel", {
            bubbles: true,
            composed: true
          })
        );
      }
    });
  }

  setOpen(open: boolean): void {
    if (!this.panel || !this.input) {
      return;
    }

    this.panel.classList.toggle("open", open);
    this.panel.setAttribute("aria-hidden", String(!open));

    if (open) {
      this.input.value = "";
      this.input.focus();
    }
  }

  clear(): void {
    if (this.input) {
      this.input.value = "";
    }
  }
}

customElements.define("gw-command-palette", GwCommandPalette);
