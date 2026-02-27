import { describe, expect, it, vi } from "vitest";
import { AppStore } from "@/app/store";
import { appReducer } from "@/app/reducer";
import { createInitialState } from "@/app/initial-state";
import { CommandBus, type CommandContext } from "@/app/command-bus";
import { registerCoreCommands } from "@/app/commands/core";
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

function createHarness() {
  const store = new AppStore(createInitialState(), appReducer);
  const repo = new LocalStorageRepository(new MemoryStorage());
  const bus = new CommandBus();
  registerCoreCommands(bus);

  const modalSpy = vi.fn(async () => {
    return;
  });
  const confirmSpy = vi.fn(async () => true);

  const context: CommandContext = {
    getState: () => store.getState(),
    dispatch: (action) => store.dispatch(action),
    repo,
    persistCurrentProject: (force = false) => {
      const state = store.getState();
      if (!force && !state.dirty) {
        return;
      }
      repo.saveProject(state.project);
      repo.saveSettings(state.settings);
      store.dispatch({ type: "SAVE_SUCCEEDED", timestamp: 1700000000000 });
    },
    showModal: modalSpy,
    confirm: confirmSpy
  };

  return { store, repo, bus, context, modalSpy, confirmSpy };
}

describe("core commands", () => {
  it("creates, renames, and deletes chapters", async () => {
    const { bus, context, store, confirmSpy } = createHarness();

    await bus.execute('chapter new "Act One"', context);
    expect(store.getState().project.chapters).toHaveLength(2);

    await bus.execute('chapter rename 2 "Act I"', context);
    expect(store.getState().project.chapters[1].title).toBe("Act I");

    await bus.execute("chapter delete 2", context);
    expect(confirmSpy).toHaveBeenCalled();
    expect(store.getState().project.chapters).toHaveLength(1);
  });

  it("saves as new project and reopens by name", async () => {
    const { bus, context, store, repo } = createHarness();

    store.dispatch({
      type: "SET_CHAPTER_CONTENT",
      chapterId: store.getState().currentChapterId,
      content: "Alpha text"
    });

    await bus.execute('save "Project Alpha"', context);
    expect(store.getState().project.name).toBe("Project Alpha");

    store.dispatch({
      type: "SET_CHAPTER_CONTENT",
      chapterId: store.getState().currentChapterId,
      content: "Beta text"
    });

    await bus.execute('save "Project Beta"', context);
    expect(store.getState().project.name).toBe("Project Beta");

    await bus.execute('open "Project Alpha"', context);
    expect(store.getState().project.name).toBe("Project Alpha");
    expect(repo.listProjects().length).toBeGreaterThanOrEqual(2);
  });

  it("shows wordcount modal", async () => {
    const { bus, context, store, modalSpy } = createHarness();

    store.dispatch({
      type: "SET_CHAPTER_CONTENT",
      chapterId: store.getState().currentChapterId,
      content: "one two three"
    });

    const result = await bus.execute("wordcount", context);

    expect(result.ok).toBe(true);
    expect(modalSpy).toHaveBeenCalled();
  });

  it("shows help overview and topic details", async () => {
    const { bus, context, modalSpy } = createHarness();

    const helpOverview = await bus.execute("help", context);
    const helpTopic = await bus.execute("help chapter", context);

    expect(helpOverview.ok).toBe(true);
    expect(helpTopic.ok).toBe(true);
    expect(modalSpy).toHaveBeenCalledTimes(2);
    expect(modalSpy).toHaveBeenNthCalledWith(1, expect.stringContaining("Help"), expect.any(String));
    expect(modalSpy).toHaveBeenNthCalledWith(2, expect.any(String), expect.stringContaining("chapter new"));
  });

  it("toggles notes view", async () => {
    const { bus, context, store } = createHarness();

    await bus.execute("notes", context);
    expect(store.getState().view).toBe("notes");

    await bus.execute("notes", context);
    expect(store.getState().view).toBe("main");
  });

  it("creates, renames, and deletes wiki entries", async () => {
    const { bus, context, store, confirmSpy } = createHarness();

    await bus.execute('wiki new "Villain"', context);
    expect(store.getState().project.wiki).toHaveLength(2);

    await bus.execute('wiki rename 2 "Antagonist"', context);
    expect(store.getState().project.wiki[1].title).toBe("Antagonist");

    await bus.execute("wiki delete 2", context);
    expect(confirmSpy).toHaveBeenCalled();
    expect(store.getState().project.wiki).toHaveLength(1);
  });

  it("applies structure template after confirmation", async () => {
    const { bus, context, store, confirmSpy } = createHarness();

    const beforeCount = store.getState().project.chapters.length;
    const result = await bus.execute("structure 3act", context);

    expect(result.ok).toBe(true);
    expect(confirmSpy).toHaveBeenCalled();
    expect(store.getState().project.chapters.length).toBeGreaterThan(beforeCount);
    expect(store.getState().project.chapters[0].title).toContain("Act");
  });
});
