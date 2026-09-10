import { expect, test } from "@playwright/test";
import { adminLogin, createGame, joinViaAPI } from "./lobby-api";

// entry: the invite page a player actually lands on. #244 turned it
// from a bare name box into a preview of the table — GET
// /games/{id}/preview?t=<invite> returns the game meta with tokens and
// player IDs stripped, and the page renders the seat roster plus the
// refusal states (full / already started) before asking for anything.
// None of that had coverage; the join spec only exercised the form.
//
// These tests use joinViaAPI to stage the table's shape. Driving N
// browser contexts through the UI just to fill seats would be slow
// and would re-test what join.spec.ts already covers.

test.describe("invite entry page", () => {
  test("preview shows the table name, seat roster and open seats", async ({ page, request }) => {
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Preview ${Date.now()}`);
    await joinViaAPI(request, game.id, game.invite_token!, "Seated One");

    // A brand-new context: the invite is the only credential, and the
    // preview endpoint is deliberately unauthenticated.
    await page.goto(`/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`);

    await expect(page.getByRole("heading", { level: 1 })).toContainText(game.name);
    await expect(page.getByText(/1 of 4 seats/)).toBeVisible();

    const seats = page.getByRole("list", { name: "seats at this table" });
    await expect(seats.getByRole("listitem")).toHaveCount(4);
    await expect(seats.getByRole("listitem").first()).toContainText("Seated One");
    // The seat the visitor would claim is called out.
    await expect(seats).toContainText("This one's yours");
    // Nobody's deck is in yet.
    await expect(seats.getByText("deck pending")).toBeVisible();

    // The form is still the point of the page.
    await expect(page.getByPlaceholder("your name")).toBeVisible();
    await expect(page.getByRole("button", { name: "join" })).toBeDisabled();
  });

  test("a full table refuses the seat instead of failing on submit", async ({ page, request }) => {
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Full ${Date.now()}`);
    // Commander seats four.
    for (const name of ["One", "Two", "Three", "Four"]) {
      await joinViaAPI(request, game.id, game.invite_token!, name);
    }

    await page.goto(`/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`);

    await expect(page.getByRole("alert")).toContainText("This table is full.");
    // No name field to fill — the refusal replaces the form rather
    // than letting the player type a name and eat a 409.
    await expect(page.getByPlaceholder("your name")).toHaveCount(0);
    // …and a way to re-poll in case a seat frees up.
    await expect(page.getByRole("button", { name: "Check again" })).toBeVisible();
  });

  test("spectator link is read-only and lands straight on the table", async ({ page, request }) => {
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Spectate ${Date.now()}`);
    expect(game.spectator_invite).toBeTruthy();
    await joinViaAPI(request, game.id, game.invite_token!, "Seated One");

    const url = `/#/games/${game.id}/join?t=${encodeURIComponent(game.spectator_invite!)}&spectator=1`;
    await page.goto(url);

    // Same shell, different copy: the spectator flow never offers a
    // seat, and says so up front.
    await expect(page.getByText("You have a spectator link")).toBeVisible();
    await expect(page.getByText("read-only")).toBeVisible();
    await expect(page.getByRole("heading", { level: 1 })).toContainText(game.name);

    await page.getByPlaceholder("your name (chat label)").fill("Watcher");
    await page.getByRole("button", { name: "watch" }).click();

    // Spectators skip the lobby — no deck to import, no seat to
    // manage — and land on the game route directly.
    await expect(page).toHaveURL(new RegExp(`#/games/${game.id}$`));
    await expect(page.getByText("spectating")).toBeVisible();

    const session = await page.evaluate(() =>
      JSON.parse(localStorage.getItem("cmdctrl.session") ?? "null"),
    );
    // The session is a spectator session — that's what gates every
    // action affordance in Game.svelte. (player_id comes back as the
    // nil UUID rather than absent, so it isn't worth asserting on.)
    expect(session?.principal.role).toBe("spectator");
    expect(session?.gameID).toBe(game.id);
  });

  test("a player invite is rejected by the spectator flow", async ({ page, request }) => {
    // The two tokens are deliberately distinct: handing a spectator
    // the player link would let them claim a seat. Posting the player
    // invite to /spectate must fail.
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Wrong Token ${Date.now()}`);

    await page.goto(
      `/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}&spectator=1`,
    );
    await page.getByPlaceholder("your name (chat label)").fill("Sneaky");
    await page.getByRole("button", { name: "watch" }).click();

    await expect(page.getByRole("alert")).toBeVisible();
    await expect(page).toHaveURL(/\/join\?t=/);
  });
});
