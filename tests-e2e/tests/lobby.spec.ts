import { expect, test, type Page } from "@playwright/test";
import { ADMIN_TOKEN } from "./env";
import { adminLogin, createGame, joinViaAPI } from "./lobby-api";

// Tables (ADR 0158 §1): + New table in the page head opens the create
// form; each table in Now is a card with one primary button and a
// toolbar (#3499): the invite and spectator links as copy / ↻ split
// buttons, and the rest. A right-click on the card opens the same
// actions, and the admin's, as a menu ("actions for <name>").

async function openNewTable(page: Page): Promise<void> {
  await page.getByRole("button", { name: "new table" }).click();
  await expect(page.getByPlaceholder("game name")).toBeVisible();
}

function cardOf(page: Page, name: string) {
  return page.locator("ul.games > li").filter({ hasText: name });
}

test.describe("lobby — admin flow", () => {
  test.beforeEach(async ({ page }) => {
    // Tips off: on an empty lobby the admin token is offered the
    // site.help tip, and its card sits over the create button. The
    // serial suite only passed because earlier specs had left tables in
    // the lobby; as the first spec of a shard (#2661) every test here
    // timed out on that card. Tips are hints.spec's subject, not this
    // one's. Written only when absent, so the app's own saves stand.
    await page.context().addInitScript(() => {
      try {
        if (localStorage.getItem("cmdctrl.settings.v1") === null) {
          localStorage.setItem(
            "cmdctrl.settings.v1",
            JSON.stringify({ help: { tipsOff: true } }),
          );
        }
      } catch {
        // storage unavailable: fall back to the defaults
      }
    });
    await page.goto("/");
    await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
    // The token form lives on #/admin only (ADR 0112 §2 item 8).
    await page.goto("/#/admin");
    await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
    await page.getByRole("button", { name: "log in" }).click();
    await expect(page).toHaveURL(/#\/lobby$/);
  });

  test("creating a game adds it to Now with Copy invite", async ({ page }) => {
    const name = `E2E Test ${Date.now()}`;
    await openNewTable(page);
    await page.getByPlaceholder("game name").fill(name);
    await page.getByRole("button", { name: "create", exact: true }).click();

    // The new game appears in Now, and the form closes…
    const card = cardOf(page, name);
    await expect(card).toBeVisible();
    await expect(page.getByPlaceholder("game name")).toHaveCount(0);
    // …with its invite link one click away.
    await expect(
      card.getByRole("button", { name: "Copy invite" }),
    ).toBeVisible();
    // Start is offered only from two seats; with none, the primary
    // button opens the table.
    await expect(
      card.getByRole("button", { name: "Start", exact: true }),
    ).toHaveCount(0);
    await expect(
      card.getByRole("button", { name: "Open", exact: true }),
    ).toBeVisible();
  });

  test("a blank game name creates a table with a generated name (#2630)", async ({
    page,
  }) => {
    await openNewTable(page);
    await expect(
      page.getByRole("button", { name: "create", exact: true }),
    ).toBeEnabled();
    await page.getByPlaceholder("game name").fill("   ");
    // Trim-only input is a blank name too: the server picks one.
    await expect(
      page.getByRole("button", { name: "create", exact: true }),
    ).toBeEnabled();
    const rows = page.locator("ul.games > li");
    const before = await rows.count();
    await page.getByRole("button", { name: "create", exact: true }).click();
    await expect(rows).toHaveCount(before + 1);
  });

  test("the dice fills the name box with a suggestion (#2630)", async ({
    page,
  }) => {
    await openNewTable(page);
    const box = page.getByPlaceholder("game name");
    await expect(box).toHaveValue("");
    await page.getByRole("button", { name: "suggest a name" }).click();
    await expect(box).not.toHaveValue("");
  });

  test("the page re-reads the list when the window regains focus", async ({
    page,
    request,
  }) => {
    // There is no refresh button any more (ADR 0158 §1 item 6): the page
    // polls while a table in Now can change, and re-reads on focus.
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Focus ${Date.now()}`);
    const refreshed = page.waitForResponse(
      (res) => res.url().endsWith("/games") && res.request().method() === "GET",
    );
    await page.evaluate(() => window.dispatchEvent(new Event("focus")));
    await refreshed;
    await expect(cardOf(page, game.name)).toBeVisible();
  });

  test("a fresh table offers both invite links", async ({ page }) => {
    // The player invite and the spectator invite are different tokens —
    // handing a spectator the player link would let them claim a seat —
    // so the card's toolbar has a split button for each (#3499).
    const name = `Two Invites ${Date.now()}`;
    await openNewTable(page);
    await page.getByPlaceholder("game name").fill(name);
    await page.getByRole("button", { name: "create", exact: true }).click();

    const card = cardOf(page, name);
    await expect(
      card.getByRole("button", { name: "Copy invite" }),
    ).toBeVisible();
    await expect(
      card.getByRole("button", { name: "Copy spectator link" }),
    ).toBeVisible();
    await expect(
      card.getByRole("button", { name: "new invite link", exact: true }),
    ).toBeVisible();
    await expect(
      card.getByRole("button", { name: "new spectator link", exact: true }),
    ).toBeVisible();
    await expect(card.getByRole("button", { name: `more for ${name}` })).toHaveCount(0);

    // The right-click menu carries the same, and the admin's Archive.
    await card.locator(".seats").click({ button: "right" });
    const menu = page.getByRole("menu", { name: `actions for ${name}` });
    await expect(
      menu.getByRole("menuitem", { name: "Copy spectator link" }),
    ).toBeVisible();
    await expect(
      menu.getByRole("menuitem", { name: "New invite link…" }),
    ).toBeVisible();
    await expect(menu.getByRole("menuitem", { name: "Archive…" })).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(menu).toHaveCount(0);
  });

  test("the table card renders four seats and fills them as players join", async ({
    page,
    request,
  }) => {
    // A table in the lobby is a four-seat roster padded with open seats,
    // each seat showing its deck status. The lobby has no WebSocket, so
    // it polls: a seat claimed out-of-band must appear without a reload.
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Seats ${Date.now()}`);
    await page.evaluate(() => window.dispatchEvent(new Event("focus")));

    const card = cardOf(page, game.name);
    await expect(card).toBeVisible();
    const seats = card.getByRole("list", { name: "seats" });
    await expect(seats.getByRole("listitem")).toHaveCount(4);
    await expect(seats.getByText("Open seat")).toHaveCount(4);
    await expect(card.getByText("Lobby", { exact: true })).toBeVisible();

    await joinViaAPI(request, game.id, game.invite_token!, "Late Arrival");

    // The poll (5s) picks the new seat up on its own.
    await expect(seats.getByText("Late Arrival")).toBeVisible({
      timeout: 15_000,
    });
    await expect(seats.getByText("Open seat")).toHaveCount(3);
    await expect(seats.getByText("deck pending")).toBeVisible();
  });

  test("opening a game navigates to the game route", async ({ page }) => {
    const name = `Open Game ${Date.now()}`;
    await openNewTable(page);
    await page.getByPlaceholder("game name").fill(name);
    await page.getByRole("button", { name: "create", exact: true }).click();
    const card = cardOf(page, name);
    await expect(card).toBeVisible();
    await card.getByRole("button", { name: "Open", exact: true }).click();
    await expect(page).toHaveURL(/#\/games\/[0-9a-f-]+$/);
  });

  test("#/my-games lands on Tables' History", async ({ page }) => {
    // ADR 0158 §1 item 5: My games is folded into Tables.
    await page.goto("/#/my-games");
    await expect(page).toHaveURL(/#\/lobby\?show=history$/);
    await expect(page.getByRole("heading", { name: /^History/ })).toBeVisible();
    const nav = page.getByRole("navigation", { name: "Site" });
    await expect(nav.getByRole("link", { name: "My games" })).toHaveCount(0);
  });
});
