import type { AppAction } from "./actions";
import type { AppState } from "@/domain/models";

function updateProjectTimestamp(state: AppState): AppState {
  return {
    ...state,
    project: {
      ...state.project,
      updatedAt: new Date().toISOString()
    }
  };
}

export function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case "INIT_APP":
      return {
        ...state,
        project: action.project,
        settings: action.settings,
        currentChapterId: action.project.chapters[0]?.id ?? state.currentChapterId,
        dirty: false,
        statusMessage: "Project loaded"
      };
    case "SET_CHAPTER_CONTENT": {
      const chapters = state.project.chapters.map((chapter) => {
        if (chapter.id !== action.chapterId) {
          return chapter;
        }
        return {
          ...chapter,
          content: action.content
        };
      });

      const next = {
        ...state,
        project: {
          ...state.project,
          chapters
        },
        dirty: true,
        statusMessage: "Editing..."
      };

      return updateProjectTimestamp(next);
    }
    case "SET_CURSOR":
      return {
        ...state,
        cursor: action.cursor
      };
    case "SAVE_SUCCEEDED":
      return {
        ...state,
        dirty: false,
        lastSavedAt: action.timestamp,
        statusMessage: `Saved at ${new Date(action.timestamp).toLocaleTimeString()}`
      };
    case "SAVE_FAILED":
      return {
        ...state,
        statusMessage: `Save failed: ${action.error}`
      };
    case "SET_STATUS":
      return {
        ...state,
        statusMessage: action.message
      };
    case "TOGGLE_COMMAND_PALETTE":
      return {
        ...state,
        commandPaletteOpen: !state.commandPaletteOpen,
        statusMessage: !state.commandPaletteOpen
          ? "Command palette placeholder (Sprint 2)"
          : state.statusMessage
      };
    default:
      return state;
  }
}
