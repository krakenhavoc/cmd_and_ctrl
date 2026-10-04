import { expect, test, type Locator, type Page } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  openAdminClient,
  playerByID,
  seedHandWithCard,
  type AdminClient,
  type SnapshotCard,
  type SnapshotView,
} from "./s19-helpers";

// #2188 / ADR 0118 §§1–2: strict payment by default, and "Cast anyway
// (don't pay)", end to end.
//
// Neither seat seeds a setting: since ADR 0118 PR 4 a fresh browser
// starts with strict payment on, and this spec is the proof. First,
// with two untapped Forests, a click on Grizzly Bears taps both and
// casts it (§1: a clicked cast is stamped `auto_tap`). Then a card the
// seat cannot pay for still offers "Cast anyway (don't pay)" in its
// right-click popover. Choosing it casts nothing yet: the action dock
// asks "Cast Craw Wurm without paying its mana cost?" with Cast and
// Cancel (owner decision 6).
// Cancel closes it and the Wurm stays in hand. Choosing the row again,
// then Cast, puts the Wurm on the stack without paying, and the game
// log tells BOTH players: "<player> cast Craw Wurm without paying its
// mana cost" (owner decision 4).
//
// This spec selects the menu item, the dialog and the Cast and Cancel
// buttons by those exact names, which is what makes them a label
// contract (AGENTS.md §5, ADR 0111 §10). Renaming any of them is a
// breaking change.
//
// The board is staged through the admin WS, like the #1789 spec.

const BEARS = "Grizzly Bears"; // {1}{G}: exactly the two Forests
const WURM = "Craw Wurm"; // {4}{G}{G}: nothing on this board can pay it
const LAND = "Forest";
const ROW = "Cast anyway (don't pay)";
const DIALOG = `Cast ${WURM} without paying its mana cost?`;

// The ready ring is the `ready` class on the card element (Card.svelte).
const READY = /(^|\s)ready(\s|$)/;

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${BEARS}`,
      `1 ${WURM}`,
      `97 ${LAND}`,
    ].join("\n") + "\n"
  );
}

function forestsInHand(admin: AdminClient, playerID: string): SnapshotCard[] {
  return playerByID(admin.snapshot(), playerID).hand.cards.filter(
    (c) => c.name === LAND,
  );
}

// rightClick opens a hand card's popover the way the browser's
// right-click does: a `contextmenu` event on the card (Card.svelte
// handleContextMenu). It is dispatched rather than clicked because the
// hand is a fan of overlapping, animated cards, and a pointer right-click
// can wait out the whole test on a neighbour or the hover zoom that
// intercepts the pointer. What is under test is the menu, not the hit
// testing.
async function rightClick(card: Locator): Promise<void> {
  await expect(card).toBeVisible();
  await card.dispatchEvent("contextmenu", {
    bubbles: true,
    cancelable: true,
    button: 2,
  });
}

// leftClick is the same for the card's own click (Card.svelte
// handleClick), for the same reason.
async function leftClick(card: Locator): Promise<void> {
  await expect(card).toBeVisible();
  await card.dispatchEvent("click", {
    bubbles: true,
    cancelable: true,
    button: 0,
  });
}

async function openLog(page: Page): Promise<Locator> {
  const log = page.locator('[aria-label="game log"]');
  if (!(await log.isVisible())) {
    await page.getByRole("button", { name: "open game log" }).click();
  }
  await expect(log).toBeVisible();
  return log;
}

test.describe("#2188 Cast anyway (don't pay)", () => {
  test("a click taps the lands and casts; the row asks first; Cast casts unpaid and the log says so", async ({
    browser,
    request,
  }) => {
    test.slow();

    let first: JoinedPlayer | null = null;
    let second: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(
        request,
        adminToken,
        `Cast anyway 2188 ${Date.now()}`,
      );
      if (!game.invite_token)
        throw new Error("invite token missing on fresh game");

      // No settings seed: strict payment is the default (ADR 0118 §1).
      first = await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Seat One",
      );
      second = await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Seat Two",
      );

      // Both seats get the deck: the starting player is rolled (#1486).
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        first.playerID,
        makeDeck(),
      );
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        second.playerID,
        makeDeck(),
      );
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(
        adminToken,
        game.id,
        first.playerID,
        second.playerID,
      );
      await admin.sendActionAsPlayer(first.playerID, "keep_hand", {});
      await admin.sendActionAsPlayer(second.playerID, "keep_hand", {});
      await admin.waitFor((v) => v.state === "active", "game state active");
      await admin.waitFor(
        (v) =>
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the active player's precombat main",
        15_000,
      );

      const snap = admin.snapshot();
      const activeSeat = snap.turn?.active_seat ?? 0;
      const activeID = snap.seats[activeSeat]?.id;
      expect([first.playerID, second.playerID]).toContain(activeID);
      const me = activeID === first.playerID ? first : second;
      const them = me === first ? second : first;

      // --- stage the hand and the mana -------------------------------
      // seedHandWithCard draws until the named card surfaces, so the
      // hand fills up with Forests on the way.
      const bears = await seedHandWithCard(admin, me.playerID, BEARS);
      const wurm = await seedHandWithCard(admin, me.playerID, WURM);

      // Two Forests onto the battlefield (a sandbox move, which is not
      // a land drop and taps nothing).
      for (
        let i = 0;
        i < 6 && forestsInHand(admin, me.playerID).length < 2;
        i++
      ) {
        await admin.sendActionAsPlayer(me.playerID, "draw_card", {});
      }
      const forests = forestsInHand(admin, me.playerID).slice(0, 2);
      if (forests.length < 2) {
        throw new Error(`need two Forests in hand, have ${forests.length}`);
      }
      for (const f of forests) {
        await admin.sendActionAsPlayer(me.playerID, "move_card", {
          src: { kind: "hand", owner: me.playerID },
          dst: { kind: "battlefield" },
          instance_id: f.instance_id,
        });
      }
      const forestIDs = forests.map((f) => f.instance_id);
      const forestsOnBoard = (v: SnapshotView) =>
        v.battlefield.cards.filter((c) => forestIDs.includes(c.instance_id));
      await admin.waitFor(
        (v) =>
          forestsOnBoard(v).length === 2 &&
          forestsOnBoard(v).every((c) => !c.tapped),
        "two untapped Forests on the battlefield",
        10_000,
      );
      await admin.waitFor(
        (v) =>
          v.turn?.active_seat === activeSeat &&
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === activeSeat,
        "priority still with the active player in precombat main",
        10_000,
      );

      const page = me.page;
      const hand = page.locator('[aria-label="your hand"]');
      const inHand = (id: string): Locator =>
        hand.locator(`.card[data-instance-id="${id}"]`);

      // --- a click on Grizzly Bears taps both Forests and casts it -----
      // The ready ring says the server's move list has the cast, which
      // is also what wires the click (Hand.svelte).
      const bearsInHand = inHand(bears.instance_id);
      await expect(bearsInHand).toHaveClass(READY, { timeout: 15_000 });
      await leftClick(bearsInHand);

      // One click: no insufficient-mana request, no preview. The cast
      // was stamped strict + auto_tap, so the server tapped both
      // Forests for {1}{G} and put the Bears on the stack. Either seat's
      // client may pass priority on, so the Bears may have resolved
      // already.
      await admin.waitFor(
        (v) =>
          !playerByID(v, me.playerID).hand.cards.some(
            (c) => c.instance_id === bears.instance_id,
          ) &&
          ((v.stack_items ?? []).some(
            (s) => s.source_card_id === bears.instance_id,
          ) ||
            v.battlefield.cards.some(
              (c) => c.instance_id === bears.instance_id,
            )),
        "the Bears are on the stack (or resolved)",
        15_000,
      );
      const tapped = forestsOnBoard(admin.snapshot());
      expect(tapped).toHaveLength(2);
      expect(tapped.every((c) => c.tapped === true)).toBe(true);
      await expect(
        page.getByRole("dialog", { name: "insufficient mana" }),
      ).toHaveCount(0);

      // The Bears resolve and priority comes back to the active player
      // with an empty stack: the Wurm is a sorcery-speed cast.
      await admin.waitFor(
        (v) =>
          v.battlefield.cards.some(
            (c) => c.instance_id === bears.instance_id,
          ) &&
          (v.stack_items ?? []).length === 0 &&
          v.turn?.active_seat === activeSeat &&
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === activeSeat,
        "the Bears resolved, and priority is back with the active player in precombat main",
        20_000,
      );

      const wurmInHand = inHand(wurm.instance_id);
      await expect(wurmInHand).toBeVisible({ timeout: 15_000 });

      // --- the row opens the confirmation; Cancel sends nothing --------
      await rightClick(wurmInHand);
      await page.getByRole("menuitem", { name: ROW }).click();
      const dialog = page.getByRole("dialog", { name: DIALOG });
      await expect(dialog).toBeVisible();
      await dialog.getByRole("button", { name: "Cancel" }).click();
      await expect(dialog).toBeHidden();
      // Nothing was cast: the Wurm is still in hand, on the page and on
      // the server.
      await expect(wurmInHand).toBeVisible();
      expect(
        playerByID(admin.snapshot(), me.playerID).hand.cards.some(
          (c) => c.instance_id === wurm.instance_id,
        ),
      ).toBe(true);

      // --- again, and Cast: the Wurm is cast without paying ------------
      await rightClick(wurmInHand);
      await page.getByRole("menuitem", { name: ROW }).click();
      await expect(dialog).toBeVisible();
      await dialog.getByRole("button", { name: "Cast", exact: true }).click();
      await expect(dialog).toBeHidden();

      // The caster keeps priority after casting (CR 117.3c), but either
      // seat's client may pass it on, so the Wurm is on the stack or has
      // already resolved. Either way it left the hand unpaid.
      await admin.waitFor(
        (v) =>
          !playerByID(v, me.playerID).hand.cards.some(
            (c) => c.instance_id === wurm.instance_id,
          ) &&
          ((v.stack_items ?? []).some(
            (s) => s.source_card_id === wurm.instance_id,
          ) ||
            v.battlefield.cards.some(
              (c) => c.instance_id === wurm.instance_id,
            )),
        "the Wurm is on the stack (or resolved)",
        15_000,
      );

      // --- both players' logs say so -----------------------------------
      const line = `${me.name} cast ${WURM} without paying its mana cost`;
      for (const p of [me, them]) {
        const log = await openLog(p!.page);
        await expect(log.getByText(line, { exact: true })).toBeVisible({
          timeout: 15_000,
        });
      }
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });
});
