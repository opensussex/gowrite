import type { AppState } from "@/domain/models";

export function getCurrentChapter(state: AppState) {
  return state.project.chapters.find((chapter) => chapter.id === state.currentChapterId);
}

export function getCurrentChapterIndex(state: AppState): number {
  return state.project.chapters.findIndex((chapter) => chapter.id === state.currentChapterId);
}

export function getChapterByIndex(state: AppState, index: number) {
  if (index < 0 || index >= state.project.chapters.length) {
    return undefined;
  }
  return state.project.chapters[index];
}

export function getCurrentWikiEntry(state: AppState) {
  return state.project.wiki.find((entry) => entry.id === state.currentWikiId);
}

export function getCurrentWikiIndex(state: AppState): number {
  return state.project.wiki.findIndex((entry) => entry.id === state.currentWikiId);
}

export function getWikiByIndex(state: AppState, index: number) {
  if (index < 0 || index >= state.project.wiki.length) {
    return undefined;
  }
  return state.project.wiki[index];
}

export function getWordCount(text: string): number {
  const trimmed = text.trim();
  if (!trimmed) {
    return 0;
  }
  return trimmed.split(/\s+/).length;
}
