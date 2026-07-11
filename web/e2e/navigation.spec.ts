import { test, expect } from "@playwright/test";
import { mockAllAPIs } from "./fixtures";

test.describe("Navigation", () => {
  test.beforeEach(async ({ page }) => {
    await mockAllAPIs(page);
    await page.goto("/");
  });

  test("sidebar shows all navigation items", async ({ page }) => {
    const sidebar = page.locator("aside").first();
    await expect(sidebar.getByText("Dashboard")).toBeVisible();
    await expect(sidebar.getByText("Routing Settings")).toBeVisible();
    await expect(sidebar.getByText("Route Profiles")).toBeVisible();
    await expect(sidebar.getByText("Runtime")).toBeVisible();
    await expect(sidebar.getByText("Startup")).toBeVisible();
    await expect(sidebar.getByText("9Router")).toBeVisible();
    await expect(sidebar.getByText("Logs")).toBeVisible();
    await expect(sidebar.getByText("Privacy")).toBeVisible();
    await expect(sidebar.getByText("Advanced")).toBeVisible();
  });

  test("clicking routing settings navigates to routing page", async ({ page }) => {
    await page.getByRole("link", { name: "Routing Settings" }).first().click();
    await expect(page).toHaveURL(/\/routing/);
    await expect(page.getByText("Task-to-Route Mapping")).toBeVisible();
  });

  test("clicking privacy navigates to privacy page", async ({ page }) => {
    await page.getByRole("link", { name: "Privacy" }).first().click();
    await expect(page).toHaveURL(/\/privacy/);
    await expect(page.getByText("Privacy & Logging")).toBeVisible();
  });

  test("clicking downstream navigates to downstream page", async ({ page }) => {
    await page.getByRole("link", { name: "9Router" }).first().click();
    await expect(page).toHaveURL(/\/downstream/);
    await expect(page.getByText("9Router Downstream")).toBeVisible();
  });

  test("clicking advanced navigates to advanced page", async ({ page }) => {
    await page.getByRole("link", { name: "Advanced" }).first().click();
    await expect(page).toHaveURL(/\/advanced/);
  });

  test("clicking logs navigates to logs page", async ({ page }) => {
    await page.getByRole("link", { name: "Logs" }).first().click();
    await expect(page).toHaveURL(/\/logs/);
  });

  test("dashboard link navigates back to dashboard", async ({ page }) => {
    await page.getByRole("link", { name: "Routing Settings" }).first().click();
    await expect(page).toHaveURL(/\/routing/);
    await page.getByRole("link", { name: "Dashboard" }).first().click();
    await expect(page).toHaveURL(/\/admin\/$/);
  });

  test("active nav item is highlighted", async ({ page }) => {
    await page.getByRole("link", { name: "Privacy" }).first().click();
    const privacyLink = page.locator("aside").first().getByRole("link", { name: "Privacy" });
    await expect(privacyLink).toHaveClass(/bg-gradient-to-r/);
  });

  test("header shows current page title", async ({ page }) => {
    await page.getByRole("link", { name: "Privacy" }).first().click();
    await expect(page.locator("header").getByText("Privacy")).toBeVisible();
  });
});
