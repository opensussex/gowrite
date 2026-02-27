import type { AppState } from "@/domain/models";
import { createDefaultProject, createDefaultSettings } from "@/domain/defaults";

export function createInitialState(): AppState {
  const project = createDefaultProject();
  const settings = createDefaultSettings(project.id);

  return {
    project,
    settings,
    view: "main",
    currentChapterId: project.chapters[0].id,
    currentWikiId: project.wiki[0].id,
    dirty: false,
    lastSavedAt: null,
    cursor: { row: 1, column: 1 },
    commandPaletteOpen: false,
    statusMessage: "Ready"
  };
}
