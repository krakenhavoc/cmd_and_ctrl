import { expect, test, type Page } from "@playwright/test";
import { makeCommanderDeck } from "./deck-fixture";
import {
  adminLogin,
  createGame,
  joinViaAPI,
  uploadDeckAs,
  type GameMeta,
} from "./lobby-api";

// ADR 0212: a table in the lobby as a character select. The seats are
// four tiles, a seat with its deck in says Ready, the footer's readiness
// line counts the seated players, and its one primary follows the
// table's state: Choose your deck, Waiting for …, Start game. The lobby
// has no socket, so a deck seated elsewhere arrives with the next read
// (a poll every 5s, or on focus, which these tests use).

async function joinInBrowser(
  page: Page,
  game: GameMeta,
  name: string,
): Promise<void> {
  await page.goto(
    `/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`,
  );
  await page.getByPlaceholder("your name").fill(name);
  await page.getByRole("button", { name: "join" }).click();
  await expect(page).toHaveURL(/#\/lobby$/);
}

function cardOf(page: Page, name: string) {
  return page.locator("ul.games > li").filter({ hasText: name });
}

async function reread(page: Page): Promise<void> {
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
}

test.describe("lobby seats (ADR 0212)", () => {
  test.beforeEach(async ({ page }) => {
    // Tips off, as in lobby.spec: a tip's card can sit over a control.
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
  });

  test("a guest sits down with a pre-built deck and the readiness line moves", async ({
    page,
    request,
  }) => {
    const admin = await adminLogin(request);
    const game = await createGame(request, admin, `Seats Guest ${Date.now()}`);
    // The first person in hosts, and is ready.
    const host = await joinViaAPI(request, game.id, game.invite_token!, "Hal");
    await uploadDeckAs(
      request,
      admin,
      game.id,
      host.player_id,
      makeCommanderDeck(),
    );

    await joinInBrowser(page, game, "Gil");
    const card = cardOf(page, game.name);
    const seats = card.getByRole("list", { name: "seats" });
    await expect(
      seats.getByRole("listitem", { name: "seat 1: Hal" }),
    ).toContainText("Ready");
    const mine = seats.getByRole("listitem", { name: "seat 2: Gil" });
    await expect(mine).toContainText("No deck yet");
    await expect(
      card.getByRole("group", { name: "ready 1 of 2" }),
    ).toContainText("2 seats open · waiting on your deck");
    const primary = card.locator(".tfoot .btn.primary");
    await expect(primary).toHaveText("Choose your deck");

    // Seated without a deck, the picker is open under the seats. A guest
    // has no library, so the pre-built decks are listed.
    const picker = card.getByRole("region", { name: "choose your deck" });
    await expect(picker).toBeVisible();
    await picker.locator('input[name="prebuilt-deck"]').first().check();
    await picker
      .getByRole("button", { name: "Sit down with this deck" })
      .click();

    await expect(
      card.getByRole("group", { name: "ready 2 of 2" }),
    ).toContainText("2 seats open · every deck is in", { timeout: 15_000 });
    await expect(mine).toContainText("Ready");
    // Gil does not host: Hal starts.
    await expect(primary).toHaveText("Waiting for Hal to start");
    await expect(primary).toBeDisabled();
    // Their tile still opens the picker, to change the deck.
    await mine.getByRole("button", { name: "change your deck" }).click();
    await expect(picker).toBeVisible();
  });

  test("the host waits for a second player, then starts from the footer", async ({
    page,
    request,
  }) => {
    const admin = await adminLogin(request);
    const game = await createGame(request, admin, `Seats Start ${Date.now()}`);
    await joinInBrowser(page, game, "Hal");
    const card = cardOf(page, game.name);
    const primary = card.locator(".tfoot .btn.primary");
    await expect(primary).toHaveText("Choose your deck");

    const me = await page.evaluate(
      () =>
        JSON.parse(localStorage.getItem("cmdctrl.session") ?? "{}")
          .playerID as string,
    );
    await uploadDeckAs(request, admin, game.id, me, makeCommanderDeck());
    await reread(page);
    await expect(primary).toHaveText("Waiting for players");
    await expect(primary).toBeDisabled();

    const other = await joinViaAPI(
      request,
      game.id,
      game.invite_token!,
      "Olive",
    );
    await reread(page);
    await expect(primary).toHaveText("Waiting for 1 player");
    await expect(
      card.getByRole("group", { name: "ready 1 of 2" }),
    ).toContainText("waiting on Olive's deck");

    await uploadDeckAs(
      request,
      admin,
      game.id,
      other.player_id,
      makeCommanderDeck(),
    );
    await reread(page);
    await expect(primary).toHaveText("Start game · 2 players");
    await primary.click();
    // In play, the card keeps its primary in its head.
    await expect(card.getByText(/^In play/)).toBeVisible({ timeout: 15_000 });
    await expect(
      card.getByRole("button", { name: "Resume", exact: true }),
    ).toBeVisible();
  });
});
