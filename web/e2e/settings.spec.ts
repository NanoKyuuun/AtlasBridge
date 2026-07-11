import { test, expect } from "@playwright/test";
import { mockAllAPIs, MOCK_CONFIG } from "./fixtures";

test.describe("Privacy Settings", () => {
  test.beforeEach(async ({ page }) => {
    await mockAllAPIs(page);
    await page.goto("/privacy");
  });

  test("displays page heading", async ({ page }) => {
    await expect(page.getByText("Privacy & Logging")).toBeVisible();
    await expect(page.getByText("Configure privacy mode and logging behavior")).toBeVisible();
  });

  test("shows three privacy mode options", async ({ page }) => {
    await expect(page.getByText("Standard")).toBeVisible();
    await expect(page.getByText("Strict")).toBeVisible();
    await expect(page.getByText("Debug")).toBeVisible();
  });

  test("standard mode is selected by default", async ({ page }) => {
    const standardCard = page.locator("label").filter({ hasText: "Standard" });
    await expect(standardCard.locator("input")).toBeChecked();
  });

  test("can switch to strict mode", async ({ page }) => {
    const strictCard = page.locator("label").filter({ hasText: "Strict" });
    await strictCard.click();
    await expect(strictCard.locator("input")).toBeChecked();
  });

  test("debug mode shows warning", async ({ page }) => {
    const debugCard = page.locator("label").filter({ hasText: "Debug" });
    await debugCard.click();
    await expect(page.getByText("Debug mode increases logging verbosity")).toBeVisible();
  });

  test("shows logging controls section", async ({ page }) => {
    await expect(page.getByText("Logging Controls")).toBeVisible();
    await expect(page.getByText("Metadata logging")).toBeVisible();
    await expect(page.getByText("Full prompt logging")).toBeVisible();
    await expect(page.getByText("Secret redaction")).toBeVisible();
  });

  test("shows security snapshot section", async ({ page }) => {
    await expect(page.getByText("Security Snapshot")).toBeVisible();
    await expect(page.getByText("Privacy mode")).toBeVisible();
  });

  test("save button is disabled when no changes", async ({ page }) => {
    const saveBtn = page.getByRole("button", { name: "Save Changes" });
    await expect(saveBtn).toBeDisabled();
  });

  test("save button enables after change", async ({ page }) => {
    const strictCard = page.locator("label").filter({ hasText: "Strict" });
    await strictCard.click();
    const saveBtn = page.getByRole("button", { name: "Save Changes" });
    await expect(saveBtn).toBeEnabled();
  });

  test("retention days input exists", async ({ page }) => {
    await expect(page.getByText("Retention days")).toBeVisible();
    const retentionInput = page.locator("input[type='number']").filter({ hasText: "" }).last();
    await expect(retentionInput).toBeVisible();
  });
});

test.describe("Downstream Settings", () => {
  test.beforeEach(async ({ page }) => {
    await mockAllAPIs(page);
    await page.goto("/downstream");
  });

  test("displays page heading", async ({ page }) => {
    await expect(page.getByText("9Router Downstream")).toBeVisible();
  });

  test("shows connection settings", async ({ page }) => {
    await expect(page.getByText("Connection Settings")).toBeVisible();
    await expect(page.getByText("Base URL")).toBeVisible();
    await expect(page.getByText("Timeout")).toBeVisible();
  });

  test("shows downstream URL input", async ({ page }) => {
    const urlInput = page.locator("input").filter({ hasText: "" }).first();
    await expect(urlInput).toBeVisible();
  });

  test("shows connection status section", async ({ page }) => {
    await expect(page.getByText("Connection Status")).toBeVisible();
    await expect(page.getByText("Check Connection")).toBeVisible();
  });

  test("check connection button works", async ({ page }) => {
    await page.getByRole("button", { name: "Check Connection" }).click();
    await expect(page.getByText("connected")).toBeVisible();
  });

  test("save button disabled initially", async ({ page }) => {
    const saveBtn = page.getByRole("button", { name: "Save Changes" });
    await expect(saveBtn).toBeDisabled();
  });
});

test.describe("Routing Settings", () => {
  test.beforeEach(async ({ page }) => {
    await mockAllAPIs(page);
    await page.goto("/routing");
  });

  test("displays page heading", async ({ page }) => {
    await expect(page.getByText("Task-to-Route Mapping")).toBeVisible();
  });

  test("shows auto-routing toggle", async ({ page }) => {
    await expect(page.getByText("Automatic task routing")).toBeVisible();
  });

  test("shows task-to-route mapping table", async ({ page }) => {
    await expect(page.getByText("Task-to-Route Mapping")).toBeVisible();
    await expect(page.getByText("backend_engineering")).toBeVisible();
    await expect(page.getByText("frontend_engineering")).toBeVisible();
    await expect(page.getByText("debugging")).toBeVisible();
  });

  test("shows routing defaults section", async ({ page }) => {
    await expect(page.getByText("Routing Defaults")).toBeVisible();
    await expect(page.getByText("Default Route")).toBeVisible();
    await expect(page.getByText("Low Confidence Route")).toBeVisible();
  });

  test("shows confidence threshold", async ({ page }) => {
    await expect(page.getByText("Confidence Threshold")).toBeVisible();
    await expect(page.getByText("Threshold")).toBeVisible();
  });

  test("reset all button exists", async ({ page }) => {
    await expect(page.getByRole("button", { name: "Reset to Default" })).toBeVisible();
  });

  test("discard button exists", async ({ page }) => {
    await expect(page.getByRole("button", { name: "Discard" })).toBeVisible();
  });
});
