import { expect, test } from "@playwright/test";

// Smoke tests validate the app boots and the router hands us the
// expected entry view. They make no assumptions about server state
// beyond /healthz returning 200, so they're safe to run against any
// freshly-started stack.

test("login page renders when unauthenticated", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "cmd_and_ctrl" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "admin login" })).toBeVisible();
  await expect(page.getByPlaceholder("admin token")).toBeVisible();
  await expect(page.getByRole("button", { name: "log in" })).toBeDisabled();
});

test("unauthenticated lobby redirects to login", async ({ page }) => {
  // App.svelte's auth gate should bounce us back to #/login.
  await page.goto("/#/lobby");
  await expect(page).toHaveURL(/#\/login$/);
  await expect(page.getByRole("heading", { name: "admin login" })).toBeVisible();
});

test("unauthenticated game route redirects to login", async ({ page }) => {
  await page.goto("/#/games/00000000-0000-0000-0000-000000000000");
  await expect(page).toHaveURL(/#\/login$/);
});

test("invite link route is public", async ({ page }) => {
  // Even without a session, the join view is reachable via invite —
  // this is how new players onboard. The invite token is invalid, so
  // the form will error on submit, but the page should render.
  await page.goto("/#/games/00000000-0000-0000-0000-000000000000/join?t=bogus");
  await expect(page.getByRole("heading", { name: "join game" })).toBeVisible();
  await expect(page.getByPlaceholder("your name")).toBeVisible();
});

test("health endpoint is reachable through the client proxy", async ({ request }) => {
  const res = await request.get("/healthz");
  expect(res.status()).toBe(200);
  expect(await res.text()).toBe("ok");
});
