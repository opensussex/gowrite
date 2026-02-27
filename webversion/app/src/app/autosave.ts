import type { AppStore } from "./store";
import type { LocalStorageRepository } from "@/infra/local-storage-repo";

export interface AutosaveController {
  stop(): void;
  saveNow(): void;
}

interface AutosaveOptions {
  store: AppStore;
  repo: LocalStorageRepository;
  debounceMs?: number;
  heartbeatMs?: number;
  now?: () => number;
}

export function createAutosaveController({
  store,
  repo,
  debounceMs = 800,
  heartbeatMs = 60000,
  now = () => Date.now()
}: AutosaveOptions): AutosaveController {
  let debounceTimer: number | null = null;
  let isStopped = false;

  const save = (): void => {
    const state = store.getState();
    if (!state.dirty) {
      return;
    }

    try {
      repo.saveProject(state.project);
      repo.saveSettings(state.settings);
      store.dispatch({ type: "SAVE_SUCCEEDED", timestamp: now() });
    } catch (error) {
      const message = error instanceof Error ? error.message : "unknown error";
      store.dispatch({ type: "SAVE_FAILED", error: message });
    }
  };

  const unsubscribe = store.subscribe((state, action) => {
    if (isStopped) {
      return;
    }

    if (action.type === "SET_CHAPTER_CONTENT" && state.dirty) {
      if (debounceTimer !== null) {
        window.clearTimeout(debounceTimer);
      }
      debounceTimer = window.setTimeout(save, debounceMs);
    }
  });

  const heartbeat = window.setInterval(() => {
    if (isStopped) {
      return;
    }
    save();
  }, heartbeatMs);

  return {
    stop() {
      isStopped = true;
      if (debounceTimer !== null) {
        window.clearTimeout(debounceTimer);
      }
      window.clearInterval(heartbeat);
      unsubscribe();
    },
    saveNow() {
      save();
    }
  };
}
