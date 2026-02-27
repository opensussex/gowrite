import type { AppAction } from "./actions";
import type { AppState } from "@/domain/models";
import { generateId } from "@/domain/defaults";
import { STRUCTURE_TEMPLATES } from "@/domain/structures";

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
        view: "main",
        currentChapterId: action.project.chapters[0]?.id ?? state.currentChapterId,
        currentWikiId: action.project.wiki[0]?.id ?? state.currentWikiId,
        dirty: false,
        statusMessage: "Project loaded"
      };
    case "SET_PROJECT":
      return {
        ...state,
        project: action.project,
        view: "main",
        currentChapterId: action.project.chapters[0]?.id ?? state.currentChapterId,
        currentWikiId: action.project.wiki[0]?.id ?? state.currentWikiId,
        dirty: false,
        statusMessage: `Opened project '${action.project.name}'`
      };
    case "SET_VIEW":
      return {
        ...state,
        view: action.view
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
    case "SET_CHAPTER_NOTES": {
      const chapters = state.project.chapters.map((chapter) => {
        if (chapter.id !== action.chapterId) {
          return chapter;
        }
        return {
          ...chapter,
          notes: action.notes
        };
      });

      const next = {
        ...state,
        project: {
          ...state.project,
          chapters
        },
        dirty: true,
        statusMessage: "Notes updated"
      };

      return updateProjectTimestamp(next);
    }
    case "CREATE_CHAPTER": {
      const newChapter = {
        id: generateId("chapter"),
        title: action.title.trim() || "New Chapter",
        content: "",
        notes: "",
        target: 0
      };

      const next = {
        ...state,
        project: {
          ...state.project,
          chapters: [...state.project.chapters, newChapter]
        },
        currentChapterId: newChapter.id,
        dirty: true,
        statusMessage: `Created chapter '${newChapter.title}'`
      };

      return updateProjectTimestamp(next);
    }
    case "RENAME_CHAPTER": {
      const chapters = state.project.chapters.map((chapter) =>
        chapter.id === action.chapterId
          ? {
              ...chapter,
              title: action.title.trim() || chapter.title
            }
          : chapter
      );

      const next = {
        ...state,
        project: {
          ...state.project,
          chapters
        },
        dirty: true,
        statusMessage: "Chapter renamed"
      };

      return updateProjectTimestamp(next);
    }
    case "DELETE_CHAPTER": {
      if (state.project.chapters.length <= 1) {
        return {
          ...state,
          statusMessage: "Cannot delete the only chapter"
        };
      }

      const targetIndex = state.project.chapters.findIndex((chapter) => chapter.id === action.chapterId);
      if (targetIndex < 0) {
        return state;
      }

      const chapters = state.project.chapters.filter((chapter) => chapter.id !== action.chapterId);
      const fallbackIndex = Math.max(0, targetIndex - 1);
      const fallbackChapter = chapters[fallbackIndex] ?? chapters[0];

      const next = {
        ...state,
        project: {
          ...state.project,
          chapters
        },
        currentChapterId: fallbackChapter.id,
        dirty: true,
        statusMessage: "Chapter deleted"
      };

      return updateProjectTimestamp(next);
    }
    case "SET_CURRENT_CHAPTER":
      return {
        ...state,
        currentChapterId: action.chapterId,
        view: "main",
        statusMessage: "Chapter selected"
      };
    case "APPLY_STRUCTURE": {
      const template = STRUCTURE_TEMPLATES[action.name.toLowerCase()];
      if (!template) {
        return {
          ...state,
          statusMessage: `Unknown structure '${action.name}'`
        };
      }

      const chapters = template.map((entry) => ({
        ...entry,
        id: generateId("chapter")
      }));

      const next: AppState = {
        ...state,
        view: "main",
        project: {
          ...state.project,
          chapters
        },
        currentChapterId: chapters[0].id,
        dirty: true,
        statusMessage: `Applied structure '${action.name}'`
      };

      return updateProjectTimestamp(next);
    }
    case "CREATE_WIKI_ENTRY": {
      const entry = {
        id: generateId("wiki"),
        title: action.title.trim() || "New Entry",
        content: ""
      };

      const next: AppState = {
        ...state,
        view: "wiki",
        project: {
          ...state.project,
          wiki: [...state.project.wiki, entry]
        },
        currentWikiId: entry.id,
        dirty: true,
        statusMessage: `Created wiki '${entry.title}'`
      };

      return updateProjectTimestamp(next);
    }
    case "RENAME_WIKI_ENTRY": {
      const wiki = state.project.wiki.map((entry) =>
        entry.id === action.wikiId
          ? {
              ...entry,
              title: action.title.trim() || entry.title
            }
          : entry
      );

      const next = {
        ...state,
        project: {
          ...state.project,
          wiki
        },
        dirty: true,
        statusMessage: "Wiki entry renamed"
      };

      return updateProjectTimestamp(next);
    }
    case "DELETE_WIKI_ENTRY": {
      if (state.project.wiki.length <= 1) {
        return {
          ...state,
          statusMessage: "Cannot delete the only wiki entry"
        };
      }

      const targetIndex = state.project.wiki.findIndex((entry) => entry.id === action.wikiId);
      if (targetIndex < 0) {
        return state;
      }

      const wiki = state.project.wiki.filter((entry) => entry.id !== action.wikiId);
      const fallbackIndex = Math.max(0, targetIndex - 1);
      const fallback = wiki[fallbackIndex] ?? wiki[0];

      const next = {
        ...state,
        project: {
          ...state.project,
          wiki
        },
        currentWikiId: fallback.id,
        dirty: true,
        statusMessage: "Wiki entry deleted"
      };

      return updateProjectTimestamp(next);
    }
    case "SET_CURRENT_WIKI":
      return {
        ...state,
        view: "wiki",
        currentWikiId: action.wikiId,
        statusMessage: "Wiki entry selected"
      };
    case "SET_WIKI_CONTENT": {
      const wiki = state.project.wiki.map((entry) =>
        entry.id === action.wikiId
          ? {
              ...entry,
              content: action.content
            }
          : entry
      );

      const next = {
        ...state,
        project: {
          ...state.project,
          wiki
        },
        dirty: true,
        statusMessage: "Wiki updated"
      };

      return updateProjectTimestamp(next);
    }
    case "UPDATE_SETTINGS":
      return {
        ...state,
        settings: {
          ...state.settings,
          ...action.settings
        }
      };
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
    case "SET_COMMAND_PALETTE_OPEN":
      return {
        ...state,
        commandPaletteOpen: action.open,
        statusMessage: action.open ? "Command palette ready" : state.statusMessage
      };
    default:
      return state;
  }
}
