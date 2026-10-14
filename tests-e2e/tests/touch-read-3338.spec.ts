import { expect, test, type Page } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  findCardInZone,
  keepAllHands,
  openAdminClient,
  playerByID,
  type AdminClient,
} from "./s19-helpers";

// #3338: on a tablet, a player reads a card without acting on it.
//
// The render test (client/src/lib/touchInspect.render.test.ts) drives
// the gestures with stand-in events. This drives a real touchscreen:
// both seats play on a touch tablet context, and the touches go through
// Chrome's Input.dispatchTouchEvent, so the browser itself turns them
// into pointer and click events the way an iPad's finger does.
//
// In the active player's precombat main, with a Forest in hand (a tap
// on it would play it):
//   1. a long-press shows the zoom, and the Forest stays in hand;
//   2. the next tap only closes the zoom, and the Forest stays in hand;
//   3. a tap after that plays it, so the gestures did not break play.

const LAND = "Forest";
const TABLET = { hasTouch: true, isMobile: false, viewport: { width: 1024, height: 768 } };

function makeDeck(): string {
  return ["Commander:", `1 ${COMMANDER_NAME}`, "", "Mainboard:", `99 ${LAND}`].join("\n") + "\n";
}

async function touchAt(page: Page, x: number, y: number, holdMs: number): Promise<void> {
  const cdp = await page.context().newCDPSession(page);
  await cdp.send("Input.dispatchTouchEvent", { type: "touchStart", touchPoints: [{ x, y }] });
  await page.waitForTimeout(holdMs);
  await cdp.send("Input.dispatchTouchEvent", { type: "touchEnd", touchPoints: [] });
  await cdp.detach();
}

test.describe("#3338 reading a card on a tablet", () => {
  test("long-press shows the zoom without playing the card; a tap still plays it", async ({
    browser,
    request,
  }) => {
    test.slow();
    let first: JoinedPlayer | null = null;
    let second: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;
    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Touch read 3338 ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing on fresh game");
      first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One", undefined, TABLET);
      second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two", undefined, TABLET);
      await uploadDeckAs(request, adminToken, game.id, first.playerID, makeDeck());
      await uploadDeckAs(request, adminToken, game.id, second.playerID, makeDeck());
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(adminToken, game.id, first.playerID, second.playerID);
      await keepAllHands(admin);
      await admin.waitFor(
        (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
        "the active player's precombat main",
        15_000,
      );
      const snap = admin.snapshot();
      const activeID = snap.seats[snap.turn?.active_seat ?? 0]?.id;
      const me = activeID === first.playerID ? first : second;
      const forest = findCardInZone(playerByID(snap, me.playerID).hand, LAND);
      if (!forest) throw new Error("no Forest in the opening hand");

      const card = me.page
        .locator('[aria-label="your hand"]')
        .locator(`.card[data-instance-id="${forest.instance_id}"]`);
      await expect(card).toBeVisible();
      const box = await card.boundingBox();
      if (!box) throw new Error("the Forest has no box");
      const x = box.x + box.width / 2;
      const y = box.y + box.height / 3;
      const zoom = me.page.locator(".overlay:has(.scan)");
      // Every card in the deck is a Forest, and the hand fan re-lays
      // itself out, so count my lands rather than follow one instance.
      const landsPlayed = (): number =>
        admin!
          .snapshot()
          .battlefield.cards.filter((c) => c.name === LAND && c.controller === me.playerID).length;
      expect(landsPlayed()).toBe(0);

      // 1. Long-press: the zoom, and no land drop.
      await touchAt(me.page, x, y, 900);
      await expect(zoom).toBeVisible();
      await expect(zoom).toContainText(LAND);
      await me.page.waitForTimeout(500);
      expect(landsPlayed(), "a long-press played a Forest").toBe(0);

      // 2. The next tap closes the zoom, and acts on nothing.
      await touchAt(me.page, x, y, 60);
      await expect(zoom).toHaveCount(0);
      await me.page.waitForTimeout(500);
      expect(landsPlayed(), "the tap that closed the zoom played a Forest").toBe(0);

      // 3. A tap now plays one: the gestures did not break play.
      await touchAt(me.page, x, y, 60);
      await expect.poll(landsPlayed, { timeout: 10_000 }).toBe(1);
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });
});
