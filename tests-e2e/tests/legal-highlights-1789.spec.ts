import { expect, test, type Locator } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  findCardInZone,
  openAdminClient,
  playerByID,
  seedHandWithCard,
  type AdminClient,
  type SnapshotCard, keepAllHands } from "./s19-helpers";

// #1789 / ADR 0105 §9: the ready highlights, end to end.
//
// The client's unit and render tests feed legalActions.ts a digest by
// hand. This spec is the one place the REAL server's `legal_actions`
// reaches the real board: on your own main phase, with two untapped
// Forests,
//
//   - the Forest in your hand wears the ready ring (a land drop is
//     open) and its accessible name says "playable land";
//   - Grizzly Bears ({1}{G}) wears it and says "castable";
//   - Craw Wurm ({4}{G}{G}) does not: castable_here would open its
//     sorcery-speed window, but the enumerator cannot pay for it, so
//     the server lists no move and nothing lights;
//
// and after the turn passes to the opponent, nothing on your board
// wears it, because a seat that owes no decision gets no digest (§3).
//
// The board is staged through the admin WS, like the attack-all and
// S19 suites, because the property under test is what the server says
// is legal, not how the mana got there.

const CASTABLE = "Grizzly Bears"; // {1}{G}
const TOO_DEAR = "Craw Wurm"; // {4}{G}{G}
const LAND = "Forest";

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${CASTABLE}`,
      `1 ${TOO_DEAR}`,
      `97 ${LAND}`,
    ].join("\n") + "\n"
  );
}

// The ready ring is the `ready` class on the card element (Card.svelte).
const READY = /(^|\s)ready(\s|$)/;

function forestsInHand(admin: AdminClient, playerID: string): SnapshotCard[] {
  return playerByID(admin.snapshot(), playerID).hand.cards.filter((c) => c.name === LAND);
}

test.describe("#1789 ready highlights", () => {
  test("your main phase lights what you can play, and the opponent's turn lights nothing", async ({
    browser,
    request,
  }) => {
    test.slow();

    let first: JoinedPlayer | null = null;
    let second: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Ready highlights 1789 ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing on fresh game");

      first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One");
      second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two");

      // Both seats get the deck: the starting player is rolled
      // (#1486), so which seat is "you" is not known until the start.
      await uploadDeckAs(request, adminToken, game.id, first.playerID, makeDeck());
      await uploadDeckAs(request, adminToken, game.id, second.playerID, makeDeck());
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(adminToken, game.id, first.playerID, second.playerID);
      await keepAllHands(admin);
      await admin.waitFor((v) => v.state === "active", "game state active");
      await admin.waitFor(
        (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the active player's precombat main",
        15_000,
      );

      const snap = admin.snapshot();
      const activeSeat = snap.turn?.active_seat ?? 0;
      const activeID = snap.seats[activeSeat]?.id;
      expect([first.playerID, second.playerID]).toContain(activeID);
      const me = activeID === first.playerID ? first : second;

      // --- stage the hand and the mana -------------------------------
      // seedHandWithCard moves each card straight out of the library;
      // the Forests come from the opening hand, topped up below.
      const bears = await seedHandWithCard(admin, me.playerID, CASTABLE);
      const wurm = await seedHandWithCard(admin, me.playerID, TOO_DEAR);

      // Two Forests onto the battlefield (a sandbox move, which is not
      // a land drop) and at least one left in hand to play.
      for (let i = 0; i < 6 && forestsInHand(admin, me.playerID).length < 3; i++) {
        await admin.sendActionAsPlayer(me.playerID, "draw_card", {});
      }
      const forests = forestsInHand(admin, me.playerID);
      if (forests.length < 3) {
        throw new Error(`need three Forests in hand, have ${forests.length}`);
      }
      for (const f of forests.slice(0, 2)) {
        await admin.sendActionAsPlayer(me.playerID, "move_card", {
          src: { kind: "hand", owner: me.playerID },
          dst: { kind: "battlefield" },
          instance_id: f.instance_id,
        });
      }
      await admin.waitFor(
        (v) =>
          v.battlefield.cards.filter(
            (c) => c.name === LAND && c.controller === me.playerID && !c.tapped,
          ).length === 2,
        "two untapped Forests on the battlefield",
        10_000,
      );
      const handForest = findCardInZone(playerByID(admin.snapshot(), me.playerID).hand, LAND);
      if (!handForest) throw new Error("no Forest left in hand");

      // Still the active player's precombat main, with priority: the
      // admin moves above change neither.
      await admin.waitFor(
        (v) =>
          v.turn?.active_seat === activeSeat &&
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === activeSeat,
        "priority still with the active player in precombat main",
        10_000,
      );

      // --- your main phase: what you can play is lit -----------------
      const hand = me.page.locator('[aria-label="your hand"]');
      const inHand = (id: string): Locator => hand.locator(`.card[data-instance-id="${id}"]`);

      await expect(inHand(bears.instance_id)).toHaveClass(READY, { timeout: 15_000 });
      await expect(inHand(bears.instance_id)).toHaveAttribute(
        "aria-label",
        `${CASTABLE}, castable`,
      );
      await expect(inHand(handForest.instance_id)).toHaveClass(READY);
      await expect(inHand(handForest.instance_id)).toHaveAttribute(
        "aria-label",
        `${LAND}, playable land`,
      );
      // Its window is open, but the seat cannot pay six with two lands.
      await expect(inHand(wurm.instance_id)).toBeVisible();
      await expect(inHand(wurm.instance_id)).not.toHaveClass(READY);
      await expect(inHand(wurm.instance_id)).toHaveAttribute("aria-label", TOO_DEAR);

      // --- the opponent's turn: nothing is lit -----------------------
      await admin.sendActionAsPlayer(me.playerID, "pass_turn", {});
      await admin.waitFor(
        (v) =>
          v.turn?.active_seat !== activeSeat &&
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === v.turn?.active_seat,
        "cursor on the opponent's precombat main",
        20_000,
      );

      await expect(me.page.locator(".card.ready")).toHaveCount(0, { timeout: 15_000 });
      await expect(inHand(bears.instance_id)).not.toHaveClass(READY);
      await expect(inHand(handForest.instance_id)).not.toHaveClass(READY);
      await expect(inHand(bears.instance_id)).toHaveAttribute("aria-label", CASTABLE);
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });
});
