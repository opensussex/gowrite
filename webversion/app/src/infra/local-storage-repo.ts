import type { AppSettings, Project } from "@/domain/models";

const SETTINGS_KEY = "gowrite.settings";
const CURRENT_PROJECT_KEY = "gowrite.currentProjectId";
const PROJECT_INDEX_KEY = "gowrite.projectIndex";
const PROJECT_KEY_PREFIX = "gowrite.project.";

interface ProjectIndex {
  ids: string[];
}

export interface StorageLike {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
}

function parseJson<T>(raw: string | null): T | null {
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

export class LocalStorageRepository {
  constructor(private readonly storage: StorageLike = window.localStorage) {}

  saveSettings(settings: AppSettings): void {
    this.storage.setItem(SETTINGS_KEY, JSON.stringify(settings));
  }

  loadSettings(): AppSettings | null {
    return parseJson<AppSettings>(this.storage.getItem(SETTINGS_KEY));
  }

  saveProject(project: Project): void {
    const key = `${PROJECT_KEY_PREFIX}${project.id}`;
    this.storage.setItem(key, JSON.stringify(project));
    this.storage.setItem(CURRENT_PROJECT_KEY, project.id);

    const index = this.loadProjectIndex();
    if (!index.ids.includes(project.id)) {
      index.ids.push(project.id);
      this.storage.setItem(PROJECT_INDEX_KEY, JSON.stringify(index));
    }
  }

  loadCurrentProject(): Project | null {
    const currentProjectId = this.storage.getItem(CURRENT_PROJECT_KEY);
    if (!currentProjectId) {
      return null;
    }
    return this.loadProject(currentProjectId);
  }

  loadProject(projectId: string): Project | null {
    const key = `${PROJECT_KEY_PREFIX}${projectId}`;
    return parseJson<Project>(this.storage.getItem(key));
  }

  loadProjectIndex(): ProjectIndex {
    return parseJson<ProjectIndex>(this.storage.getItem(PROJECT_INDEX_KEY)) ?? { ids: [] };
  }
}
