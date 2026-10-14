import {
  expect,
  test,
  type APIRequestContext,
  type Page,
} from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import {
  adminLogin,
  createGame,
  getGameAs,
  uploadDeckAs,
  startGameAs,
} from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { joinAsPlayer, closeAll, type JoinedPlayer } from "./players";

// ADR 0177 (#3337): a three-seat table where one page closes. After the
// table's away time (set to 1 minute) the host is offered a way on: hand
// the seat to a bot, concede for them, or keep waiting. Each does what it
// says, and the closed page, reopened, takes its seat back from the bot.
//
// Prereqs match mulligan.spec.ts: server + Scryfall bulk dump. Slow by
// nature: the away time is a real minute, and the server's away clock
// ticks every 5 seconds.

const dockOf = (page: Page) =>
  page.getByRole("region", { name: "actions", exact: true });
// Cy's seat's presence chip, as another seat sees it.
const cyChip = (page: Page) =>
  page.getByLabel(L.seatBoard("Cy")).locator(".tag.presence.offline");
const offerOf = (page: Page) =>
  dockOf(page).getByRole("dialog", { name: L.awayOffer });

async function setAwayAfter(
  req: APIRequestContext,
  token: string,
  gameID: string,
  minutes: number,
): Promise<void> {
  const res = await req.patch(`/games/${gameID}/settings`, {
    data: { away_after: minutes },
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok())
    throw new Error(
      `set away_after failed: ${res.status()} ${await res.text()}`,
    );
}

// keepUntilOffered keeps the opening hand of whichever present seat the
// mulligan window is waiting on, until the window reaches the closed
// seat and the host is offered a way on without it.
async function keepUntilOffered(host: Page, present: Page[]): Promise<void> {
  await expect
    .poll(
      async () => {
        if (await offerOf(host).isVisible()) return true;
        for (const p of present) {
          const keep = dockOf(p)
            .getByRole("dialog", { name: /keep or mulligan/i })
            .getByRole("button", { name: "Keep hand" });
          if ((await keep.count()) > 0 && (await keep.isEnabled()))
            await keep.click();
        }
        return false;
      },
      { timeout: 150_000, intervals: [2_000] },
    )
    .toBe(true);
}

async function seatTable(
  browser: Parameters<typeof joinAsPlayer>[0],
  request: APIRequestContext,
  name: string,
): Promise<{
  admin: string;
  gameID: string;
  ana: JoinedPlayer;
  bo: JoinedPlayer;
  cy: JoinedPlayer;
}> {
  const admin = await adminLogin(request);
  const game = await createGame(request, admin, `${name} ${Date.now()}`);
  const ana = await joinAsPlayer(browser, game.id, game.invite_token!, "Ana");
  const bo = await joinAsPlayer(browser, game.id, game.invite_token!, "Bo");
  const cy = await joinAsPlayer(browser, game.id, game.invite_token!, "Cy");
  await setAwayAfter(request, admin, game.id, 1);
  const deck = makeCommanderDeck();
  for (const p of [ana, bo, cy])
    await uploadDeckAs(request, admin, game.id, p.playerID, deck);
  await startGameAs(request, admin, game.id);
  return { admin, gameID: game.id, ana, bo, cy };
}

test.describe("moving the table on without an away player", () => {
  let players: (JoinedPlayer | null)[] = [];

  test.afterEach(async () => {
    await closeAll(...players);
    players = [];
  });

  test("keep waiting, then a bot, then the player takes the seat back", async ({
    browser,
    request,
  }) => {
    test.setTimeout(300_000);
    const { admin, gameID, ana, bo, cy } = await seatTable(
      browser,
      request,
      "Away",
    );
    players = [ana, bo, cy];

    // Cy's page closes. The session survives, as a closed laptop's does.
    const cyState = await cy.context.storageState();
    await cy.context.close();
    players = [ana, bo];

    // Every seat sees who is gone.
    await expect(cyChip(ana.page)).toBeVisible({
      timeout: 20_000,
    });
    await expect(cyChip(bo.page)).toBeVisible({
      timeout: 20_000,
    });

    await keepUntilOffered(ana.page, [ana.page, bo.page]);
    const offer = offerOf(ana.page);
    await expect(offer).toContainText("Cy has been offline for");
    // The host is at the table, so only the host is asked.
    await expect(offerOf(bo.page)).toHaveCount(0);

    // Keep waiting hides the offer for another away time.
    await offer.getByRole("button", { name: L.awayWait }).click();
    await expect(offerOf(ana.page)).toHaveCount(0);
    await expect(offerOf(ana.page)).toBeVisible({ timeout: 90_000 });

    // Hand Cy's seat to a bot: the seat says so, and the table plays on.
    await offerOf(ana.page)
      .getByRole("button", { name: L.awayStandIn("Cy") })
      .click();
    await expect(offerOf(ana.page)).toHaveCount(0);
    await expect(ana.page.locator(".tag.stand-in")).toHaveText("(bot)");
    await expect
      .poll(async () =>
        (await getGameAs(request, admin, gameID)).players.find(
          (p) => p.name === "Cy",
        ),
      )
      .toMatchObject({ stand_in: "heuristic" });

    // Cy comes back. The bot keeps playing until Cy presses the button.
    const back = await browser.newContext({ storageState: cyState });
    const cyPage = await back.newPage();
    players = [ana, bo, { ...cy, context: back, page: cyPage }];
    await cyPage.goto(`/#/games/${gameID}`);
    const banner = cyPage
      .getByRole("status")
      .filter({ hasText: "A bot has been playing your seat" });
    await expect(banner).toBeVisible({ timeout: 20_000 });
    await banner.getByRole("button", { name: L.takeBackSeat }).click();
    await expect(banner).toHaveCount(0);
    await expect(ana.page.locator(".tag.stand-in")).toHaveCount(0);
    await expect
      .poll(async () =>
        (await getGameAs(request, admin, gameID)).players.find(
          (p) => p.name === "Cy",
        ),
      )
      .not.toHaveProperty("stand_in");
  });

  test("conceding for an away player", async ({ browser, request }) => {
    test.setTimeout(240_000);
    const { ana, bo, cy } = await seatTable(browser, request, "Away concede");
    players = [ana, bo, cy];
    await cy.context.close();
    players = [ana, bo];

    await keepUntilOffered(ana.page, [ana.page, bo.page]);
    await offerOf(ana.page)
      .getByRole("button", { name: L.awayConcede("Cy") })
      .click();
    await expect(offerOf(ana.page)).toHaveCount(0);
    // Cy has left the game; two players play on.
    await expect(ana.page.locator(".identity.eliminated")).toHaveCount(1, {
      timeout: 20_000,
    });
  });
});
