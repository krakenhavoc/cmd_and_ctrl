import { expect, test } from "@playwright/test";
import { ADMIN_TOKEN } from "./env";

test.describe("admin login", () => {
  test.beforeEach(async ({ page }) => {
    // Start every test from a clean session so the lobby redirect
    // doesn't skip us past the login form.
    await page.goto("/");
    await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
    await page.goto("/#/login");
  });

  test("wrong token surfaces a visible error", async ({ page }) => {
    await page.getByPlaceholder("admin token").fill("not-the-real-token-but-long");
    await page.getByRole("button", { name: "log in" }).click();
    await expect(page.getByRole("alert")).toBeVisible();
    // We stay on the login page — no redirect to lobby.
    await expect(page).not.toHaveURL(/#\/lobby/);
  });

  test("valid token lands on the lobby", async ({ page }) => {
    await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
    await page.getByRole("button", { name: "log in" }).click();
    await expect(page).toHaveURL(/#\/lobby$/);
    await expect(page.getByRole("heading", { name: /cmd_and_ctrl · lobby/ })).toBeVisible();
    // Admin-only UI surfaces — create-game form should be present.
    await expect(page.getByRole("heading", { name: "create game" })).toBeVisible();
  });

  test("session persists across reload", async ({ page }) => {
    await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
    await page.getByRole("button", { name: "log in" }).click();
    await expect(page).toHaveURL(/#\/lobby$/);

    await page.reload();
    // App.svelte's auth gate kicks /login → /lobby when a session is
    // live. We should land on the lobby without seeing the login form.
    await expect(page).toHaveURL(/#\/lobby$/);
    await expect(page.getByRole("heading", { name: "create game" })).toBeVisible();
  });

  test("logout clears the session", async ({ page }) => {
    await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
    await page.getByRole("button", { name: "log in" }).click();
    await expect(page).toHaveURL(/#\/lobby$/);

    await page.getByRole("button", { name: "log out", exact: true }).click();
    await expect(page).toHaveURL(/#\/login$/);

    // Visiting the lobby directly should now redirect us back to
    // login — the session is truly gone.
    await page.goto("/#/lobby");
    await expect(page).toHaveURL(/#\/login$/);
  });

  test("pasting a full invite URL navigates to the join view", async ({ page }) => {
    const url = `${page.url().split("#")[0]}#/games/11111111-2222-3333-4444-555555555555/join?t=xyz`;
    await page.getByLabel("invite code or link").fill(url);
    await page.getByRole("button", { name: "join" }).click();
    await expect(page).toHaveURL(/#\/games\/[0-9a-f-]+\/join\?t=xyz$/);
    await expect(page.getByRole("heading", { name: "join game" })).toBeVisible();
  });
});
