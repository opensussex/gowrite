import { describe, expect, it } from "vitest";
import { LocalStorageRepository, type StorageLike } from "@/infra/local-storage-repo";
import { createDefaultProject, createDefaultSettings } from "@/domain/defaults";

class MemoryStorage implements StorageLike {
  private readonly map = new Map<string, string>();

  getItem(key: string): string | null {
    return this.map.get(key) ?? null;
  }

  setItem(key: string, value: string): void {
    this.map.set(key, value);
  }

  removeItem(key: string): void {
    this.map.delete(key);
  }
}

describe("LocalStorageRepository", () => {
  it("saves and loads current project", () => {
    const storage = new MemoryStorage();
    const repo = new LocalStorageRepository(storage);
    const project = createDefaultProject("Test Project");

    repo.saveProject(project);

    const loaded = repo.loadCurrentProject();
    expect(loaded?.id).toBe(project.id);
    expect(loaded?.name).toBe("Test Project");
  });

  it("saves and loads settings", () => {
    const storage = new MemoryStorage();
    const repo = new LocalStorageRepository(storage);
    const settings = createDefaultSettings("project_1");

    repo.saveSettings(settings);

    expect(repo.loadSettings()).toEqual(settings);
  });
});
