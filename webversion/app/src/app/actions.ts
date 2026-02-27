import type { AppSettings, CursorPosition, Project, ViewMode } from "@/domain/models";

export type AppAction =
  | { type: "INIT_APP"; project: Project; settings: AppSettings }
  | { type: "SET_PROJECT"; project: Project }
  | { type: "SET_VIEW"; view: ViewMode }
  | { type: "SET_CHAPTER_CONTENT"; chapterId: string; content: string }
  | { type: "SET_CHAPTER_NOTES"; chapterId: string; notes: string }
  | { type: "CREATE_CHAPTER"; title: string }
  | { type: "RENAME_CHAPTER"; chapterId: string; title: string }
  | { type: "DELETE_CHAPTER"; chapterId: string }
  | { type: "SET_CURRENT_CHAPTER"; chapterId: string }
  | { type: "APPLY_STRUCTURE"; name: string }
  | { type: "CREATE_WIKI_ENTRY"; title: string }
  | { type: "RENAME_WIKI_ENTRY"; wikiId: string; title: string }
  | { type: "DELETE_WIKI_ENTRY"; wikiId: string }
  | { type: "SET_CURRENT_WIKI"; wikiId: string }
  | { type: "SET_WIKI_CONTENT"; wikiId: string; content: string }
  | { type: "UPDATE_SETTINGS"; settings: Partial<AppSettings> }
  | { type: "SET_CURSOR"; cursor: CursorPosition }
  | { type: "SAVE_SUCCEEDED"; timestamp: number }
  | { type: "SAVE_FAILED"; error: string }
  | { type: "SET_STATUS"; message: string }
  | { type: "SET_COMMAND_PALETTE_OPEN"; open: boolean };
