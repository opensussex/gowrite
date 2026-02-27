export interface HelpTopic {
  id: string;
  title: string;
  aliases: string[];
  type: "command" | "feature" | "shortcut";
  summary: string;
  usage?: string[];
  details: string[];
  examples?: string[];
}

const HELP_TOPICS: HelpTopic[] = [
  {
    id: "help",
    title: "help",
    aliases: ["?"],
    type: "command",
    summary: "Show this help system or details for one topic.",
    usage: ["help", "help <topic>"],
    details: [
      "Use 'help' with no args to see all commands/features.",
      "Use 'help <topic>' for detailed usage and examples.",
      "Topics include commands and features such as: save, open, chapter, chapters, notes, wiki, structure, wordcount, autosave, typewriter."
    ],
    examples: ["help chapter", "help open", "help autosave"]
  },
  {
    id: "save",
    title: "save",
    aliases: [],
    type: "command",
    summary: "Persist current project to local storage.",
    usage: ["save", "save <project name>"],
    details: [
      "Running 'save' writes the current project and settings.",
      "Running 'save <project name>' performs save-as behavior by creating a new project record with the same content.",
      "Saved projects are listed and searchable by the open/load command."
    ],
    examples: ["save", "save Draft One", "save \"My Novel v2\""]
  },
  {
    id: "open",
    title: "open / load",
    aliases: ["load"],
    type: "command",
    summary: "Open a project from local storage.",
    usage: ["open", "open <project id>", "open <project name>", "load <project name>"],
    details: [
      "'open' with no args shows currently available projects.",
      "'open <name>' matches exact name first, then fuzzy name contains.",
      "'open <id>' loads by project id directly."
    ],
    examples: ["open", "open Project Alpha", "load \"Story Draft\""]
  },
  {
    id: "chapter",
    title: "chapter",
    aliases: [],
    type: "command",
    summary: "Create, rename, or delete chapters.",
    usage: [
      "chapter new <title>",
      "chapter rename <title>",
      "chapter rename <index> <title>",
      "chapter delete",
      "chapter delete <index>"
    ],
    details: [
      "Chapter indexes are 1-based in command usage.",
      "Rename/delete default to the current chapter when index is omitted.",
      "Delete requires confirmation and cannot remove the only remaining chapter."
    ],
    examples: [
      "chapter new \"Act 1: Setup\"",
      "chapter rename 2 \"Inciting Incident\"",
      "chapter delete 3"
    ]
  },
  {
    id: "chapters",
    title: "chapters",
    aliases: ["list"],
    type: "command",
    summary: "Open chapter picker modal and jump to a chapter using keyboard.",
    usage: ["chapters"],
    details: [
      "Opens a modal list of chapters.",
      "Use Up/Down arrow keys to move selection.",
      "Press Enter to select chapter, Esc to cancel."
    ],
    examples: ["chapters"]
  },
  {
    id: "notes",
    title: "notes",
    aliases: [],
    type: "command",
    summary: "Toggle between main editor and chapter notes view.",
    usage: ["notes"],
    details: [
      "Notes are stored per chapter.",
      "Running 'notes' from main view opens notes.",
      "Running 'notes' again returns to the main editor."
    ],
    examples: ["notes"]
  },
  {
    id: "wiki",
    title: "wiki",
    aliases: [],
    type: "command",
    summary: "Open wiki view and manage wiki entries.",
    usage: [
      "wiki",
      "wiki new <title>",
      "wiki rename <title>",
      "wiki rename <index> <title>",
      "wiki delete",
      "wiki delete <index>"
    ],
    details: [
      "Wiki entries are project-level reference notes.",
      "Indexes are 1-based in commands.",
      "Delete requires confirmation and cannot remove the only wiki entry."
    ],
    examples: ["wiki new Villain", "wiki rename 2 Antagonist", "wiki delete 3"]
  },
  {
    id: "structure",
    title: "structure",
    aliases: ["template"],
    type: "command",
    summary: "Apply a chapter structure template.",
    usage: ["structure <name>"],
    details: [
      "Available templates: 3act, hero, cat, fichtean, horror.",
      "Applying a template replaces current chapters after confirmation.",
      "Wiki entries are preserved."
    ],
    examples: ["structure 3act", "structure horror"]
  },
  {
    id: "wordcount",
    title: "wordcount",
    aliases: [],
    type: "command",
    summary: "Show word, line, and character counts for the current chapter.",
    usage: ["wordcount"],
    details: ["Stats appear in a modal dialog so you can quickly inspect draft size."],
    examples: ["wordcount"]
  },
  {
    id: "autosave",
    title: "autosave",
    aliases: ["persistence", "storage"],
    type: "feature",
    summary: "Automatic local persistence for writing safety.",
    details: [
      "Edits are debounced and saved automatically.",
      "A periodic heartbeat save runs every 60 seconds when content is dirty.",
      "Refreshing the page restores the current project from local storage."
    ]
  },
  {
    id: "typewriter",
    title: "typewriter mode",
    aliases: ["focus", "center"],
    type: "feature",
    summary: "Keeps writing focus by centering the active line when possible.",
    details: [
      "Typewriter mode is enabled in settings by default.",
      "Current behavior uses caret-line scroll centering logic.",
      "It works best once content exceeds the visible editor height."
    ]
  },
  {
    id: "shortcuts",
    title: "keyboard shortcuts",
    aliases: ["keys", "hotkeys"],
    type: "shortcut",
    summary: "Fast keyboard controls.",
    usage: ["Ctrl/Cmd+K", "Ctrl/Cmd+S", "Ctrl/Cmd+N", "Ctrl/Cmd+W", "Esc"],
    details: [
      "Ctrl/Cmd+K: open/close command palette.",
      "Ctrl/Cmd+S: save current project immediately.",
      "Ctrl/Cmd+N: toggle notes view.",
      "Ctrl/Cmd+W: toggle wiki view.",
      "Esc: close command palette when open."
    ]
  }
];

function formatSection(label: string, lines: string[]): string {
  const body = lines.map((line) => `- ${line}`).join("\n");
  return `${label}:\n${body}`;
}

function formatTopic(topic: HelpTopic): { title: string; message: string } {
  const lines: string[] = [];
  lines.push(`Type: ${topic.type}`);
  lines.push(`Summary: ${topic.summary}`);

  if (topic.aliases.length > 0) {
    lines.push(`Aliases: ${topic.aliases.join(", ")}`);
  }

  if (topic.usage && topic.usage.length > 0) {
    lines.push("");
    lines.push(formatSection("Usage", topic.usage));
  }

  if (topic.details.length > 0) {
    lines.push("");
    lines.push(formatSection("Details", topic.details));
  }

  if (topic.examples && topic.examples.length > 0) {
    lines.push("");
    lines.push(formatSection("Examples", topic.examples));
  }

  return {
    title: `Help: ${topic.title}`,
    message: lines.join("\n")
  };
}

function formatOverview(): { title: string; message: string } {
  const commands = HELP_TOPICS.filter((topic) => topic.type === "command").map(
    (topic) => `${topic.id}: ${topic.summary}`
  );
  const features = HELP_TOPICS.filter((topic) => topic.type === "feature").map(
    (topic) => `${topic.id}: ${topic.summary}`
  );
  const shortcuts = HELP_TOPICS.filter((topic) => topic.type === "shortcut").map(
    (topic) => `${topic.id}: ${topic.summary}`
  );

  const message = [
    "gowrite Help",
    "",
    formatSection("Commands", commands),
    "",
    formatSection("Features", features),
    "",
    formatSection("Shortcuts", shortcuts),
    "",
    "Tip:",
    "- Run 'help <topic>' for detailed usage (example: help chapter)."
  ].join("\n");

  return {
    title: "Help",
    message
  };
}

export function resolveHelpTopic(query: string): HelpTopic | null {
  const normalized = query.trim().toLowerCase();
  if (!normalized) {
    return null;
  }

  const exact = HELP_TOPICS.find((topic) => topic.id === normalized || topic.title.toLowerCase() === normalized);
  if (exact) {
    return exact;
  }

  const alias = HELP_TOPICS.find((topic) => topic.aliases.map((alias) => alias.toLowerCase()).includes(normalized));
  if (alias) {
    return alias;
  }

  return HELP_TOPICS.find((topic) => topic.id.includes(normalized) || topic.title.toLowerCase().includes(normalized)) ?? null;
}

export function renderHelp(queryArgs: string[]): { title: string; message: string } {
  const query = queryArgs.join(" ").trim();
  if (!query) {
    return formatOverview();
  }

  const topic = resolveHelpTopic(query);
  if (!topic) {
    return {
      title: "Help",
      message: `No help topic found for '${query}'.\n\nRun 'help' to see available topics.`
    };
  }

  return formatTopic(topic);
}

export function getHelpTopics(): HelpTopic[] {
  return [...HELP_TOPICS];
}
