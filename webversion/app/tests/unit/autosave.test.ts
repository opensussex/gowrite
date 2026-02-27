import { describe, expect, it, vi } from "vitest";
import { AppStore } from "@/app/store";
import { appReducer } from "@/app/reducer";
import { createInitialState } from "@/app/initial-state";
import { createAutosaveController } from "@/app/autosave";
import { LocalStorageRepository, type StorageLike } from "@/infra/local-storage-repo";

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

describe("autosave", () => {
  it("persists dirty edits after debounce", () => {
    vi.useFakeTimers();

    const store = new AppStore(createInitialState(), appReducer);
    const repo = new LocalStorageRepository(new MemoryStorage());
    const controller = createAutosaveController({
      store,
      repo,
      debounceMs: 10,
      heartbeatMs: 60000,
      now: () => 1700000000000
    });

    store.dispatch({
      type: "SET_CHAPTER_CONTENT",
      chapterId: store.getState().currentChapterId,
      content: "Autosave me"
    });

    vi.advanceTimersByTime(20);

    expect(store.getState().dirty).toBe(false);
    expect(store.getState().lastSavedAt).toBe(1700000000000);

    controller.stop();
    vi.useRealTimers();
  });
});
