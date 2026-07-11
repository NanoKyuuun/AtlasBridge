import { test, expect } from "@playwright/test";
import { mockAllAPIs, MOCK_STATUS, MOCK_CONFIG } from "./fixtures";

test.describe("Dashboard", () => {
  test.beforeEach(async ({ page }) => {
    await mockAllAPIs(page);
    await page.goto("/");
  });

  test("displays main heading", async ({ page }) => {
    await expect(page.getByText("AtlasBridge Control Center")).toBeVisible();
    await expect(page.getByText("Route every AI coding task to the right model path")).toBeVisible();
  });

  test("shows proxy status card", async ({ page }) => {
    await expect(page.getByText("Proxy Status")).toBeVisible();
    await expect(page.getByText("running")).toBeVisible();
  });

  test("shows OpenAI-compatible endpoint", async ({ page }) => {
    await expect(page.getByText("OpenAI-compatible endpoint")).toBeVisible();
    await expect(page.getByText("http://127.0.0.1:20127/v1")).toBeVisible();
  });

  test("shows downstream endpoint", async ({ page }) => {
    await expect(page.getByText("Downstream 9Router endpoint")).toBeVisible();
    await expect(page.getByText("http://127.0.0.1:20128/v1")).toBeVisible();
  });

  test("shows startup mode", async ({ page }) => {
    await expect(page.getByText("Startup mode")).toBeVisible();
    await expect(page.getByText("Auto start")).toBeVisible();
  });

  test("shows default route", async ({ page }) => {
    await expect(page.getByText("Default route")).toBeVisible();
    await expect(page.getByText("route.default")).toBeVisible();
  });

  test("shows architecture flow diagram", async ({ page }) => {
    await expect(page.getByText("Architecture Flow")).toBeVisible();
    await expect(page.getByText("AtlasBridge")).toBeVisible();
    await expect(page.getByText("9Router")).toBeVisible();
    await expect(page.getByText("AI Providers")).toBeVisible();
  });

  test("shows model aliases table", async ({ page }) => {
    await expect(page.getByText("Model Aliases")).toBeVisible();
    await expect(page.getByText("atlas-auto")).toBeVisible();
    await expect(page.getByText("smart-auto")).toBeVisible();
  });

  test("shows combo tester section", async ({ page }) => {
    await expect(page.getByText("Combo Tester")).toBeVisible();
    await expect(page.getByPlaceholder("Model name")).toBeVisible();
  });

  test("combo tester presets are clickable", async ({ page }) => {
    const preset = page.getByRole("button", { name: "combo.default" });
    await preset.click();
    await expect(page.getByPlaceholder("Model name")).toHaveValue("combo.default");
  });

  test("combo test shows result", async ({ page }) => {
    await page.getByPlaceholder("Model name").fill("combo.default");
    await page.getByRole("button", { name: "Test Combo" }).click();
    await expect(page.getByText("gpt-4o")).toBeVisible();
    await expect(page.getByText("245ms")).toBeVisible();
  });

  test("start proxy button works", async ({ page }) => {
    const startBtn = page.getByRole("button", { name: "Start Proxy" }).first();
    await startBtn.click();
    // Button should still be visible after click (no error)
    await expect(startBtn).toBeVisible();
  });

  test("quick actions section exists", async ({ page }) => {
    await expect(page.getByText("Quick Actions")).toBeVisible();
    await expect(page.getByRole("button", { name: "Stop Proxy" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Restart Proxy" })).toBeVisible();
  });
});
