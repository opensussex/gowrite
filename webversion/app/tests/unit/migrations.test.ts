import { describe, expect, it } from "vitest";
import { migrateProject, migrateSettings } from "@/infra/migrations";
import { createDefaultSettings } from "@/domain/defaults";

describe("migrations", () => {
  it("migrates Go-style project shape", () => {
    const raw = {
      Chapters: [
        {
          Title: "Opening",
          Content: "Hello",
          Notes: "N",
          Target: 10
        }
      ],
      Wiki: [{ Title: "Lore", Content: "World" }]
    };

    const project = migrateProject(raw);
    expect(project).not.toBeNull();
    expect(project?.chapters[0].title).toBe("Opening");
    expect(project?.wiki[0].title).toBe("Lore");
  });

  it("migrates legacy chapters array", () => {
    const legacy = [{ Title: "Legacy One", Content: "Text" }];
    const project = migrateProject(legacy);

    expect(project).not.toBeNull();
    expect(project?.chapters).toHaveLength(1);
    expect(project?.wiki).toHaveLength(1);
  });

  it("fills missing settings fields from fallback", () => {
    const fallback = createDefaultSettings("project_x");
    const migrated = migrateSettings({ theme: "retro", centered: true }, fallback);

    expect(migrated.theme).toBe("retro");
    expect(migrated.centered).toBe(true);
    expect(migrated.typewriterMode).toBe(true);
    expect(migrated.lastProjectId).toBe("project_x");
  });
});
