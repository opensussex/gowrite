import { test, expect } from "@playwright/test";

test("persists writing across reload", async ({ page }) => {
  await page.goto("/");

  const editor = page.getByLabel("Chapter editor");
  const text = "Sprint 1 baseline persistence text.";

  await editor.fill(text);
  await page.waitForTimeout(1200);
  await page.reload();

  await expect(editor).toHaveValue(text);
});
