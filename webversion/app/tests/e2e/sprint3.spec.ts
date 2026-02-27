import { expect, test } from "@playwright/test";

async function runCommand(page: import("@playwright/test").Page, command: string): Promise<void> {
  await page.keyboard.press("Control+k");
  const input = page.getByLabel("Command input");
  await expect(input).toBeVisible();
  await input.fill(command);
  await input.press("Enter");
}

test("notes view stores per-chapter notes", async ({ page }) => {
  await page.goto("/");

  await runCommand(page, "notes");
  const notesEditor = page.getByLabel("Chapter notes editor");
  await expect(notesEditor).toBeVisible();

  await notesEditor.fill("Remember to foreshadow the hidden map.");

  await runCommand(page, "notes");
  await expect(page.getByLabel("Chapter editor")).toBeVisible();

  await runCommand(page, "notes");
  await expect(notesEditor).toHaveValue("Remember to foreshadow the hidden map.");
});

test("wiki command lifecycle", async ({ page }) => {
  await page.goto("/");

  await runCommand(page, "wiki");
  const wikiEditor = page.getByLabel("Wiki entry editor");
  await expect(wikiEditor).toBeVisible();

  await runCommand(page, 'wiki new "Lore"');
  await expect(page.getByRole("button", { name: "Lore" })).toBeVisible();

  await wikiEditor.fill("Ancient empire collapsed after the eclipse.");
  await runCommand(page, 'wiki rename 2 "World Lore"');
  await expect(page.getByRole("button", { name: "World Lore" })).toBeVisible();

  await runCommand(page, "wiki delete 2");
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.getByRole("button", { name: "Yes" }).click();
  await expect(page.getByRole("button", { name: "World Lore" })).toHaveCount(0);
});

test("structure command applies chapter template", async ({ page }) => {
  await page.goto("/");

  await runCommand(page, "structure hero");
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.getByRole("button", { name: "Yes" }).click();

  await expect(page.getByText(/CHAPTER\s+1\/[2-9]/)).toBeVisible();
  await expect(page.getByText(/CHAPTER\s+:\s+ORDINARY WORLD/i)).toBeVisible();
});

test("chapters command opens modal list and supports arrow selection", async ({ page }) => {
  await page.goto("/");

  await runCommand(page, 'chapter new "Act One"');
  await runCommand(page, 'chapter new "Act Two"');
  await expect(page.getByText(/CHAPTER\s+3\/3/)).toBeVisible();

  await runCommand(page, "chapters");
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.keyboard.press("ArrowUp");
  await page.keyboard.press("Enter");

  await expect(page.getByText(/CHAPTER\s+2\/3/)).toBeVisible();
  await expect(page.getByText(/CHAPTER\s+:\s+ACT ONE/i)).toBeVisible();
});
