import type { AppAction } from "./actions";
import type { AppState } from "@/domain/models";
import type { LocalStorageRepository } from "@/infra/local-storage-repo";
import { parseCommandInput, type ParsedCommand } from "./command-parser";

export interface CommandResult {
  ok: boolean;
  message?: string;
  error?: string;
}

export interface CommandContext {
  getState(): AppState;
  dispatch(action: AppAction): void;
  repo: LocalStorageRepository;
  persistCurrentProject(force?: boolean): void;
  showModal(title: string, message: string): Promise<void>;
  confirm(title: string, message: string): Promise<boolean>;
  selectFromList(title: string, items: string[], initialIndex?: number): Promise<number | null>;
}

export type CommandHandler = (
  context: CommandContext,
  args: string[],
  parsed: ParsedCommand
) => Promise<CommandResult> | CommandResult;

export class CommandBus {
  private readonly handlers = new Map<string, CommandHandler>();

  register(name: string, handler: CommandHandler, aliases: string[] = []): void {
    this.handlers.set(name.toLowerCase(), handler);
    aliases.forEach((alias) => {
      this.handlers.set(alias.toLowerCase(), handler);
    });
  }

  async execute(rawInput: string, context: CommandContext): Promise<CommandResult> {
    const parsed = parseCommandInput(rawInput);
    if (!parsed) {
      return { ok: false, error: "Command is empty" };
    }

    const handler = this.handlers.get(parsed.name);
    if (!handler) {
      return { ok: false, error: `Unknown command: ${parsed.name}` };
    }

    try {
      return await handler(context, parsed.args, parsed);
    } catch (error) {
      return {
        ok: false,
        error: error instanceof Error ? error.message : "Unhandled command error"
      };
    }
  }
}
