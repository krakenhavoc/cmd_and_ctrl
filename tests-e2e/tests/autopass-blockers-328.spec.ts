import { expect, test } from "@playwright/test";
import { adminLogin, createGame, uploadDeckAs, startGameAs } from "./lobby-api";
import { makeCommanderDeck, COMMANDER_NAME } from "./deck-fixture";
import { joinAsPlayer } from "./players";
import { openAdminClient, playerByID, findCardInZone, type SnapshotView } from "./s19-helpers";

// #328 — "blockers stage did not happen with autopass".
//
// The reporter had the autopass toggle engaged, the cursor reached
// the declare-blockers step on their opponent's turn, and their
// client passed priority 2ms after the snapshot arrived. Three
// attackers connected unblocked and they dropped 26 → 18 life with an
// untapped creature on the table the whole time.
//
// Root cause: blocking is a TURN-BASED ACTION (CR 509.1), not a
// response, so it was invisible to every auto-pass gate — the smart-
// skip predicate only ever asked "does this player have a legal
// response?", and the autopass toggle skipped even that. Nothing in
// the client distinguished the one window a defending player cannot
// afford to lose from a routine upkeep pass.
//
// This spec reproduces the exact board: the defender's only creature
// is summoning sick (Aang had just landed), which does NOT stop it
// blocking per CR 302.6. Before the fix the defender's client passes
// declare_blockers on its own and the cursor runs on to combat
// damage; after it, the window holds until a human acts.

test.describe("#328 autopass skips the blocking window", () => {
  test("autopass holds the defender's declare-blockers window", async ({ browser, request }) => {
    test.slow();

    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Blockers 328 ${Date.now()}`);
    if (!game.invite_token) throw new Error("invite token missing on fresh game");

    const alice = await joinAsPlayer(browser, game.id, game.invite_token, "Alice");
    const bob = await joinAsPlayer(browser, game.id, game.invite_token, "Bob");

    const deck = makeCommanderDeck();
    await uploadDeckAs(request, adminToken, game.id, alice.playerID, deck);
    await uploadDeckAs(request, adminToken, game.id, bob.playerID, deck);
    await startGameAs(request, adminToken, game.id);

    const admin = await openAdminClient(adminToken, game.id, alice.playerID, bob.playerID);

    try {
      await admin.sendActionAsPlayer(alice.playerID, "keep_hand", {});
      await admin.sendActionAsPlayer(bob.playerID, "keep_hand", {});
      await admin.waitFor((v) => v.state === "active", "game state active");
      await admin.waitFor(
        (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the active player's precombat main",
        15_000,
      );

      const seatIDs = admin.snapshot().seats.map((s) => s.id);
      const firstSeat = admin.snapshot().turn?.active_seat ?? 0;
      // The player who takes turn 1 becomes the DEFENDER: after one
      // pass_turn the other seat is active, and only the active
      // seat's untap has cleared summoning sickness. That asymmetry
      // is the reported board — the defender's creature is sick and
      // blocks anyway.
      const defenderID = seatIDs[firstSeat];
      const attackerID = seatIDs[(firstSeat + 1) % seatIDs.length];
      const defenderSeat = firstSeat;
      const defender = [alice, bob].find((p) => p.playerID === defenderID)!;
      const attacker = [alice, bob].find((p) => p.playerID === attackerID)!;

      // Each seat's commander is the only creature either deck has
      // (99 Plains otherwise), so put both on the battlefield.
      const putCommanderOut = async (playerID: string): Promise<void> => {
        const cmd = findCardInZone(playerByID(admin.snapshot(), playerID).command, COMMANDER_NAME);
        if (!cmd) throw new Error(`${COMMANDER_NAME} not in ${playerID}'s command zone`);
        await admin.sendActionAsPlayer(playerID, "move_card", {
          src: { kind: "command", owner: playerID },
          dst: { kind: "battlefield" },
          instance_id: cmd.instance_id,
        });
      };
      await putCommanderOut(defenderID);
      await putCommanderOut(attackerID);
      await admin.waitFor(
        (v) => v.battlefield.cards.filter((c) => c.name === COMMANDER_NAME).length === 2,
        "both commanders on the battlefield",
        10_000,
      );

      // Hand the turn over. The attacker's untap clears ITS creature's
      // summoning sickness so it can be declared; the defender's stays
      // sick, exactly as in the replay.
      await admin.sendActionAsPlayer(defenderID, "pass_turn", {});
      await admin.waitFor(
        (v) =>
          v.turn?.active_seat !== defenderSeat &&
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === v.turn?.active_seat,
        "cursor on the attacker's precombat main",
        20_000,
      );

      // The reproduction condition: the DEFENDER turns autopass on.
      // Done now, on the opponent's turn — the S13.6 safety belt
      // clears the toggle on the viewer's own precombat_main, so
      // enabling it during their own turn would just switch off.
      const autopassBtn = defender.page.locator("button.action.autopass");
      await expect(autopassBtn).toBeVisible({ timeout: 10_000 });
      await autopassBtn.click();
      await expect(autopassBtn).toHaveAttribute("aria-pressed", "true");

      const attackerSeat = admin.snapshot().turn?.active_seat ?? 0;
      // passAsAttacker rotates priority off the attacking seat. The
      // attacker's browser runs default settings and stops on the
      // combat declarations; the defender's browser has the toggle on
      // and passes its own windows, which is the whole point.
      const passAsAttacker = async (atStep: string): Promise<void> => {
        await admin.waitFor(
          (v) => v.turn?.step === atStep && v.turn?.priority_holder === attackerSeat,
          `attacker holds priority at ${atStep}`,
          20_000,
        );
        await admin.sendActionAsPlayer(attackerID, "pass_priority", {});
      };

      await passAsAttacker("precombat_main");
      await admin.waitFor(
        (v) => v.turn?.step === "declare_attackers" && v.turn?.priority_holder === attackerSeat,
        "cursor reached declare_attackers",
        20_000,
      );

      const atkCard = admin
        .snapshot()
        .battlefield.cards.find((c) => c.name === COMMANDER_NAME && c.controller === attackerID);
      if (!atkCard) throw new Error("attacker's commander missing from the battlefield");
      await admin.sendActionAsPlayer(attackerID, "declare_attacker", {
        attacker: atkCard.instance_id,
        target: defenderID,
      });
      await admin.waitFor(
        (v) =>
          v.battlefield.cards.some(
            (c) => c.instance_id === atkCard.instance_id && c.attacking_target === defenderID,
          ),
        "attack declared against the defender",
        10_000,
      );

      await passAsAttacker("declare_attackers");
      await admin.waitFor(
        (v) => v.turn?.step === "declare_blockers",
        "cursor reached declare_blockers",
        20_000,
      );

      // The engine must be telling the table that the defender owes a
      // block decision — the wire contract the client's guard reads.
      const blockSeats = (v: SnapshotView): number[] => v.turn?.block_decision_seats ?? [];
      expect(
        blockSeats(admin.snapshot()),
        "server must flag the defender as owing a block decision",
      ).toContain(defenderSeat);

      // Hand priority to the defender. This is the moment that failed:
      // their client fired pass_priority within a couple of ms.
      await passAsAttacker("declare_blockers");
      await admin.waitFor(
        (v) => v.turn?.priority_holder === defenderSeat,
        "priority reached the defender in declare_blockers",
        20_000,
      );

      // Hold the line. Auto-pass reacts to the snapshot it just got,
      // so a few seconds is many orders of magnitude more than the
      // 2ms the reporter's client took.
      await defender.page.waitForTimeout(4_000);

      const settled = admin.snapshot();
      expect(
        settled.turn?.step,
        "the defender's blocking window was auto-passed — #328",
      ).toBe("declare_blockers");
      expect(
        settled.turn?.priority_holder,
        "the defender must still hold priority to declare blockers",
      ).toBe(defenderSeat);

      // And the defender's board says so too.
      const stepLabel = (
        (await defender.page.locator(".step-label").first().textContent()) ?? ""
      ).trim();
      expect(stepLabel).toBe("Declare Blockers");

      // Declining is still legal: an explicit pass moves the game on,
      // which is what separates "the player chose not to block" from
      // "the client chose for them".
      await admin.sendActionAsPlayer(defenderID, "pass_priority", {});
      await admin.waitFor(
        (v) => v.turn?.step !== "declare_blockers",
        "an explicit pass still leaves the blocking window",
        20_000,
      );
      void attacker;
    } finally {
      admin.close();
      await alice.context.close();
      await bob.context.close();
    }
  });
});
