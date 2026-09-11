import { expect, test } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { makeCommanderDeck, COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  findCardOnBattlefield,
  openAdminClient,
  type AdminClient,
  type SnapshotView,
} from "./s19-helpers";

// #318 "[in-app] Feature request missing attack all button".
//
// The reporter's complaint was ergonomic: declaring a wide board one
// creature at a time is a chore. The fix is a cluster in the
// declare-attackers attention strip plus one bulk `declare_attackers`
// action, and the three things worth guarding end to end are:
//
//   1. one click declares EVERY eligible creature, at the opponent
//      named on the button (never a spread across the table);
//   2. creatures that can't attack are skipped silently — the count
//      on the button is honest before the click, and no error frame
//      comes back;
//   3. a SINGLE undo takes the whole declaration back, tap state
//      included. That is the whole reason the client sends one bulk
//      action instead of looping the per-creature verb: N actions
//      would mean N undo entries against a per-turn budget of one.
//
// The board is staged through the admin WS (same machinery as the
// S19 trigger suite) because casting four creatures through the UI
// would test the mana system, not the combat cluster.

// A deck whose non-basics are four vanilla green bears. Vanilla
// matters: anything with an ETB would put a trigger on the stack
// when admin-moved onto the battlefield and this test would be about
// the stack instead of about combat. Singleton applies to non-basics,
// hence four different bears rather than 4x one.
const BEARS = ["Grizzly Bears", "Runeclaw Bear", "Balduvian Bears", "Alpine Grizzly"] as const;

function makeBearDeck(): string {
  const lines = ["Commander:", `1 ${COMMANDER_NAME}`, "", "Mainboard:"];
  for (const bear of BEARS) lines.push(`1 ${bear}`);
  lines.push(`${99 - BEARS.length} Forest`);
  return lines.join("\n") + "\n";
}

function attackersAt(v: SnapshotView, defenderID: string): string[] {
  return v.battlefield.cards
    .filter((c) => (c as { attacking_target?: string }).attacking_target === defenderID)
    .map((c) => c.name ?? "?")
    .sort();
}

test.describe("#318 attack with all", () => {
  test("one click declares every eligible creature, one undo takes it back", async ({
    browser,
    request,
  }) => {
    test.slow();

    let attacker: JoinedPlayer | null = null;
    let defender: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Attack all 318 ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing on fresh game");

      attacker = await joinAsPlayer(browser, game.id, game.invite_token, "Attacker");
      defender = await joinAsPlayer(browser, game.id, game.invite_token, "Defender");

      await uploadDeckAs(request, adminToken, game.id, attacker.playerID, makeBearDeck());
      await uploadDeckAs(request, adminToken, game.id, defender.playerID, makeCommanderDeck());
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(adminToken, game.id, attacker.playerID, defender.playerID);
      await admin.sendActionAsPlayer(attacker.playerID, "keep_hand", {});
      await admin.sendActionAsPlayer(defender.playerID, "keep_hand", {});
      await admin.waitFor((v) => v.state === "active", "game state active");
      await admin.waitFor(
        (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the attacker's precombat main",
        15_000,
      );

      // Seat 0 is the attacker (first to join, first to act).
      const active = admin.snapshot().turn?.active_seat;
      expect(
        admin.snapshot().seats[active ?? 0]?.id,
        "attacker must be the active seat for the declare-attackers step",
      ).toBe(attacker.playerID);

      for (const bear of BEARS) {
        await adminMoveByName(admin, attacker.playerID, bear, "library", "battlefield");
      }
      await admin.waitFor(
        (v) => BEARS.every((b) => findCardOnBattlefield(v, b) !== null),
        "all four bears on the battlefield",
      );

      // Everything just entered, so everything is summoning-sick.
      // untap_all clears the flag the same way the untap step does
      // (CR 302.1), which is the cheapest way to get a board that has
      // "been around since last turn".
      await admin.sendActionAsPlayer(attacker.playerID, "untap_all", {});

      // Tap one bear so the run has a genuinely ineligible creature
      // in it. Tapped is the one eligibility reason that doesn't go
      // through the printed-keyword pipeline (#317), so it is the
      // stable one to assert on end to end.
      const benched = findCardOnBattlefield(admin.snapshot(), BEARS[3]);
      if (!benched) throw new Error(`${BEARS[3]} missing from the battlefield`);
      await admin.sendActionAsPlayer(attacker.playerID, "tap", {
        instance_id: benched.instance_id,
      });
      await admin.waitFor(
        (v) => findCardOnBattlefield(v, BEARS[3])?.tapped === true,
        `${BEARS[3]} tapped`,
      );

      // One admin step off the attacker's precombat_main stop. From
      // begin_combat both browsers auto-pass (it isn't a default stop),
      // so the cursor lands on declare_attackers by itself — issuing a
      // second advance_step here would race that auto-pass and
      // overshoot into declare_blockers.
      await admin.sendAction("advance_step", {});
      await admin.waitFor(
        (v) => v.turn?.step === "declare_attackers",
        "cursor on declare_attackers",
        20_000,
      );

      // --- 1 + 2: the honest count, then one click ---------------
      const cluster = attacker.page.locator(".att.attack-all");
      await expect(cluster).toBeVisible({ timeout: 15_000 });
      await expect(cluster).toContainText("3");
      await expect(cluster, "the cluster must say why the fourth is out").toContainText("1 tapped");

      // Two opponents would render one button per seat; a duel
      // renders the single named button. Either way the control names
      // the defender — "attack all" on its own is ambiguous in
      // Commander and the UI never leaves it unsaid.
      const attackAll = cluster.getByRole("button", { name: /Attack Defender with all 3/ });
      await expect(attackAll).toBeVisible();
      await attackAll.click();

      await admin.waitFor(
        (v) => attackersAt(v, defender!.playerID).length === 3,
        "three attackers declared at the defender",
        15_000,
      );
      const declared = attackersAt(admin.snapshot(), defender.playerID);
      expect(declared, "the tapped bear must be skipped silently").toEqual(
        [BEARS[0], BEARS[1], BEARS[2]].sort(),
      );
      // No creature may be pointed anywhere but the named seat.
      expect(attackersAt(admin.snapshot(), attacker.playerID)).toEqual([]);

      // --- 3: one undo reverses the whole declaration ------------
      const undo = cluster.getByRole("button", { name: /Undo/i });
      await expect(undo).toBeVisible();
      await undo.click();

      await admin.waitFor(
        (v) => attackersAt(v, defender!.playerID).length === 0,
        "one undo cleared every declaration",
        15_000,
      );
      // Tap state came back with it — that is the property a loop of
      // per-creature declarations could not have given us.
      const after = admin.snapshot();
      for (const bear of [BEARS[0], BEARS[1], BEARS[2]]) {
        expect(findCardOnBattlefield(after, bear)?.tapped ?? false, `${bear} untapped again`).toBe(
          false,
        );
      }
    } finally {
      admin?.close();
      await closeAll(attacker, defender);
    }
  });
});
