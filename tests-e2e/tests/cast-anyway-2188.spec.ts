import { expect, test, type Locator, type Page } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import { openAdminClient, playerByID, seedHandWithCard, type AdminClient } from "./s19-helpers";

// #2188 / ADR 0118 §2: "Cast anyway (don't pay)", end to end.
//
// With strict payment on (opt-in until ADR 0118 PR 4 flips the
// default, so both seats seed it), a card the seat cannot pay for still
// offers "Cast anyway (don't pay)" in its right-click popover. Choosing
// it casts nothing yet: the action dock asks "Cast Craw Wurm without
// paying its mana cost?" with Cast and Cancel (owner decision 6).
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

const WURM = "Craw Wurm"; // {4}{G}{G}: nothing on this board can pay it
const LAND = "Forest";
const ROW = "Cast anyway (don't pay)";
const DIALOG = `Cast ${WURM} without paying its mana cost?`;

const STRICT = { gameplay: { strictMana: true } };

function makeDeck(): string {
  return (
    ["Commander:", `1 ${COMMANDER_NAME}`, "", "Mainboard:", `1 ${WURM}`, `98 ${LAND}`].join(
      "\n",
    ) + "\n"
  );
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
  test("the row asks first; Cancel sends nothing; Cast casts unpaid and the log says so", async ({
    browser,
    request,
  }) => {
    test.slow();

    let first: JoinedPlayer | null = null;
    let second: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Cast anyway 2188 ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing on fresh game");

      first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One", STRICT);
      second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two", STRICT);

      // Both seats get the deck: the starting player is rolled (#1486).
      await uploadDeckAs(request, adminToken, game.id, first.playerID, makeDeck());
      await uploadDeckAs(request, adminToken, game.id, second.playerID, makeDeck());
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(adminToken, game.id, first.playerID, second.playerID);
      await admin.sendActionAsPlayer(first.playerID, "keep_hand", {});
      await admin.sendActionAsPlayer(second.playerID, "keep_hand", {});
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
      const them = me === first ? second : first;

      const wurm = await seedHandWithCard(admin, me.playerID, WURM);
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
      const wurmInHand = hand.locator(`.card[data-instance-id="${wurm.instance_id}"]`);
      await expect(wurmInHand).toBeVisible({ timeout: 15_000 });

      // --- the row opens the confirmation; Cancel sends nothing --------
      await wurmInHand.click({ button: "right" });
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
      await wurmInHand.click({ button: "right" });
      await page.getByRole("menuitem", { name: ROW }).click();
      await expect(dialog).toBeVisible();
      await dialog.getByRole("button", { name: "Cast", exact: true }).click();
      await expect(dialog).toBeHidden();

      // The caster keeps priority after casting (CR 117.3c), but either
      // seat's client may pass it on, so the Wurm is on the stack or has
      // already resolved. Either way it left the hand unpaid.
      await admin.waitFor(
        (v) =>
          !playerByID(v, me.playerID).hand.cards.some((c) => c.instance_id === wurm.instance_id) &&
          ((v.stack_items ?? []).some((s) => s.source_card_id === wurm.instance_id) ||
            v.battlefield.cards.some((c) => c.instance_id === wurm.instance_id)),
        "the Wurm is on the stack (or resolved)",
        15_000,
      );

      // --- both players' logs say so -----------------------------------
      const line = `${me.name} cast ${WURM} without paying its mana cost`;
      for (const p of [me, them]) {
        const log = await openLog(p!.page);
        await expect(log.getByText(line, { exact: true })).toBeVisible({ timeout: 15_000 });
      }
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });
});
