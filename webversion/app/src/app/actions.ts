import type { AppSettings, CursorPosition, Project } from "@/domain/models";

export type AppAction =
  | { type: "INIT_APP"; project: Project; settings: AppSettings }
  | { type: "SET_CHAPTER_CONTENT"; chapterId: string; content: string }
  | { type: "SET_CURSOR"; cursor: CursorPosition }
  | { type: "SAVE_SUCCEEDED"; timestamp: number }
  | { type: "SAVE_FAILED"; error: string }
  | { type: "SET_STATUS"; message: string }
  | { type: "TOGGLE_COMMAND_PALETTE" };
