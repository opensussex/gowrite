import type { AppSettings, Chapter, Project, WikiEntry } from "./models";
import { SCHEMA_VERSION } from "./models";

export function generateId(prefix: string): string {
  const random = Math.random().toString(36).slice(2, 10);
  return `${prefix}_${random}`;
}

export function createDefaultChapter(): Chapter {
  return {
    id: generateId("chapter"),
    title: "The Beginning",
    content: "",
    notes: "",
    target: 0
  };
}

export function createDefaultWikiEntry(): WikiEntry {
  return {
    id: generateId("wiki"),
    title: "General Notes",
    content: ""
  };
}

export function createDefaultProject(name = "Untitled Project"): Project {
  const now = new Date().toISOString();
  return {
    id: generateId("project"),
    name,
    chapters: [createDefaultChapter()],
    wiki: [createDefaultWikiEntry()],
    createdAt: now,
    updatedAt: now,
    schemaVersion: SCHEMA_VERSION
  };
}

export function createDefaultSettings(projectId: string): AppSettings {
  return {
    theme: "dark",
    centered: false,
    focusMode: false,
    typewriterMode: true,
    lastProjectId: projectId
  };
}
