import type { AppSettings, Project } from "@/domain/models";
import { migrateProject, migrateSettings } from "./migrations";

const SETTINGS_KEY = "gowrite.settings";
const CURRENT_PROJECT_KEY = "gowrite.currentProjectId";
const PROJECT_INDEX_KEY = "gowrite.projectIndex";
const PROJECT_KEY_PREFIX = "gowrite.project.";

interface ProjectIndex {
  ids: string[];
}

export interface ProjectSummary {
  id: string;
  name: string;
  updatedAt: string;
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

  loadSettings(fallback?: AppSettings): AppSettings | null {
    const raw = parseJson<unknown>(this.storage.getItem(SETTINGS_KEY));
    if (!fallback) {
      return raw as AppSettings | null;
    }
    if (raw === null) {
      return fallback;
    }
    return migrateSettings(raw, fallback);
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

  setCurrentProjectId(projectId: string): void {
    this.storage.setItem(CURRENT_PROJECT_KEY, projectId);
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
    const raw = parseJson<unknown>(this.storage.getItem(key));
    if (raw === null) {
      return null;
    }
    return migrateProject(raw);
  }

  loadProjectIndex(): ProjectIndex {
    const raw = parseJson<ProjectIndex>(this.storage.getItem(PROJECT_INDEX_KEY));
    if (!raw || !Array.isArray(raw.ids)) {
      return { ids: [] };
    }

    const ids = raw.ids.filter((id) => typeof id === "string" && id.trim().length > 0);
    return { ids };
  }

  listProjects(): ProjectSummary[] {
    const summaries: ProjectSummary[] = [];
    const index = this.loadProjectIndex();

    index.ids.forEach((id) => {
      const project = this.loadProject(id);
      if (!project) {
        return;
      }
      summaries.push({
        id: project.id,
        name: project.name,
        updatedAt: project.updatedAt
      });
    });

    return summaries.sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
  }

  findProject(query: string): Project | null {
    const normalized = query.trim().toLowerCase();
    if (!normalized) {
      return null;
    }

    const direct = this.loadProject(query.trim());
    if (direct) {
      return direct;
    }

    const summaries = this.listProjects();
    const exact = summaries.find((summary) => summary.name.toLowerCase() === normalized);
    if (exact) {
      return this.loadProject(exact.id);
    }

    const fuzzy = summaries.find((summary) => summary.name.toLowerCase().includes(normalized));
    if (fuzzy) {
      return this.loadProject(fuzzy.id);
    }

    return null;
  }
}
