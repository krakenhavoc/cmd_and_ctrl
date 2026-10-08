import { expect, test, type Page } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  findCardOnBattlefield,
  keepAllHands,
  openAdminClient,
  type AdminClient,
  type SnapshotView,
} from "./s19-helpers";

// #2614 / ADR 0134: combat you can watch.
//
// When combat damage lands, the attacker's art-only copy lunges at what
// it hit and snaps back, while its live tile hides for the flight. This
// guards the two things only a real browser shows:
//
//   1. with combat motion on, a strike copy (`[data-strike-copy]`) is
//      drawn during combat damage, and within 2 s it is gone and the
//      attacker's tile is visible again;
//   2. with `animations.combat` off, no copy is ever drawn, and the life
//      total still changes.
//
// The setting is seeded into each browser's settings blob rather than
// toggled through the Settings row, so no label is selected on.
//
// The board is staged through the admin WS (the #318 spec's machinery):
// one vanilla bear on the attacker's side, attacking unblocked.

const BEAR = "Grizzly Bears";
const SETTINGS_VERSION = 21;

function makeBearDeck(): string {
  return ["Commander:", `1 ${COMMANDER_NAME}`, "", "Mainboard:", `1 ${BEAR}`, "98 Forest"].join(
    "\n",
  ) + "\n";
}

function lifeOf(v: SnapshotView, playerID: string): number {
  const seat = v.seats.find((s) => s.id === playerID) as unknown as { life?: number } | undefined;
  return seat?.life ?? NaN;
}

// Count every strike copy the page ever mounts, so a copy that lives
// half a second is not missed between polls.
async function watchStrikes(page: Page): Promise<void> {
  await page.evaluate(() => {
    const w = window as unknown as { __strikes: number };
    w.__strikes = 0;
    new MutationObserver((records) => {
      for (const r of records) {
        for (const n of r.addedNodes) {
          if (!(n instanceof Element)) continue;
          if (n.matches("[data-strike-copy]")) w.__strikes++;
          else w.__strikes += n.querySelectorAll("[data-strike-copy]").length;
        }
      }
    }).observe(document.body, { childList: true, subtree: true });
  });
}

async function strikesSeen(page: Page): Promise<number> {
  return page.evaluate(() => (window as unknown as { __strikes: number }).__strikes);
}

test.describe("#2614 combat you can watch", () => {
  test("an unblocked attacker lunges; with combat motion off, nothing moves", async ({
    browser,
    request,
  }, testInfo) => {
    test.slow();

    let first: JoinedPlayer | null = null;
    let second: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Combat strike 2614 ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing on fresh game");

      // Seat One watches at half speed (a 1 s strike, easy to catch);
      // Seat Two has combat motion off. Which of them attacks is the
      // opening roll's call, so the assertions below follow the roles.
      first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One", {
        __version: SETTINGS_VERSION,
        animations: { speed: 2 },
      });
      second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two", {
        __version: SETTINGS_VERSION,
        animations: { combat: false },
      });
      await uploadDeckAs(request, adminToken, game.id, first.playerID, makeBearDeck());
      await uploadDeckAs(request, adminToken, game.id, second.playerID, makeBearDeck());
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(adminToken, game.id, first.playerID, second.playerID);
      await keepAllHands(admin);
      await admin.waitFor((v) => v.state === "active", "game state active");
      await admin.waitFor(
        (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the attacker's precombat main",
        15_000,
      );

      const snap = admin.snapshot();
      const activeSeat = snap.turn?.active_seat ?? 0;
      const attackerID = snap.seats[activeSeat]?.id;
      const attacker = attackerID === first.playerID ? first : second;
      const defender = attacker === first ? second : first;
      const defenderSeat = snap.seats.findIndex((s) => s.id === defender.playerID);

      await adminMoveByName(admin, attacker.playerID, BEAR, "library", "battlefield");
      await admin.waitFor((v) => findCardOnBattlefield(v, BEAR) !== null, "the bear is out");
      // Clear summoning sickness, as the #318 spec does (CR 302.6).
      await admin.sendActionAsPlayer(attacker.playerID, "untap_all", {});
      const bear = findCardOnBattlefield(admin.snapshot(), BEAR);
      if (!bear) throw new Error(`${BEAR} missing from the battlefield`);

      await admin.sendAction("advance_step", {});
      await admin.waitFor(
        (v) => v.turn?.step === "declare_attackers",
        "cursor on declare_attackers",
        20_000,
      );
      await admin.sendActionAsPlayer(attacker.playerID, "declare_attacker", {
        attacker: bear.instance_id,
        target: defender.playerID,
      });
      await admin.waitFor(
        (v) =>
          v.battlefield.cards.some(
            (c) => c.instance_id === bear.instance_id && c.attacking_target === defender.playerID,
          ),
        "the bear attacks the defender",
      );

      await watchStrikes(first.page);
      await watchStrikes(second.page);
      const strikeCopy = first.page.locator("[data-strike-copy]");
      // Started before the damage, so the copy cannot come and go unseen.
      const copyMounted = first.page
        .waitForSelector("[data-strike-copy]", { state: "attached", timeout: 60_000 })
        .then(
          () => true,
          () => false,
        );

      // Walk combat to its damage: the defender declares no blocks, and
      // whoever holds priority passes. A browser's own auto-pass may get
      // there first, so a refused pass is fine.
      const lifeBefore = lifeOf(admin.snapshot(), defender.playerID);
      const deadline = Date.now() + 60_000;
      while (lifeOf(admin.snapshot(), defender.playerID) === lifeBefore) {
        if (Date.now() > deadline) throw new Error("combat damage never landed");
        const v = admin.snapshot();
        const t = v.turn;
        try {
          if (t?.step === "declare_blockers" && (t.block_pending_seats ?? []).includes(defenderSeat)) {
            await admin.sendActionAsPlayer(defender.playerID, "finish_blocks", {});
          } else if (t && t.priority_holder >= 0) {
            const holder = v.seats[t.priority_holder]?.id;
            if (holder) await admin.sendActionAsPlayer(holder, "pass_priority", {});
          }
        } catch {
          // Raced a browser's auto-pass; look again.
        }
        await admin
          .waitFor(
            (n) =>
              lifeOf(n, defender.playerID) !== lifeBefore ||
              n.turn?.step !== t?.step ||
              n.turn?.priority_holder !== t?.priority_holder,
            "combat moved on",
            3_000,
          )
          .catch(() => undefined);
      }
      expect(lifeOf(admin.snapshot(), defender.playerID)).toBe(lifeBefore - 2);

      // --- 1: Seat One (motion on) sees the strike ------------------
      expect(await copyMounted, "a strike copy is drawn during combat damage").toBe(true);
      await testInfo.attach("strike in flight (Seat One)", {
        body: await first.page.screenshot(),
        contentType: "image/png",
      });
      await expect(strikeCopy, "the copy is gone within 2 s").toHaveCount(0, { timeout: 2_500 });
      expect(await strikesSeen(first.page)).toBeGreaterThan(0);
      await expect(
        first.page.locator(`[data-instance-id="${bear.instance_id}"]`).first(),
        "the attacker's tile is visible again",
      ).toBeVisible();

      // --- 2: Seat Two (combat off) sees no copy, and the life change -
      await expect(
        second.page.locator(`[data-seat-id="${defender.playerID}"]`).first(),
      ).toHaveAttribute("aria-label", new RegExp(`${lifeBefore - 2} life`), { timeout: 10_000 });
      await second.page.waitForTimeout(1_500);
      expect(await strikesSeen(second.page), "no strike copy with combat motion off").toBe(0);
      await testInfo.attach("after combat (Seat Two, motion off)", {
        body: await second.page.screenshot(),
        contentType: "image/png",
      });
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });
});
