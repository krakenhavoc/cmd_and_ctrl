import { expect, test } from "@playwright/test";
import { L } from "../../client/src/lib/labels";

// Smoke tests validate the app boots and the router hands us the
// expected entry view. They make no assumptions about server state
// beyond /healthz returning 200, so they're safe to run against any
// freshly-started stack.
//
// Selector note: #244 ("art-forward dark redesign") rewrote the entry
// pages, and ADR 0156 replaced the sign-in page with the front page. The admin panel heading
// (on #/admin only, since ADR 0112) is "Admin log in"; the
// placeholders and button labels survived. The
// assertions below deliberately lean on roles + form labels rather
// than the decorative copy around them.

test("landing page renders the front page", async ({ page }) => {
  // ADR 0156 §4: #/ is the front page, and #/login an alias of it.
  for (const path of ["/", "/#/login"]) {
    await page.goto(path);
    await expect(page.getByRole("heading", { level: 1 })).toContainText("over the wire");
    // The invite box is the first card under the hero. (With Discord
    // configured, Sign in with Discord leads the hero; these stacks set
    // no CMDCTRL_DISCORD_* values, so that button stays hidden.)
    await expect(page.getByRole("heading", { name: "Have an invite?" })).toBeVisible();
    await expect(page.getByLabel("invite code or link")).toBeVisible();
    await expect(page.getByRole("link", { name: L.tryPracticeGame })).toHaveAttribute(
      "href",
      "#/practice",
    );
    // No token form: it lives on #/admin only (ADR 0112 §2 item 8).
    await expect(page.getByRole("heading", { name: "Admin log in" })).toHaveCount(0);
    await expect(page.getByPlaceholder("admin token")).toHaveCount(0);
  }
});

test("#/home is retired and goes to the front page", async ({ page }) => {
  await page.goto("/#/home");
  await expect(page).toHaveURL(/#\/$/);
  await expect(page.getByRole("heading", { name: "Have an invite?" })).toBeVisible();
});

test("the gate page is public, and ?retry=1 leads with Discord not answering", async ({
  page,
}) => {
  // ADR 0156 §5: where the Discord callback sends a refused sign-in.
  await page.goto("/#/request-access");
  await expect(page).toHaveURL(/#\/request-access$/);
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    "Ask for an account on our Discord",
  );
  await expect(page.getByRole("link", { name: L.tryPracticeGame })).toBeVisible();
  await page.goto("/#/request-access?retry=1");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    "Discord didn't answer — try again",
  );
});

test("a signed-out visitor opens a guest practice game from the front page", async ({ page }) => {
  // ADR 0156 §2: no account, no sign-in: POST /games/practice/guest.
  await page.goto("/");
  await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
  await page.goto("/");
  await page.getByRole("link", { name: L.tryPracticeGame }).click();
  await expect(page).toHaveURL(/#\/games\/[0-9a-f-]+$/, { timeout: 20_000 });
  const role = await page.evaluate(
    () => JSON.parse(localStorage.getItem("cmdctrl.session") ?? "null")?.principal?.role,
  );
  expect(role).toBe("player");
});

test("#/admin leads with the token form and hides the invite paste", async ({ page }) => {
  // The admin route is the same component with `admin` set: no
  // invite card, an explanatory note, and a link back to the
  // player-first landing.
  await page.goto("/#/admin");
  await expect(page.getByRole("heading", { name: "Admin log in" })).toBeVisible();
  await expect(page.getByPlaceholder("admin token")).toBeVisible();
  await expect(page.getByLabel("invite code or link")).toHaveCount(0);
  await expect(page.getByRole("link", { name: "Back" })).toBeVisible();
});

test("unauthenticated lobby redirects to login", async ({ page }) => {
  // App.svelte's auth gate should bounce us back to #/login.
  await page.goto("/#/lobby");
  await expect(page).toHaveURL(/#\/login$/);
  await expect(page.getByRole("heading", { name: "Have an invite?" })).toBeVisible();
});

test("unauthenticated game route redirects to login", async ({ page }) => {
  await page.goto("/#/games/00000000-0000-0000-0000-000000000000");
  await expect(page).toHaveURL(/#\/login$/);
});

test("the decks page is public, and #/deck-check opens it", async ({ page }) => {
  // ADR 0112 §3: one decks page, public, with "Decks" as the header's
  // only deck link. #/deck-check is a permanent alias because the
  // Discord bot's /c2-deck-check replies link there.
  for (const hash of ["/#/decks", "/#/deck-check"]) {
    await page.goto(hash);
    await expect(page).not.toHaveURL(/#\/login$/);
    await expect(page.getByRole("heading", { name: "Decks", level: 1 })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Check a deck" })).toBeVisible();
    await expect(page.getByLabel("deck link")).toBeVisible();
    await expect(page.getByRole("button", { name: "Check this deck" })).toBeVisible();
    // Signed out, the library is one line asking for Discord.
    await expect(page.getByText("Saved decks are kept for your Discord account.")).toBeVisible();
  }
  const nav = page.getByRole("navigation", { name: "Site" });
  await expect(nav.getByRole("link", { name: "Decks" })).toHaveAttribute("href", "#/decks");
  await expect(nav.getByRole("link", { name: "Deck check" })).toHaveCount(0);
  await expect(nav.getByRole("link", { name: "My decks" })).toHaveCount(0);
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
