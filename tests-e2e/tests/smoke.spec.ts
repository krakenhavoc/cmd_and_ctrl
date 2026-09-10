import { expect, test } from "@playwright/test";

// Smoke tests validate the app boots and the router hands us the
// expected entry view. They make no assumptions about server state
// beyond /healthz returning 200, so they're safe to run against any
// freshly-started stack.
//
// Selector note: #244 ("art-forward dark redesign") rewrote the entry
// pages. The wordmark is now "CMD & CTRL" and the admin panel heading
// is "Admin log in"; the placeholders and button labels survived. The
// assertions below deliberately lean on roles + form labels rather
// than the decorative copy around them.

test("landing page renders the player-first entry", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "CMD & CTRL", level: 1 })).toBeVisible();
  // Players lead: the invite paste box is the first card on #/login.
  await expect(page.getByRole("heading", { name: "Have an invite?" })).toBeVisible();
  await expect(page.getByLabel("invite link")).toBeVisible();
  // …with the admin token form under it.
  await expect(page.getByRole("heading", { name: "Admin log in" })).toBeVisible();
  await expect(page.getByPlaceholder("admin token")).toBeVisible();
  await expect(page.getByRole("button", { name: "log in" })).toBeDisabled();
});

test("#/admin leads with the token form and hides the invite paste", async ({ page }) => {
  // The admin route is the same component with `admin` set: no
  // invite card, an explanatory note, and a link back to the
  // player-first landing.
  await page.goto("/#/admin");
  await expect(page.getByRole("heading", { name: "Admin log in" })).toBeVisible();
  await expect(page.getByPlaceholder("admin token")).toBeVisible();
  await expect(page.getByLabel("invite link")).toHaveCount(0);
  await expect(page.getByRole("link", { name: "Back" })).toBeVisible();
});

test("unauthenticated lobby redirects to login", async ({ page }) => {
  // App.svelte's auth gate should bounce us back to #/login.
  await page.goto("/#/lobby");
  await expect(page).toHaveURL(/#\/login$/);
  await expect(page.getByRole("heading", { name: "Admin log in" })).toBeVisible();
});

test("unauthenticated game route redirects to login", async ({ page }) => {
  await page.goto("/#/games/00000000-0000-0000-0000-000000000000");
  await expect(page).toHaveURL(/#\/login$/);
});

test("invite link route is public", async ({ page }) => {
  // Even without a session, the join view is reachable via invite —
  // this is how new players onboard. The invite token is invalid, so
  // the preview fails and the heading falls back to the generic
  // "join game" copy, but the page should render its form.
  await page.goto("/#/games/00000000-0000-0000-0000-000000000000/join?t=bogus");
  await expect(page.getByRole("heading", { name: "join game" })).toBeVisible();
  await expect(page.getByPlaceholder("your name")).toBeVisible();
});

test("health endpoint is reachable through the client proxy", async ({ request }) => {
  const res = await request.get("/healthz");
  expect(res.status()).toBe(200);
  expect(await res.text()).toBe("ok");
});
