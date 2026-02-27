type ModalValue = boolean | number | null;
type ModalMode = "info" | "confirm" | "list";

interface Resolver {
  (value: ModalValue): void;
}

export class GwModal extends HTMLElement {
  private readonly root = this.attachShadow({ mode: "open" });
  private overlay!: HTMLElement;
  private titleEl!: HTMLElement;
  private messageEl!: HTMLElement;
  private listEl!: HTMLElement;
  private okButton!: HTMLButtonElement;
  private cancelButton!: HTMLButtonElement;
  private resolver: Resolver | null = null;
  private mode: ModalMode = "info";
  private listItems: string[] = [];
  private selectedIndex = 0;

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
          width: min(560px, 94vw);
          max-height: 82vh;
          border: 1px solid var(--tui-border-strong);
          background: linear-gradient(180deg, rgb(10 20 14 / 98%) 0%, rgb(6 12 9 / 98%) 100%);
          padding: 12px;
          color: var(--tui-fg-main);
          box-shadow: 0 16px 28px rgb(0 0 0 / 35%);
          display: grid;
          grid-template-rows: auto auto minmax(0, 1fr) auto;
          gap: 8px;
        }

        h2 {
          margin: 0;
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

        .list {
          display: none;
          border: 1px solid var(--tui-border);
          padding: 6px;
          overflow-y: auto;
          max-height: 44vh;
          outline: none;
          background: rgb(4 7 5 / 80%);
        }

        .list.open {
          display: block;
        }

        .item {
          display: block;
          width: 100%;
          border: 1px solid var(--tui-border);
          background: var(--tui-bg-0);
          color: var(--tui-fg-soft);
          padding: 7px;
          margin-bottom: 6px;
          text-align: left;
          cursor: pointer;
          font: 13px/1.35 var(--tui-font);
        }

        .item.active {
          border-color: var(--tui-border-strong);
          color: var(--tui-accent);
          box-shadow: inset 0 0 0 1px rgb(125 255 168 / 20%);
        }

        .actions {
          margin-top: 4px;
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
          <div id="list" class="list" role="listbox" tabindex="0" aria-label="Selection list"></div>
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
    this.listEl = this.root.querySelector("#list") as HTMLElement;
    this.okButton = this.root.querySelector("#ok") as HTMLButtonElement;
    this.cancelButton = this.root.querySelector("#cancel") as HTMLButtonElement;

    this.okButton.addEventListener("click", () => {
      if (this.mode === "list") {
        this.close(this.selectedIndex);
        return;
      }
      this.close(true);
    });

    this.cancelButton.addEventListener("click", () => {
      if (this.mode === "list") {
        this.close(null);
        return;
      }
      this.close(false);
    });

    this.overlay.addEventListener("keydown", (event) => {
      if (event.key === "Escape") {
        event.preventDefault();
        if (this.mode === "list") {
          this.close(null);
        } else {
          this.close(false);
        }
      }
    });

    this.listEl.addEventListener("keydown", (event) => {
      if (this.mode !== "list") {
        return;
      }

      if (event.key === "ArrowDown") {
        event.preventDefault();
        this.selectedIndex = Math.min(this.selectedIndex + 1, this.listItems.length - 1);
        this.renderList();
        return;
      }

      if (event.key === "ArrowUp") {
        event.preventDefault();
        this.selectedIndex = Math.max(this.selectedIndex - 1, 0);
        this.renderList();
        return;
      }

      if (event.key === "Enter") {
        event.preventDefault();
        this.close(this.selectedIndex);
      }
    });
  }

  async showInfo(title: string, message: string): Promise<void> {
    this.mode = "info";
    this.open(title, message, {
      showCancel: false,
      okLabel: "OK",
      cancelLabel: "Cancel",
      showList: false
    });

    await new Promise<ModalValue>((resolve) => {
      this.resolver = resolve;
    });
  }

  async confirm(title: string, message: string): Promise<boolean> {
    this.mode = "confirm";
    this.open(title, message, {
      showCancel: true,
      okLabel: "Yes",
      cancelLabel: "No",
      showList: false
    });

    const result = await new Promise<ModalValue>((resolve) => {
      this.resolver = resolve;
    });

    return result === true;
  }

  async selectFromList(title: string, items: string[], initialIndex = 0): Promise<number | null> {
    this.mode = "list";
    this.listItems = [...items];
    this.selectedIndex = Math.max(0, Math.min(initialIndex, this.listItems.length - 1));

    this.open(title, "Use Up/Down to select. Enter to choose. Esc to cancel.", {
      showCancel: true,
      okLabel: "Select",
      cancelLabel: "Cancel",
      showList: true
    });

    const result = await new Promise<ModalValue>((resolve) => {
      this.resolver = resolve;
    });

    if (typeof result === "number") {
      return result;
    }
    return null;
  }

  private open(
    title: string,
    message: string,
    options: { showCancel: boolean; okLabel: string; cancelLabel: string; showList: boolean }
  ): void {
    this.titleEl.textContent = title;
    this.messageEl.textContent = message;
    this.cancelButton.style.display = options.showCancel ? "inline-block" : "none";
    this.okButton.textContent = options.okLabel;
    this.cancelButton.textContent = options.cancelLabel;

    this.listEl.classList.toggle("open", options.showList);
    if (options.showList) {
      this.renderList();
    }

    this.overlay.classList.add("open");
    this.overlay.setAttribute("aria-hidden", "false");

    if (options.showList) {
      this.listEl.focus();
    } else {
      this.okButton.focus();
    }
  }

  private renderList(): void {
    this.listEl.innerHTML = "";

    this.listItems.forEach((item, index) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = `item${index === this.selectedIndex ? " active" : ""}`;
      button.setAttribute("role", "option");
      button.setAttribute("aria-selected", String(index === this.selectedIndex));
      button.textContent = item;
      button.addEventListener("click", () => {
        this.selectedIndex = index;
        this.renderList();
        this.listEl.focus();
      });
      button.addEventListener("dblclick", () => {
        this.close(index);
      });
      this.listEl.appendChild(button);
    });

    const active = this.listEl.querySelector(".item.active") as HTMLElement | null;
    active?.scrollIntoView({ block: "nearest" });
  }

  private close(value: ModalValue): void {
    this.overlay.classList.remove("open");
    this.overlay.setAttribute("aria-hidden", "true");
    this.listEl.classList.remove("open");

    if (this.resolver) {
      this.resolver(value);
      this.resolver = null;
    }
  }
}

customElements.define("gw-modal", GwModal);
