import { generateId } from "@/domain/defaults";
import type { Project } from "@/domain/models";
import { getChapterByIndex, getCurrentChapter, getCurrentWikiEntry, getWikiByIndex, getWordCount } from "@/app/selectors";
import { CommandBus, type CommandContext, type CommandResult } from "@/app/command-bus";
import { renderHelp } from "@/app/help-registry";
import { getStructureNames, STRUCTURE_TEMPLATES } from "@/domain/structures";

function parseOneBasedIndex(raw: string | undefined): number | null {
  if (!raw) {
    return null;
  }
  const parsed = Number.parseInt(raw, 10);
  if (!Number.isFinite(parsed) || parsed < 1) {
    return null;
  }
  return parsed - 1;
}

function cloneAsNewProject(project: Project, name: string): Project {
  const now = new Date().toISOString();
  return {
    ...project,
    id: generateId("project"),
    name,
    createdAt: now,
    updatedAt: now,
    chapters: project.chapters.map((chapter) => ({ ...chapter })),
    wiki: project.wiki.map((entry) => ({ ...entry }))
  };
}

async function handleSave(context: CommandContext, args: string[]): Promise<CommandResult> {
  const name = args.join(" ").trim();
  const state = context.getState();

  if (name && name !== state.project.name) {
    const newProject = cloneAsNewProject(state.project, name);
    context.dispatch({ type: "SET_PROJECT", project: newProject });
    context.dispatch({ type: "UPDATE_SETTINGS", settings: { lastProjectId: newProject.id } });
  }

  context.persistCurrentProject(true);

  const latest = context.getState();
  return {
    ok: true,
    message: `Saved project '${latest.project.name}'`
  };
}

async function handleOpen(context: CommandContext, args: string[]): Promise<CommandResult> {
  const query = args.join(" ").trim();
  if (!query) {
    const projects = context.repo.listProjects();
    if (projects.length === 0) {
      await context.showModal("Open Project", "No projects found in local storage.");
      return { ok: true, message: "No projects found" };
    }

    const list = projects.map((project) => `- ${project.name} (${project.id})`).join("\n");
    await context.showModal("Open Project", `Available projects:\n\n${list}`);
    return { ok: true, message: "Listed available projects" };
  }

  const project = context.repo.findProject(query);
  if (!project) {
    return { ok: false, error: `Project '${query}' not found` };
  }

  context.dispatch({ type: "SET_PROJECT", project });
  context.dispatch({ type: "UPDATE_SETTINGS", settings: { lastProjectId: project.id } });

  const settings = context.getState().settings;
  context.repo.setCurrentProjectId(project.id);
  context.repo.saveSettings(settings);

  return {
    ok: true,
    message: `Opened project '${project.name}'`
  };
}

function resolveChapterFromArgs(context: CommandContext, args: string[]): { chapterId: string; titleArgsOffset: number } {
  const state = context.getState();
  const maybeIndex = parseOneBasedIndex(args[0]);
  if (maybeIndex !== null) {
    const chapter = getChapterByIndex(state, maybeIndex);
    if (!chapter) {
      throw new Error(`Chapter ${maybeIndex + 1} not found`);
    }
    return {
      chapterId: chapter.id,
      titleArgsOffset: 1
    };
  }

  const current = getCurrentChapter(state);
  if (!current) {
    throw new Error("No current chapter selected");
  }

  return {
    chapterId: current.id,
    titleArgsOffset: 0
  };
}

async function handleChapter(context: CommandContext, args: string[]): Promise<CommandResult> {
  const sub = args[0]?.toLowerCase();
  if (!sub) {
    return { ok: false, error: "Usage: chapter <new|rename|delete> ..." };
  }

  if (sub === "new") {
    const title = args.slice(1).join(" ").trim() || "New Chapter";
    context.dispatch({ type: "CREATE_CHAPTER", title });
    return { ok: true, message: `Created chapter '${title}'` };
  }

  if (sub === "rename") {
    const chapterRef = resolveChapterFromArgs(context, args.slice(1));
    const newTitle = args.slice(1 + chapterRef.titleArgsOffset).join(" ").trim();
    if (!newTitle) {
      return { ok: false, error: "Usage: chapter rename [index] <title>" };
    }

    context.dispatch({
      type: "RENAME_CHAPTER",
      chapterId: chapterRef.chapterId,
      title: newTitle
    });
    return { ok: true, message: `Renamed chapter to '${newTitle}'` };
  }

  if (sub === "delete") {
    const state = context.getState();
    if (state.project.chapters.length <= 1) {
      return { ok: false, error: "Cannot delete the only chapter" };
    }

    const maybeIndex = parseOneBasedIndex(args[1]);
    const chapter = maybeIndex !== null ? getChapterByIndex(state, maybeIndex) : getCurrentChapter(state);

    if (!chapter) {
      return { ok: false, error: "Chapter not found" };
    }

    const confirmed = await context.confirm("Delete Chapter", `Delete '${chapter.title}'?`);
    if (!confirmed) {
      return { ok: true, message: "Delete canceled" };
    }

    context.dispatch({ type: "DELETE_CHAPTER", chapterId: chapter.id });
    return { ok: true, message: `Deleted chapter '${chapter.title}'` };
  }

  return { ok: false, error: `Unknown chapter subcommand '${sub}'` };
}

async function handleWordCount(context: CommandContext): Promise<CommandResult> {
  const state = context.getState();
  if (state.view === "wiki") {
    const currentWiki = getCurrentWikiEntry(state);
    if (!currentWiki) {
      return { ok: false, error: "No current wiki entry" };
    }
    const words = getWordCount(currentWiki.content);
    const lines = currentWiki.content.length === 0 ? 0 : currentWiki.content.split("\n").length;
    const chars = currentWiki.content.length;
    await context.showModal("Word Count (Wiki)", `Words: ${words}\nLines: ${lines}\nChars: ${chars}`);
    return {
      ok: true,
      message: `Word count: ${words}`
    };
  }

  const chapter = getCurrentChapter(state);
  if (!chapter) {
    return { ok: false, error: "No current chapter" };
  }

  const text = state.view === "notes" ? chapter.notes : chapter.content;
  const words = getWordCount(text);
  const lines = text.length === 0 ? 0 : text.split("\n").length;
  const chars = text.length;

  await context.showModal("Word Count", `Words: ${words}\nLines: ${lines}\nChars: ${chars}`);

  return {
    ok: true,
    message: `Word count: ${words}`
  };
}

async function handleHelp(context: CommandContext, args: string[]): Promise<CommandResult> {
  const details = renderHelp(args);
  await context.showModal(details.title, details.message);
  return {
    ok: true,
    message: args.length > 0 ? `Help opened for '${args.join(" ")}'` : "Help opened"
  };
}

function parseWikiArgs(context: CommandContext, args: string[]): { wikiId: string; titleArgsOffset: number } {
  const state = context.getState();
  const maybeIndex = parseOneBasedIndex(args[0]);
  if (maybeIndex !== null) {
    const entry = getWikiByIndex(state, maybeIndex);
    if (!entry) {
      throw new Error(`Wiki entry ${maybeIndex + 1} not found`);
    }
    return {
      wikiId: entry.id,
      titleArgsOffset: 1
    };
  }

  const current = getCurrentWikiEntry(state);
  if (!current) {
    throw new Error("No current wiki entry selected");
  }
  return {
    wikiId: current.id,
    titleArgsOffset: 0
  };
}

async function handleWiki(context: CommandContext, args: string[]): Promise<CommandResult> {
  const sub = args[0]?.toLowerCase();
  if (!sub) {
    context.dispatch({ type: "SET_VIEW", view: "wiki" });
    return { ok: true, message: "Wiki view opened" };
  }

  if (sub === "new") {
    const title = args.slice(1).join(" ").trim() || "New Entry";
    context.dispatch({ type: "CREATE_WIKI_ENTRY", title });
    return { ok: true, message: `Created wiki entry '${title}'` };
  }

  if (sub === "rename") {
    const target = parseWikiArgs(context, args.slice(1));
    const title = args.slice(1 + target.titleArgsOffset).join(" ").trim();
    if (!title) {
      return { ok: false, error: "Usage: wiki rename [index] <title>" };
    }
    context.dispatch({ type: "RENAME_WIKI_ENTRY", wikiId: target.wikiId, title });
    return { ok: true, message: `Renamed wiki entry to '${title}'` };
  }

  if (sub === "delete") {
    const state = context.getState();
    if (state.project.wiki.length <= 1) {
      return { ok: false, error: "Cannot delete the only wiki entry" };
    }
    const target = parseWikiArgs(context, args.slice(1));
    const entry = state.project.wiki.find((item) => item.id === target.wikiId);
    if (!entry) {
      return { ok: false, error: "Wiki entry not found" };
    }
    const confirmed = await context.confirm("Delete Wiki Entry", `Delete '${entry.title}'?`);
    if (!confirmed) {
      return { ok: true, message: "Delete canceled" };
    }
    context.dispatch({ type: "DELETE_WIKI_ENTRY", wikiId: target.wikiId });
    return { ok: true, message: `Deleted wiki entry '${entry.title}'` };
  }

  return { ok: false, error: `Unknown wiki subcommand '${sub}'` };
}

async function handleNotes(context: CommandContext): Promise<CommandResult> {
  const state = context.getState();
  const nextView = state.view === "notes" ? "main" : "notes";
  context.dispatch({ type: "SET_VIEW", view: nextView });
  return {
    ok: true,
    message: nextView === "notes" ? "Notes view opened" : "Returned to main editor"
  };
}

async function handleStructure(context: CommandContext, args: string[]): Promise<CommandResult> {
  const name = args[0]?.toLowerCase();
  if (!name) {
    return {
      ok: false,
      error: `Usage: structure <name>. Available: ${getStructureNames().join(", ")}`
    };
  }

  if (!STRUCTURE_TEMPLATES[name]) {
    return {
      ok: false,
      error: `Unknown structure '${name}'. Available: ${getStructureNames().join(", ")}`
    };
  }

  const confirmed = await context.confirm(
    "Apply Structure",
    `Apply '${name}' template? This replaces current chapters.`
  );
  if (!confirmed) {
    return { ok: true, message: "Structure apply canceled" };
  }

  context.dispatch({ type: "APPLY_STRUCTURE", name });
  return {
    ok: true,
    message: `Applied structure '${name}'`
  };
}

export function registerCoreCommands(bus: CommandBus): void {
  bus.register("save", handleSave);
  bus.register("open", handleOpen, ["load"]);
  bus.register("chapter", handleChapter);
  bus.register("notes", handleNotes);
  bus.register("wiki", handleWiki);
  bus.register("structure", handleStructure);
  bus.register("wordcount", handleWordCount);
  bus.register("help", handleHelp, ["h", "?"]);
}
