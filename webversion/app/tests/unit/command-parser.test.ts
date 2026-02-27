import { describe, expect, it } from "vitest";
import { parseCommandInput, tokenizeCommand } from "@/app/command-parser";

describe("command parser", () => {
  it("tokenizes quoted arguments", () => {
    const tokens = tokenizeCommand('chapter rename 2 "Act One"');
    expect(tokens).toEqual(["chapter", "rename", "2", "Act One"]);
  });

  it("parses command and args", () => {
    const parsed = parseCommandInput("save MyProject");
    expect(parsed).toEqual({
      name: "save",
      args: ["MyProject"],
      raw: "save MyProject"
    });
  });

  it("returns null for empty command", () => {
    expect(parseCommandInput("   ")).toBeNull();
  });
});
