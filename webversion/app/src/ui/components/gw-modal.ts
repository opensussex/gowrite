interface Resolver {
  (value: boolean): void;
}

export class GwModal extends HTMLElement {
  private readonly root = this.attachShadow({ mode: "open" });
  private overlay!: HTMLElement;
  private titleEl!: HTMLElement;
  private messageEl!: HTMLElement;
  private okButton!: HTMLButtonElement;
  private cancelButton!: HTMLButtonElement;
  private resolver: Resolver | null = null;

  connectedCallback(): void {
    this.root.innerHTML = `
      <style>
        :host {
          display: block;
          position: fixed;
          inset: 0;
          pointer-events: none;
          font-family: var(--tui-font);
        }

        .overlay {
          position: absolute;
          inset: 0;
          background: rgb(0 0 0 / 55%);
          display: none;
          align-items: center;
          justify-content: center;
          pointer-events: none;
        }

        .overlay.open {
          display: flex;
          pointer-events: auto;
        }

        .dialog {
          width: min(520px, 92vw);
          max-height: 82vh;
          border: 1px solid var(--tui-border-strong);
          background: linear-gradient(180deg, rgb(10 20 14 / 98%) 0%, rgb(6 12 9 / 98%) 100%);
          padding: 12px;
          color: var(--tui-fg-main);
          box-shadow: 0 16px 28px rgb(0 0 0 / 35%);
          display: grid;
          grid-template-rows: auto minmax(0, 1fr) auto;
          gap: 8px;
        }

        h2 {
          margin: 0 0 8px;
          font-size: 13px;
          color: var(--tui-accent);
          text-transform: uppercase;
          letter-spacing: 0.06em;
        }

        p {
          margin: 0;
          white-space: pre-wrap;
          color: var(--tui-fg-soft);
          font-size: 13px;
          line-height: 1.5;
          overflow-y: auto;
          padding-right: 2px;
        }

        .actions {
          margin-top: 12px;
          display: flex;
          justify-content: flex-end;
          gap: 8px;
        }

        button {
          border: 1px solid var(--tui-border);
          background: var(--tui-bg-0);
          color: var(--tui-fg-main);
          padding: 6px 10px;
          cursor: pointer;
          font: 13px/1.2 var(--tui-font);
          text-transform: uppercase;
        }

        button:hover {
          border-color: var(--tui-border-strong);
          color: var(--tui-accent);
        }
      </style>
      <div class="overlay" role="dialog" aria-modal="true" aria-hidden="true">
        <div class="dialog">
          <h2 id="title">Notice</h2>
          <p id="message"></p>
          <div class="actions">
            <button id="cancel">Cancel</button>
            <button id="ok">OK</button>
          </div>
        </div>
      </div>
    `;

    this.overlay = this.root.querySelector(".overlay") as HTMLElement;
    this.titleEl = this.root.querySelector("#title") as HTMLElement;
    this.messageEl = this.root.querySelector("#message") as HTMLElement;
    this.okButton = this.root.querySelector("#ok") as HTMLButtonElement;
    this.cancelButton = this.root.querySelector("#cancel") as HTMLButtonElement;

    this.okButton.addEventListener("click", () => this.close(true));
    this.cancelButton.addEventListener("click", () => this.close(false));
  }

  async showInfo(title: string, message: string): Promise<void> {
    this.open(title, message, false);
    await new Promise<boolean>((resolve) => {
      this.resolver = resolve;
    });
  }

  async confirm(title: string, message: string): Promise<boolean> {
    this.open(title, message, true);
    return await new Promise<boolean>((resolve) => {
      this.resolver = resolve;
    });
  }

  private open(title: string, message: string, showCancel: boolean): void {
    this.titleEl.textContent = title;
    this.messageEl.textContent = message;
    this.cancelButton.style.display = showCancel ? "inline-block" : "none";
    this.overlay.classList.add("open");
    this.overlay.setAttribute("aria-hidden", "false");
    this.okButton.focus();
  }

  private close(value: boolean): void {
    this.overlay.classList.remove("open");
    this.overlay.setAttribute("aria-hidden", "true");
    if (this.resolver) {
      this.resolver(value);
      this.resolver = null;
    }
  }
}

customElements.define("gw-modal", GwModal);
