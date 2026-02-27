import { describe, expect, it } from "vitest";
import { appReducer } from "@/app/reducer";
import { createInitialState } from "@/app/initial-state";

describe("appReducer", () => {
  it("marks state dirty when chapter content changes", () => {
    const initial = createInitialState();
    const chapterId = initial.currentChapterId;

    const next = appReducer(initial, {
      type: "SET_CHAPTER_CONTENT",
      chapterId,
      content: "A first draft paragraph"
    });

    expect(next.dirty).toBe(true);
    expect(next.project.chapters[0].content).toContain("first draft");
  });

  it("clears dirty flag on save success", () => {
    const initial = createInitialState();
    const dirty = appReducer(initial, {
      type: "SET_CHAPTER_CONTENT",
      chapterId: initial.currentChapterId,
      content: "Dirty"
    });

    const saved = appReducer(dirty, {
      type: "SAVE_SUCCEEDED",
      timestamp: 1700000000000
    });

    expect(saved.dirty).toBe(false);
    expect(saved.lastSavedAt).toBe(1700000000000);
  });
});
