import { expect, test, type Browser, type APIRequestContext } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  findCardInPlayerGraveyard,
  findCardInPlayerHand,
  findCardOnBattlefield,
  keepAllHands,
  openAdminClient,
  playerByID,
  seedHandWithCard,
  type AdminClient,
  type SnapshotView,
} from "./s19-helpers";

// #3270 / CR 608.2g: "you may cast a spell from your hand without paying its
// mana cost" casts the spell NOW. CR 608.2g: "If an effect specifically
// instructs or allows a player to cast a spell during resolution, they do so
// by following the steps in rules 601.2a-i, except no player receives
// priority after it's cast."
//
// Surtland Elementalist ("whenever this creature attacks, you may cast an
// instant or sorcery spell from your hand without paying its mana cost") is
// the staging card: its trigger asks in the declare attackers step, where a
// sorcery can not be cast on its own. Since #4250 (#4282) the offer is the
// resolution_cast prompt, asked while the trigger is still resolving: a sheet
// in the dock that lists every way to cast each card on offer, targets
// included, and "Keep it in your hand". Nobody gets priority until it is
// answered, and the player never touches the card in their hand.
//
//   1. Divination (a sorcery, no target): its cast in the sheet puts it on
//      the stack at once, in the declare attackers step. Nobody clicked the
//      card. It resolves and draws two.
//   2. Lightning Bolt (an instant with a target): the sheet offers a cast at
//      each legal target; the one at the defender's creature casts it there.
//   3. "Keep it in your hand" casts nothing and leaves the card in hand.

const ELEMENTALIST = "Surtland Elementalist";
const DIVINATION = "Divination";
const BOLT = "Lightning Bolt";
const BEARS = "Grizzly Bears";

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${ELEMENTALIST}`,
      `1 ${DIVINATION}`,
      `1 ${BOLT}`,
      `1 ${BEARS}`,
      "95 Forest",
    ].join("\n") + "\n"
  );
}

// offerOpen: the Elementalist's free cast is being offered to `chooser`,
// inside the trigger's resolution (the resolution_cast prompt, #4250).
function offerOpen(v: SnapshotView, chooser: string): boolean {
  return (v.pending_choices ?? []).some(
    (c) => c.kind === "resolution_cast" && c.chooser === chooser && c.may_cast_keyword === "hand_cast",
  );
}

interface Staged {
  attacker: JoinedPlayer;
  defender: JoinedPlayer;
  admin: AdminClient;
}

// stage seats two players, puts the Elementalist on the attacker's side and
// the Bears on the defender's, seeds the attacker's hand, declares the attack
// and passes priority until the attacker's free-cast offer is open.
async function stage(
  browser: Browser,
  request: APIRequestContext,
  holder: { first?: JoinedPlayer; second?: JoinedPlayer; admin?: AdminClient },
  hand: string[],
): Promise<Staged> {
  const adminToken = await adminLogin(request);
  const game = await createGame(request, adminToken, `Free cast 3270 ${Date.now()}`);
  if (!game.invite_token) throw new Error("invite token missing on fresh game");
  const first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One");
  holder.first = first;
  const second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two");
  holder.second = second;
  for (const p of [first, second]) await p.page.setViewportSize({ width: 1600, height: 900 });
  await uploadDeckAs(request, adminToken, game.id, first.playerID, makeDeck());
  await uploadDeckAs(request, adminToken, game.id, second.playerID, makeDeck());
  await startGameAs(request, adminToken, game.id);

  const admin = await openAdminClient(adminToken, game.id, first.playerID, second.playerID);
  holder.admin = admin;
  await keepAllHands(admin);
  await admin.waitFor((v) => v.state === "active", "game state active");
  await admin.waitFor(
    (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
    "cursor settled on the attacker's precombat main",
    15_000,
  );

  const snap = admin.snapshot();
  const activeSeat = snap.turn?.active_seat ?? 0;
  const attacker = snap.seats[activeSeat]?.id === first.playerID ? first : second;
  const defender = attacker === first ? second : first;

  await adminMoveByName(admin, attacker.playerID, ELEMENTALIST, "library", "battlefield");
  await adminMoveByName(admin, defender.playerID, BEARS, "library", "battlefield");
  await admin.waitFor(
    (v) => [ELEMENTALIST, BEARS].every((n) => findCardOnBattlefield(v, n)),
    "both creatures are out",
  );
  // Clear summoning sickness (CR 302.6).
  await admin.sendActionAsPlayer(attacker.playerID, "untap_all", {});
  for (const name of hand) await seedHandWithCard(admin, attacker.playerID, name);

  await admin.sendAction("advance_step", {});
  await admin.waitFor((v) => v.turn?.step === "declare_attackers", "declare_attackers", 20_000);
  const elementalist = findCardOnBattlefield(admin.snapshot(), ELEMENTALIST)!;
  await admin.sendActionAsPlayer(attacker.playerID, "declare_attacker", {
    attacker: elementalist.instance_id,
    target: defender.playerID,
  });

  // Pass priority until the attack trigger resolves into the free-cast offer.
  const deadline = Date.now() + 60_000;
  for (;;) {
    const v = admin.snapshot();
    if (offerOpen(v, attacker.playerID)) break;
    if (Date.now() > deadline) throw new Error("the free-cast offer never opened");
    const t = v.turn;
    try {
      if (t && t.priority_holder >= 0) {
        const holderID = v.seats[t.priority_holder]?.id;
        if (holderID) await admin.sendActionAsPlayer(holderID, "pass_priority", {});
      }
    } catch {
      // Raced a browser's own auto-pass; look again.
    }
    await admin
      .waitFor(
        (n) =>
          n.turn?.priority_holder !== t?.priority_holder ||
          (n.pending_choices ?? []).length !== (v.pending_choices ?? []).length,
        "priority moved",
        3_000,
      )
      .catch(() => undefined);
  }
  return { attacker, defender, admin };
}

// stackHolds: a spell's stack item names its card in source_card_id (the
// label is for abilities).
function stackHolds(v: SnapshotView, cardID: string): boolean {
  return (v.stack_items ?? []).some((s) => s.source_card_id === cardID);
}

// stackTargets: the targets of the spell `cardID` on the stack.
function stackTargets(v: SnapshotView, cardID: string): string[] {
  const item = (v.stack_items ?? []).find((s) => s.source_card_id === cardID);
  return (item?.targets ?? []).map((t) => t.id);
}

test.describe("#3270 a free cast from hand is cast during the resolution", () => {
  test("a sorcery picked in the declare attackers step is cast at once, with no click on the card", async ({
    browser,
    request,
  }, testInfo) => {
    test.slow();
    const holder: { first?: JoinedPlayer; second?: JoinedPlayer; admin?: AdminClient } = {};
    try {
      const s = await stage(browser, request, holder, [DIVINATION, BOLT]);
      const page = s.attacker.page;
      const handBefore = playerByID(s.admin.snapshot(), s.attacker.playerID).hand.count;
      const divination = findCardInPlayerHand(s.admin.snapshot(), s.attacker.playerID, DIVINATION)!;
      expect(s.admin.snapshot().turn?.step).toBe("declare_attackers");

      // The offer is asked while the trigger resolves, over both cards.
      const prompt = page.getByRole("dialog", { name: /Surtland Elementalist/i });
      await expect(prompt).toBeVisible({ timeout: 20_000 });
      await expect(prompt).toContainText("is still resolving");
      await expect(prompt.getByRole("button", { name: /^Cast Lightning Bolt\b/ }).first()).toBeVisible();
      await testInfo.attach("the offer", { body: await page.screenshot(), contentType: "image/png" });
      await prompt.getByRole("button", { name: /^Cast Divination\b/ }).click();

      // The cast starts by itself: Divination is on the stack, in a step a
      // sorcery could never be cast in, and nothing clicked it in the hand.
      await s.admin.waitFor((v) => stackHolds(v, divination.instance_id), "Divination cast onto the stack", 20_000);
      expect(s.admin.snapshot().turn?.step).toBe("declare_attackers");
      // "Cast a spell": the one cast spent the offer, and the Bolt stays.
      await s.admin.waitFor((v) => !offerOpen(v, s.attacker.playerID), "the offer is spent", 10_000);
      expect(findCardInPlayerHand(s.admin.snapshot(), s.attacker.playerID, BOLT)).not.toBeNull();
      await testInfo.attach("Divination on the stack", {
        body: await page.screenshot(),
        contentType: "image/png",
      });

      // Let it resolve: both players pass.
      for (let i = 0; i < 8 && stackHolds(s.admin.snapshot(), divination.instance_id); i++) {
        const v = s.admin.snapshot();
        const t = v.turn;
        const holderID = t && t.priority_holder >= 0 ? v.seats[t.priority_holder]?.id : undefined;
        if (!holderID) break;
        await s.admin.sendActionAsPlayer(holderID, "pass_priority", {}).catch(() => undefined);
        await s.admin
          .waitFor((n) => n.turn?.priority_holder !== t?.priority_holder || !stackHolds(n, divination.instance_id), "passed", 3_000)
          .catch(() => undefined);
      }
      const after = await s.admin.waitFor(
        (v) => findCardInPlayerGraveyard(v, s.attacker.playerID, DIVINATION) !== null,
        "Divination resolved into the graveyard",
        20_000,
      );
      // The card left the hand and two came in: net +1.
      expect(playerByID(after, s.attacker.playerID).hand.count).toBe(handBefore + 1);
    } finally {
      holder.admin?.close();
      await closeAll(holder.first ?? null, holder.second ?? null);
    }
  });

  test("an instant on offer is cast at the target its line names", async ({ browser, request }, testInfo) => {
    test.slow();
    const holder: { first?: JoinedPlayer; second?: JoinedPlayer; admin?: AdminClient } = {};
    try {
      const s = await stage(browser, request, holder, [BOLT]);
      const page = s.attacker.page;
      const bolt = findCardInPlayerHand(s.admin.snapshot(), s.attacker.playerID, BOLT)!;
      const bears = findCardOnBattlefield(s.admin.snapshot(), BEARS)!;

      // Each legal target is its own line: both players and both creatures.
      const prompt = page.getByRole("dialog", { name: /Surtland Elementalist/i });
      await expect(prompt).toBeVisible({ timeout: 20_000 });
      const casts = prompt.getByRole("button", { name: /^Cast Lightning Bolt\b/ });
      await expect(casts).toHaveCount(4);
      await testInfo.attach("the targets on offer", { body: await page.screenshot(), contentType: "image/png" });
      await prompt.getByRole("button", { name: new RegExp(`^Cast Lightning Bolt\\b.* targeting ${BEARS}$`) }).click();

      await s.admin.waitFor((v) => stackHolds(v, bolt.instance_id), "Lightning Bolt cast onto the stack", 20_000);
      const v = s.admin.snapshot();
      expect(v.turn?.step).toBe("declare_attackers");
      expect(stackTargets(v, bolt.instance_id)).toEqual([bears.instance_id]);
      expect(findCardInPlayerHand(v, s.attacker.playerID, BOLT)).toBeNull();
    } finally {
      holder.admin?.close();
      await closeAll(holder.first ?? null, holder.second ?? null);
    }
  });

  test("keeping it in hand casts nothing", async ({ browser, request }) => {
    test.slow();
    const holder: { first?: JoinedPlayer; second?: JoinedPlayer; admin?: AdminClient } = {};
    try {
      const s = await stage(browser, request, holder, [DIVINATION]);
      const page = s.attacker.page;
      const divination = findCardInPlayerHand(s.admin.snapshot(), s.attacker.playerID, DIVINATION)!;
      const prompt = page.getByRole("dialog", { name: /Surtland Elementalist/i });
      await expect(prompt).toBeVisible({ timeout: 20_000 });
      await prompt.getByRole("button", { name: "Keep it in your hand", exact: true }).click();
      await s.admin.waitFor((v) => !offerOpen(v, s.attacker.playerID), "the offer was answered", 20_000);
      // Give a (wrong) auto-started cast time to show up.
      await page.waitForTimeout(1500);
      const v = s.admin.snapshot();
      expect(stackHolds(v, divination.instance_id)).toBe(false);
      expect(findCardInPlayerHand(v, s.attacker.playerID, DIVINATION)).not.toBeNull();
    } finally {
      holder.admin?.close();
      await closeAll(holder.first ?? null, holder.second ?? null);
    }
  });
});
