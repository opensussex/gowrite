import type { AppState } from "@/domain/models";

export function getCurrentChapter(state: AppState) {
  return state.project.chapters.find((chapter) => chapter.id === state.currentChapterId);
}

export function getWordCount(text: string): number {
  const trimmed = text.trim();
  if (!trimmed) {
    return 0;
  }
  return trimmed.split(/\s+/).length;
}
