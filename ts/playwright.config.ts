import { defineConfig, devices } from "@playwright/test";

// Harnais UX Matrix Gate : le renderer générique est prouvé une fois sur les
// trois classes de device imposées (desktop + iPhone + iPad). Chaque plugin
// ajoute ensuite ses propres snapshots.
export default defineConfig({
  testDir: "./e2e",
  snapshotDir: "./e2e/__snapshots__",
  reporter: [["html", { open: "never" }], ["list"]],
  use: { screenshot: "on" },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"], viewport: { width: 1280, height: 800 } } },
    { name: "iphone", use: { ...devices["iPhone 13"] } },
    { name: "ipad", use: { ...devices["iPad Pro 11"] } },
  ],
});
