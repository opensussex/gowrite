import { expect, test } from "@playwright/test";

async function runCommand(page: import("@playwright/test").Page, command: string): Promise<void> {
  await page.keyboard.press("Control+k");
  const input = page.getByLabel("Command input");
  await expect(input).toBeVisible();
  await input.fill(command);
  await input.press("Enter");
}

test("chapter lifecycle via command palette", async ({ page }) => {
  await page.goto("/");

  await runCommand(page, 'chapter new "Act One"');
  await expect(page.getByText(/CHAPTER 2\/2/)).toBeVisible();

  await runCommand(page, 'chapter rename 2 "Act I"');
  await runCommand(page, "chapter delete 2");

  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  await page.getByRole("button", { name: "OK" }).click();

  await expect(page.getByText(/CHAPTER 1\/1/)).toBeVisible();
});

test("project save/load lifecycle", async ({ page }) => {
  await page.goto("/");

  const editor = page.getByLabel("Chapter editor");
  await editor.fill("Alpha project text");
  await runCommand(page, 'save "Project Alpha"');

  await editor.fill("Beta project text");
  await runCommand(page, 'save "Project Beta"');

  await runCommand(page, 'open "Project Alpha"');
  await expect(editor).toHaveValue("Alpha project text");
});

test("help command shows command guide", async ({ page }) => {
  await page.goto("/");

  await runCommand(page, "help chapter");
  await expect(page.getByRole("dialog")).toBeVisible();
  await expect(page.getByText(/chapter new/i)).toBeVisible();
  await page.getByRole("button", { name: "OK" }).click();
});
