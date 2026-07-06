import { test, expect } from "@playwright/test";

// Harnais UX Matrix : rendu du renderer générique sur desktop/iphone/ipad.
// Chaque projet Playwright rejoue ce test ; les captures alimentent la gate
// de non-régression visuelle. Un plugin fournit son fixture HTML de démo.
//
// NB: nécessite un serveur de démo (`ui/demo.html`) servant le renderer avec
// un descripteur + reading figés. Squelette volontairement minimal.

test("tuile plugin — visuel stable", async ({ page }) => {
  await page.goto("/demo.html");
  const tile = page.locator('[data-plugin="sungrow-solar"]');
  await expect(tile).toBeVisible();
  await expect(tile).toHaveScreenshot();
});

test("état obsolète signalé", async ({ page }) => {
  await page.goto("/demo.html?stale=1");
  await expect(page.locator(".ess-plugin__stale")).toBeVisible();
});

test("aucune mutation armoire émise", async ({ page }) => {
  const blocked: string[] = [];
  page.on("request", (r) => {
    if (/inject|\/scenarios\/.*\/launch|\/api\/web\/actions/.test(r.url()) && r.method() !== "GET") {
      blocked.push(r.url());
    }
  });
  await page.goto("/demo.html");
  expect(blocked).toEqual([]);
});
