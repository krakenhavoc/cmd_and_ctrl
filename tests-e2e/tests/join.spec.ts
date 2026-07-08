import { expect, test } from "@playwright/test";
import { adminLogin, createGame } from "./lobby-api";

test.describe("invite → join", () => {
  test("invalid invite token shows an error on submit", async ({ page, request }) => {
    // Create a real game so the URL's gameID is valid; only the
    // invite token is bogus. This catches the server's invite-token
    // validation path specifically (rather than a "game not found"
    // path we'd get with a random UUID).
    const token = await adminLogin(request);
    const game = await createGame(request, token, `Invalid Invite ${Date.now()}`);

    await page.goto(`/#/games/${game.id}/join?t=not-a-real-invite-token`);
    await expect(page.getByRole("heading", { name: "join game" })).toBeVisible();
    await page.getByPlaceholder("your name").fill("Rejected Player");
    await page.getByRole("button", { name: "join" }).click();
    await expect(page.locator("p.error")).toBeVisible();
    // We stay on the join page — no redirect to the game view.
    await expect(page).toHaveURL(/\/join\?t=/);
  });

  test("missing invite token renders the 'token missing' warning", async ({ page }) => {
    await page.goto(`/#/games/00000000-0000-0000-0000-000000000000/join`);
    await expect(page.getByText(/invite token missing from URL/i)).toBeVisible();
    // The form should not render when the token slot is empty.
    await expect(page.getByPlaceholder("your name")).toHaveCount(0);
  });

  test("valid invite link seats the player and lands them in the lobby", async ({ page, request }) => {
    const token = await adminLogin(request);
    const game = await createGame(request, token, `Join Flow ${Date.now()}`);
    expect(game.invite_token).toBeTruthy();

    // A brand-new browser context = no admin session. The invite
    // flow must be usable without one.
    await page.goto(`/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`);
    await expect(page.getByRole("heading", { name: "join game" })).toBeVisible();
    await page.getByPlaceholder("your name").fill("E2E Player");
    await page.getByRole("button", { name: "join" }).click();

    // Players land in the lobby first (deck import, seat status)
    // rather than the game route — s085 (#43). Spectators skip this.
    await expect(page).toHaveURL(/#\/lobby$/);
    await expect(page.getByText(/seat 0: E2E Player/)).toBeVisible();

    // We can't easily assert on the full Pixi canvas from here, but
    // the session store should now carry a player role tied to this
    // game. Reading it back through the DOM is flaky with async
    // mount; pull it from localStorage instead.
    const session = await page.evaluate(() =>
      JSON.parse(localStorage.getItem("cmdctrl.session") ?? "null"),
    );
    expect(session?.principal.role).toBe("player");
    expect(session?.gameID).toBe(game.id);
    expect(session?.playerID).toBeTruthy();
  });

  test("empty player name keeps the join button disabled", async ({ page, request }) => {
    const token = await adminLogin(request);
    const game = await createGame(request, token, `Empty Name ${Date.now()}`);

    await page.goto(`/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`);
    const joinBtn = page.getByRole("button", { name: "join" });
    await expect(joinBtn).toBeDisabled();
    await page.getByPlaceholder("your name").fill("   ");
    // Whitespace-only still counts as empty by the UI's trim check.
    await expect(joinBtn).toBeDisabled();
    await page.getByPlaceholder("your name").fill("OK");
    await expect(joinBtn).toBeEnabled();
  });
});
