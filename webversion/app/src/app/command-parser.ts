export interface ParsedCommand {
  name: string;
  args: string[];
  raw: string;
}

const TOKEN_REGEX = /"([^"\\]|\\.)*"|'([^'\\]|\\.)*'|[^\s]+/g;

function unquote(token: string): string {
  const first = token[0];
  const last = token[token.length - 1];
  if ((first === '"' && last === '"') || (first === "'" && last === "'")) {
    return token.slice(1, -1).replace(/\\(["'\\])/g, "$1");
  }
  return token;
}

export function tokenizeCommand(input: string): string[] {
  const trimmed = input.trim();
  if (!trimmed) {
    return [];
  }

  const matches = trimmed.match(TOKEN_REGEX) ?? [];
  return matches.map(unquote);
}

export function parseCommandInput(input: string): ParsedCommand | null {
  const tokens = tokenizeCommand(input);
  if (tokens.length === 0) {
    return null;
  }

  const [name, ...args] = tokens;
  return {
    name: name.toLowerCase(),
    args,
    raw: input
  };
}
