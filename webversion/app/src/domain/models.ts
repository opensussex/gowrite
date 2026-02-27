export type ViewMode = "main" | "notes" | "analyze" | "wiki";

export interface Chapter {
  id: string;
  title: string;
  content: string;
  notes: string;
  target: number;
}

export interface WikiEntry {
  id: string;
  title: string;
  content: string;
}

export interface Project {
  id: string;
  name: string;
  chapters: Chapter[];
  wiki: WikiEntry[];
  createdAt: string;
  updatedAt: string;
  schemaVersion: number;
}

export interface AppSettings {
  theme: "dark" | "light" | "retro";
  centered: boolean;
  focusMode: boolean;
  typewriterMode: boolean;
  lastProjectId: string;
}

export interface CursorPosition {
  row: number;
  column: number;
}

export interface AppState {
  project: Project;
  settings: AppSettings;
  view: ViewMode;
  currentChapterId: string;
  currentWikiId: string;
  dirty: boolean;
  lastSavedAt: number | null;
  cursor: CursorPosition;
  commandPaletteOpen: boolean;
  statusMessage: string;
}

export const SCHEMA_VERSION = 1;
