import { describe, expect, it } from "vitest";
import { getHelpTopics, renderHelp, resolveHelpTopic } from "@/app/help-registry";

describe("help registry", () => {
  it("includes shipped command topics", () => {
    const topics = getHelpTopics();
    const ids = topics.map((topic) => topic.id);
    expect(ids).toContain("save");
    expect(ids).toContain("open");
    expect(ids).toContain("chapter");
    expect(ids).toContain("chapters");
    expect(ids).toContain("notes");
    expect(ids).toContain("wiki");
    expect(ids).toContain("structure");
    expect(ids).toContain("wordcount");
    expect(ids).toContain("help");
  });

  it("includes shipped feature topics", () => {
    const topics = getHelpTopics();
    const ids = topics.map((topic) => topic.id);
    expect(ids).toContain("autosave");
    expect(ids).toContain("typewriter");
    expect(ids).toContain("shortcuts");
  });

  it("renders overview and topic details", () => {
    const overview = renderHelp([]);
    expect(overview.message).toContain("Commands");

    const chapterHelp = renderHelp(["chapter"]);
    expect(chapterHelp.title).toContain("chapter");
    expect(chapterHelp.message).toContain("chapter new");
  });

  it("resolves aliases", () => {
    const topic = resolveHelpTopic("load");
    expect(topic?.id).toBe("open");
  });

  it("resolves chapter list alias", () => {
    const topic = resolveHelpTopic("list");
    expect(topic?.id).toBe("chapters");
  });

  it("renders structure topic details", () => {
    const structureHelp = renderHelp(["structure"]);
    expect(structureHelp.message).toContain("3act");
    expect(structureHelp.message).toContain("replaces current chapters");
  });
});
