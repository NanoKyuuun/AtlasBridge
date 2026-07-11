import { test, expect } from "@playwright/test";
import { mockAllAPIs, MOCK_STATUS } from "./fixtures";

test.describe("Login Flow", () => {
  test.beforeEach(async ({ page }) => {
    await mockAllAPIs(page, { authRequired: true });
  });

  test("redirects to login when auth is required", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(/\/login/);
    await expect(page.getByText("Admin Authentication")).toBeVisible();
  });

  test("shows token input field", async ({ page }) => {
    await page.goto("/login");
    await expect(page.locator("#token-input")).toBeVisible();
    await expect(page.getByPlaceholder("Paste your admin token here")).toBeVisible();
  });

  test("unlock button is disabled when token is empty", async ({ page }) => {
    await page.goto("/login");
    const button = page.getByRole("button", { name: "Unlock Dashboard" });
    await expect(button).toBeDisabled();
  });

  test("shows error on invalid token", async ({ page }) => {
    await page.goto("/login");
    await page.locator("#token-input").fill("wrong-token");
    await page.getByRole("button", { name: "Unlock Dashboard" }).click();
    await expect(page.getByText("Invalid token")).toBeVisible();
  });

  test("successful login redirects to dashboard", async ({ page }) => {
    await page.goto("/login");
    await page.locator("#token-input").fill("test-token-123");
    await page.getByRole("button", { name: "Unlock Dashboard" }).click();
    await expect(page).toHaveURL(/\/admin\/$/);
  });

  test("shows loading state during verification", async ({ page }) => {
    await page.goto("/login");
    await page.locator("#token-input").fill("test-token-123");
    await page.getByRole("button", { name: "Unlock Dashboard" }).click();
    await expect(page.getByText("Verifying...")).toBeVisible();
  });
});

test.describe("Login - No Auth Required", () => {
  test("bypasses login when auth is not required", async ({ page }) => {
    await mockAllAPIs(page, { authRequired: false });
    await page.goto("/");
    await expect(page).toHaveURL(/\/admin\/$/);
    await expect(page.getByText("AtlasBridge Control Center")).toBeVisible();
  });
});
