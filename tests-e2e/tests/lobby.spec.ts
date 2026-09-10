import { expect, test } from "@playwright/test";
import { ADMIN_TOKEN } from "./env";
import { adminLogin, createGame, joinViaAPI } from "./lobby-api";

test.describe("lobby — admin flow", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
    await page.goto("/#/login");
    await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
    await page.getByRole("button", { name: "log in" }).click();
    await expect(page).toHaveURL(/#\/lobby$/);
  });

  test("creating a game adds it to the list with an invite button", async ({ page }) => {
    const name = `E2E Test ${Date.now()}`;
    await page.getByPlaceholder("game name").fill(name);
    await page.getByRole("button", { name: "create" }).click();

    // The new game appears in the list…
    const gameRow = page.locator("ul.games > li").filter({ hasText: name });
    await expect(gameRow).toBeVisible();
    // …and the invite-token cache flips on the "copy invite" button
    // for freshly created games (the server strips invite tokens on
    // subsequent list calls, so the button only shows this session).
    await expect(gameRow.getByRole("button", { name: "copy invite" })).toBeVisible();
    // Start is gated until every seat uploads a deck — with zero
    // seats we expect start to be absent (the UI only renders it
    // when there are ≥2 players).
    await expect(gameRow.getByRole("button", { name: "start" })).toHaveCount(0);
    // Opening the game navigates to the game route.
    await expect(gameRow.getByRole("button", { name: "open" })).toBeVisible();
  });

  test("empty game name keeps the create button disabled", async ({ page }) => {
    await expect(page.getByRole("button", { name: "create" })).toBeDisabled();
    await page.getByPlaceholder("game name").fill("   ");
    // Trim-only input should still leave the button disabled.
    await expect(page.getByRole("button", { name: "create" })).toBeDisabled();
    await page.getByPlaceholder("game name").fill("Valid Name");
    await expect(page.getByRole("button", { name: "create" })).toBeEnabled();
  });

  test("refresh button re-fetches the games list", async ({ page }) => {
    const name = `Refresh Test ${Date.now()}`;
    await page.getByPlaceholder("game name").fill(name);
    await page.getByRole("button", { name: "create" }).click();
    await expect(page.locator("ul.games > li").filter({ hasText: name })).toBeVisible();

    // Fire a network listener so we can assert refresh actually
    // hits /games — not just a client-side repaint.
    const refreshPromise = page.waitForResponse(
      (res) => res.url().endsWith("/games") && res.request().method() === "GET",
    );
    await page.getByRole("button", { name: "refresh" }).click();
    await refreshPromise;
    await expect(page.locator("ul.games > li").filter({ hasText: name })).toBeVisible();
  });

  test("a fresh table offers both invite links", async ({ page }) => {
    // The player invite and the spectator invite are different
    // tokens — handing a spectator the player link would let them
    // claim a seat — so the lobby surfaces them as two buttons.
    // Both are session-cached client-side: the list endpoint strips
    // them, so they only appear on the table you just created.
    const name = `Two Invites ${Date.now()}`;
    await page.getByPlaceholder("game name").fill(name);
    await page.getByRole("button", { name: "create" }).click();

    const row = page.locator("ul.games > li").filter({ hasText: name });
    await expect(row.getByRole("button", { name: "copy invite" })).toBeVisible();
    await expect(row.getByRole("button", { name: "Spectator link" })).toBeVisible();
  });

  test("the table card renders four seats and fills them as players join", async ({
    page,
    request,
  }) => {
    // #244 turned each lobby row into a table card: a four-seat
    // roster padded with open slots, each seat showing its deck
    // status. The lobby has no WebSocket, so it polls — a seat
    // claimed out-of-band must appear without a reload.
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Seats ${Date.now()}`);
    await page.getByRole("button", { name: "refresh" }).click();

    const row = page.locator("ul.games > li").filter({ hasText: game.name });
    await expect(row).toBeVisible();
    const seats = row.getByRole("list", { name: "seats" });
    await expect(seats.getByRole("listitem")).toHaveCount(4);
    await expect(seats.getByText("Open seat")).toHaveCount(4);
    await expect(row).toContainText("0 of 4 seats");

    await joinViaAPI(request, game.id, game.invite_token!, "Late Arrival");

    // The poll (5s) picks the new seat up on its own.
    await expect(seats.getByText("seat 1: Late Arrival")).toBeVisible({ timeout: 15_000 });
    await expect(seats.getByText("Open seat")).toHaveCount(3);
    await expect(seats.getByText("deck pending")).toBeVisible();
  });

  test("opening a game navigates to the game route", async ({ page }) => {
    const name = `Open Game ${Date.now()}`;
    await page.getByPlaceholder("game name").fill(name);
    await page.getByRole("button", { name: "create" }).click();
    const row = page.locator("ul.games > li").filter({ hasText: name });
    await expect(row).toBeVisible();
    await row.getByRole("button", { name: "open" }).click();
    await expect(page).toHaveURL(/#\/games\/[0-9a-f-]+$/);
  });
});
