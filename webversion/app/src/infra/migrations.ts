import { createDefaultWikiEntry, generateId } from "@/domain/defaults";
import { SCHEMA_VERSION, type AppSettings, type Chapter, type Project, type WikiEntry } from "@/domain/models";

function asRecord(input: unknown): Record<string, unknown> | null {
  if (!input || typeof input !== "object" || Array.isArray(input)) {
    return null;
  }
  return input as Record<string, unknown>;
}

function asString(input: unknown, fallback: string): string {
  if (typeof input === "string" && input.trim()) {
    return input;
  }
  return fallback;
}

function asNumber(input: unknown, fallback: number): number {
  if (typeof input === "number" && Number.isFinite(input)) {
    return input;
  }
  return fallback;
}

function migrateChapter(input: unknown, index: number): Chapter {
  const record = asRecord(input) ?? {};
  return {
    id: asString(record.id, generateId("chapter")),
    title: asString(record.title ?? record.Title, `Chapter ${index + 1}`),
    content: asString(record.content ?? record.Content, ""),
    notes: asString(record.notes ?? record.Notes, ""),
    target: asNumber(record.target ?? record.Target, 0)
  };
}

function migrateWikiEntry(input: unknown, index: number): WikiEntry {
  const record = asRecord(input) ?? {};
  return {
    id: asString(record.id, generateId("wiki")),
    title: asString(record.title ?? record.Title, `Wiki ${index + 1}`),
    content: asString(record.content ?? record.Content, "")
  };
}

function normalizeChapterArray(input: unknown): Chapter[] {
  if (!Array.isArray(input)) {
    return [];
  }
  return input.map((chapter, index) => migrateChapter(chapter, index));
}

function normalizeWikiArray(input: unknown): WikiEntry[] {
  if (!Array.isArray(input)) {
    return [];
  }
  return input.map((entry, index) => migrateWikiEntry(entry, index));
}

export function migrateProject(input: unknown): Project | null {
  if (Array.isArray(input)) {
    const chapters = normalizeChapterArray(input);
    if (chapters.length === 0) {
      return null;
    }
    const now = new Date().toISOString();
    return {
      id: generateId("project"),
      name: "Imported Project",
      chapters,
      wiki: [createDefaultWikiEntry()],
      createdAt: now,
      updatedAt: now,
      schemaVersion: SCHEMA_VERSION
    };
  }

  const record = asRecord(input);
  if (!record) {
    return null;
  }

  const chapters = normalizeChapterArray(record.chapters ?? record.Chapters);
  if (chapters.length === 0) {
    return null;
  }

  const wiki = normalizeWikiArray(record.wiki ?? record.Wiki);
  const now = new Date().toISOString();

  return {
    id: asString(record.id, generateId("project")),
    name: asString(record.name, "Untitled Project"),
    chapters,
    wiki: wiki.length > 0 ? wiki : [createDefaultWikiEntry()],
    createdAt: asString(record.createdAt, now),
    updatedAt: asString(record.updatedAt, now),
    schemaVersion: asNumber(record.schemaVersion, SCHEMA_VERSION)
  };
}

export function migrateSettings(input: unknown, fallback: AppSettings): AppSettings {
  const record = asRecord(input);
  if (!record) {
    return fallback;
  }

  const themeRaw = record.theme;
  const theme =
    themeRaw === "dark" || themeRaw === "light" || themeRaw === "retro" ? themeRaw : fallback.theme;

  return {
    ...fallback,
    ...record,
    theme,
    centered: typeof record.centered === "boolean" ? record.centered : fallback.centered,
    focusMode: typeof record.focusMode === "boolean" ? record.focusMode : fallback.focusMode,
    typewriterMode: typeof record.typewriterMode === "boolean" ? record.typewriterMode : fallback.typewriterMode,
    lastProjectId: asString(record.lastProjectId, fallback.lastProjectId)
  };
}
