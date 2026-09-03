import { expect, test } from "@playwright/test";
import { CARDS } from "./s19-deck-fixture";
import {
  adminMoveByName,
  findCardInPlayerGraveyard,
  findCardInPlayerHand,
  findCardOnBattlefield,
  playerByID,
  resolveStack,
  seedHandWithCard,
  setupS19Game,
  triggerOnStack,
  type S19Setup,
} from "./s19-helpers";

// S19 trigger e2e suite. Each test spins up a fresh 2-player game
// (≈10s setup) and exercises one S19 trigger end-to-end through the
// real WS path: admin moves a card onto the battlefield, the trigger
// harvester fires, the chooser's browser sees the modal (or doesn't,
// for mandatory triggers), and the test asserts on the resulting
// snapshot.
//
// All assertions read from the admin WS snapshot — the same view
// the server broadcast to every client — so a missing trigger,
// wrong target pick, or stuck PendingChoice surfaces as a hard
// failure here. The browser-side dialog assertions cover the UI
// rendering layer that the Go integration tests can't reach.
//
// Triggers use the stack: a harvested trigger (or an optional one
// answered "Yes") sits in stack_items until both players pass
// priority. Each test asserts the trigger is waiting there, then
// resolves it through the players' own "next" buttons via
// resolveStack before asserting on the effect.

test.describe("S19 ETB triggers", () => {
  test.describe.configure({ mode: "serial" });
  let setup: S19Setup | null = null;

  test.afterEach(async () => {
    if (setup) {
      await setup.shutdown();
      setup = null;
    }
  });

  test("Mulldrifter mandatory ETB draws 2 cards (no prompt)", async ({ browser, request }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;

    // Seed Mulldrifter into the caster's hand FIRST, then capture
    // the pre-move hand size — the seed step itself drew an
    // unpredictable number of cards out of the library, so any
    // baseline taken before seeding would be off by N.
    const mulldrifter = await seedHandWithCard(admin, caster.playerID, CARDS.Mulldrifter);
    const handBefore = playerByID(admin.snapshot(), caster.playerID).hand.count;

    await admin.sendActionAsPlayer(caster.playerID, "move_card", {
      src: { kind: "hand", owner: caster.playerID },
      dst: { kind: "battlefield" },
      instance_id: mulldrifter.instance_id,
    });

    // The mandatory trigger goes straight onto the stack; the hand
    // only shrinks by the Mulldrifter itself until it resolves.
    const staged = await admin.waitFor(
      (v) => triggerOnStack(v, CARDS.Mulldrifter) !== null,
      "Mulldrifter ETB trigger on the stack",
    );
    expect(playerByID(staged, caster.playerID).hand.count).toBe(handBefore - 1);
    // No trigger prompt should queue for a mandatory ability.
    expect(staged.pending_choices ?? []).toHaveLength(0);
    // The stack overlay shows the trigger to both seats.
    // (The overlay renders the label twice — fallback art + title —
    // so anchor on the first match.)
    await expect(caster.page.getByText(/Mulldrifter — draw two cards/i).first()).toBeVisible();
    await expect(opponent.page.getByText(/Mulldrifter — draw two cards/i).first()).toBeVisible();

    await resolveStack(setup);

    // Mulldrifter leaves the hand (-1) and the trigger draws 2 (+2)
    // → net hand delta is +1.
    const after = await admin.waitFor(
      (v) => playerByID(v, caster.playerID).hand.count === handBefore - 1 + 2,
      `caster hand reaches ${handBefore + 1} (Mulldrifter exits hand, ETB draws 2)`,
    );
    const handAfter = playerByID(after, caster.playerID).hand.count;
    expect(handAfter).toBe(handBefore + 1);
    expect(after.pending_choices ?? []).toHaveLength(0);

    // Caster's browser should not have a trigger prompt dialog.
    await expect(caster.page.getByRole("dialog", { name: /trigger/i })).toHaveCount(0);
  });

  test("Reclamation Sage optional ETB → Yes destroys opponent's artifact", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;

    // Stage the opponent's artifact on their battlefield.
    await adminMoveByName(admin, opponent.playerID, CARDS.SolRing, "library", "battlefield");
    await admin.waitFor(
      (v) => findCardOnBattlefield(v, CARDS.SolRing) !== null,
      "Sol Ring on battlefield",
    );

    // Caster's Reclamation Sage hits the battlefield → optional
    // prompt queues for the caster.
    await adminMoveByName(
      admin,
      caster.playerID,
      CARDS.ReclamationSage,
      "library",
      "battlefield",
    );

    const promptVisible = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "trigger prompt queued for caster",
    );
    const prompt = (promptVisible.pending_choices ?? []).find(
      (c) => c.kind === "trigger_prompt",
    );
    expect(prompt?.reason).toMatch(/Reclamation Sage/i);

    // Caster's browser shows the dialog; opponent's does not.
    await expect(caster.page.getByText(/Reclamation Sage/i, { exact: false })).toBeVisible({
      timeout: 5000,
    });
    await expect(opponent.page.getByText(/Reclamation Sage —/i, { exact: false })).toHaveCount(0);

    // Click "Yes" in the caster's dialog.
    await caster.page.getByRole("button", { name: /^Yes$/ }).click();

    // "Yes" puts the trigger on the stack; the Sol Ring survives
    // until it resolves.
    const staged = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).length === 0 &&
        triggerOnStack(v, CARDS.ReclamationSage) !== null,
      "Reclamation Sage trigger on the stack after Yes",
    );
    expect(findCardOnBattlefield(staged, CARDS.SolRing)).not.toBeNull();

    await resolveStack(setup);

    // The Sol Ring should be in the opponent's graveyard, off the
    // battlefield, and the prompt should clear.
    const after = await admin.waitFor(
      (v) =>
        findCardOnBattlefield(v, CARDS.SolRing) === null &&
        findCardInPlayerGraveyard(v, opponent.playerID, CARDS.SolRing) !== null,
      "Sol Ring routed to opponent's graveyard",
    );
    expect(after.pending_choices ?? []).toHaveLength(0);
  });

  test("Reclamation Sage optional ETB → No leaves artifact untouched", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;

    await adminMoveByName(admin, opponent.playerID, CARDS.SolRing, "library", "battlefield");
    await adminMoveByName(
      admin,
      caster.playerID,
      CARDS.ReclamationSage,
      "library",
      "battlefield",
    );
    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "trigger prompt queued",
    );

    await caster.page.getByRole("button", { name: /^No$/ }).click();

    const after = await admin.waitFor(
      (v) => (v.pending_choices ?? []).length === 0,
      "prompt drained after No",
    );
    // Declined: nothing reaches the stack, Sol Ring still on
    // battlefield, not in graveyard.
    expect(triggerOnStack(after, CARDS.ReclamationSage)).toBeNull();
    expect(findCardOnBattlefield(after, CARDS.SolRing)).not.toBeNull();
    expect(findCardInPlayerGraveyard(after, opponent.playerID, CARDS.SolRing)).toBeNull();
  });

  test("Acidic Slime mandatory ETB destroys opponent's artifact (no prompt)", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;

    await adminMoveByName(admin, opponent.playerID, CARDS.SolRing, "library", "battlefield");

    await adminMoveByName(admin, caster.playerID, CARDS.AcidicSlime, "library", "battlefield");

    const staged = await admin.waitFor(
      (v) => triggerOnStack(v, CARDS.AcidicSlime) !== null,
      "Acidic Slime ETB trigger on the stack",
    );
    expect(findCardOnBattlefield(staged, CARDS.SolRing)).not.toBeNull();
    expect(staged.pending_choices ?? []).toHaveLength(0);

    await resolveStack(setup);

    const after = await admin.waitFor(
      (v) => findCardInPlayerGraveyard(v, opponent.playerID, CARDS.SolRing) !== null,
      "Sol Ring destroyed by Acidic Slime",
    );
    expect(findCardOnBattlefield(after, CARDS.SolRing)).toBeNull();
    // Mandatory trigger — no pending choice queued.
    expect(after.pending_choices ?? []).toHaveLength(0);
  });

  test("Eternal Witness optional ETB → Yes returns top of graveyard to hand", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster } = setup;

    // Seed graveyard: admin moves Lightning Bolt from caster's
    // library to graveyard.
    await adminMoveByName(admin, caster.playerID, CARDS.LightningBolt, "library", "graveyard");
    await admin.waitFor(
      (v) => findCardInPlayerGraveyard(v, caster.playerID, CARDS.LightningBolt) !== null,
      "Lightning Bolt in caster graveyard",
    );

    await adminMoveByName(
      admin,
      caster.playerID,
      CARDS.EternalWitness,
      "library",
      "battlefield",
    );

    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "Eternal Witness prompt queued",
    );

    await expect(caster.page.getByText(/Eternal Witness/i, { exact: false })).toBeVisible();

    await caster.page.getByRole("button", { name: /^Yes$/ }).click();

    const staged = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).length === 0 &&
        triggerOnStack(v, CARDS.EternalWitness) !== null,
      "Eternal Witness trigger on the stack after Yes",
    );
    expect(findCardInPlayerHand(staged, caster.playerID, CARDS.LightningBolt)).toBeNull();

    await resolveStack(setup);

    const after = await admin.waitFor(
      (v) => findCardInPlayerHand(v, caster.playerID, CARDS.LightningBolt) !== null,
      "Lightning Bolt returned to caster's hand",
    );
    expect(findCardInPlayerGraveyard(after, caster.playerID, CARDS.LightningBolt)).toBeNull();
    expect(after.pending_choices ?? []).toHaveLength(0);
  });

  test("Eternal Witness optional ETB → No leaves graveyard untouched", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster } = setup;

    await adminMoveByName(admin, caster.playerID, CARDS.LightningBolt, "library", "graveyard");
    await adminMoveByName(
      admin,
      caster.playerID,
      CARDS.EternalWitness,
      "library",
      "battlefield",
    );
    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "prompt queued",
    );

    await caster.page.getByRole("button", { name: /^No$/ }).click();

    const after = await admin.waitFor(
      (v) => (v.pending_choices ?? []).length === 0,
      "prompt drained after No",
    );
    // Lightning Bolt still in graveyard; nothing returned to hand.
    expect(findCardInPlayerGraveyard(after, caster.playerID, CARDS.LightningBolt)).not.toBeNull();
    expect(findCardInPlayerHand(after, caster.playerID, CARDS.LightningBolt)).toBeNull();
  });

  test("Solemn Simulacrum optional ETB → Yes fetches Forest tapped", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster } = setup;

    // Library card identities are redacted on the wire (CR 400.2
    // private zone), so we can't search by Forest name directly.
    // Confirm there's a non-trivial library to fetch from instead —
    // the caster deck ships with 91 Forests + 8 non-basics + commander.
    expect(playerByID(admin.snapshot(), caster.playerID).library.count).toBeGreaterThan(50);

    await adminMoveByName(
      admin,
      caster.playerID,
      CARDS.SolemnSimulacrum,
      "library",
      "battlefield",
    );

    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "Solemn prompt queued",
    );
    await expect(caster.page.getByText(/Solemn Simulacrum/i, { exact: false })).toBeVisible();

    await caster.page.getByRole("button", { name: /^Yes$/ }).click();

    const staged = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).length === 0 &&
        triggerOnStack(v, CARDS.SolemnSimulacrum) !== null,
      "Solemn trigger on the stack after Yes",
    );
    expect(findCardOnBattlefield(staged, CARDS.Forest)).toBeNull();

    await resolveStack(setup);

    // A Forest now lives on the battlefield, tapped, controlled by
    // the caster.
    const after = await admin.waitFor(
      (v) => {
        const forest = findCardOnBattlefield(v, CARDS.Forest);
        return forest !== null && forest.controller === caster.playerID;
      },
      "fetched Forest on battlefield under caster's control",
    );
    const forest = findCardOnBattlefield(after, CARDS.Forest);
    expect(forest?.tapped).toBe(true);
    expect(after.pending_choices ?? []).toHaveLength(0);
  });

  test("Solemn Simulacrum optional ETB → No leaves library untouched", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster } = setup;

    await adminMoveByName(
      admin,
      caster.playerID,
      CARDS.SolemnSimulacrum,
      "library",
      "battlefield",
    );
    // Capture library count AFTER the seed-and-move step settles,
    // BEFORE we resolve the prompt. The seed step drew an unknown
    // number of cards into the caster's hand; we want to assert
    // that the decline path doesn't fetch anything else from the
    // library on top of that.
    const promptView = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "prompt queued",
    );
    const libraryAtPrompt = playerByID(promptView, caster.playerID).library.count;

    await caster.page.getByRole("button", { name: /^No$/ }).click();

    const after = await admin.waitFor(
      (v) => (v.pending_choices ?? []).length === 0,
      "prompt drained",
    );
    const libraryAfter = playerByID(after, caster.playerID).library.count;
    // Decline doesn't fetch — the library count at decline time
    // should equal the library count when the prompt queued.
    expect(libraryAfter).toBe(libraryAtPrompt);
    expect(findCardOnBattlefield(after, CARDS.Forest)).toBeNull();
  });

  test("Optional trigger prompt is visible only to the chooser", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;

    await adminMoveByName(admin, opponent.playerID, CARDS.SolRing, "library", "battlefield");
    await adminMoveByName(
      admin,
      caster.playerID,
      CARDS.ReclamationSage,
      "library",
      "battlefield",
    );

    // Wait for the prompt to register on the wire.
    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "prompt queued",
    );

    // Caster sees the dialog with Yes/No buttons; opponent has no
    // trigger dialog at all (the only legitimate dialog they could
    // see is mulligan, which has long since closed).
    await expect(caster.page.getByRole("button", { name: /^Yes$/ })).toBeVisible({
      timeout: 5000,
    });
    await expect(caster.page.getByRole("button", { name: /^No$/ })).toBeVisible();
    await expect(opponent.page.getByText(/Reclamation Sage —/i)).toHaveCount(0);
    await expect(opponent.page.getByRole("button", { name: /^Yes$/ })).toHaveCount(0);
  });

  test("Trigger prompt Question copy reaches the dialog header", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster } = setup;

    await adminMoveByName(
      admin,
      caster.playerID,
      CARDS.SolemnSimulacrum,
      "library",
      "battlefield",
    );

    const v = await admin.waitFor(
      (s) =>
        (s.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "Solemn prompt queued",
    );
    const prompt = (v.pending_choices ?? []).find((c) => c.kind === "trigger_prompt");
    // Server-side Question text from the Spec.
    expect(prompt?.reason).toMatch(/Solemn Simulacrum.*basic land/i);

    // Client-side modal should render the same text — proves the
    // wire shape's `reason` field flows through to the dialog
    // without truncation or escape errors.
    await expect(
      caster.page.getByText(prompt!.reason!, { exact: false }),
    ).toBeVisible({ timeout: 5000 });
  });

  // S19 sub-PR 6: pay-unless. The opponent's Smothering Tithe taxes
  // the caster's draw; the caster (not the Tithe's controller) gets
  // the pay prompt, declines, and the opponent gets a Treasure.
  test("Smothering Tithe: opponent's draw → pay {2} prompt → Don't pay → Treasure", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;

    // Stage the Tithe under the opponent AFTER any hand seeding —
    // seeding draws cards, and each draw would trigger it.
    await adminMoveByName(admin, opponent.playerID, CARDS.SmotheringTithe, "library", "battlefield");
    await admin.waitFor(
      (v) => findCardOnBattlefield(v, CARDS.SmotheringTithe) !== null,
      "Smothering Tithe on battlefield",
    );

    // Caster draws one card → trigger on the stack (mandatory, no
    // prompt yet — the question comes at resolution).
    await admin.sendActionAsPlayer(caster.playerID, "draw_card", {});
    await admin.waitFor(
      (v) => triggerOnStack(v, CARDS.SmotheringTithe) !== null,
      "Tithe trigger on the stack after the caster drew",
    );
    expect(admin.snapshot().pending_choices ?? []).toHaveLength(0);

    await resolveStack(setup);

    // Resolution queues the pay prompt for the CASTER.
    const prompted = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "pay_unless" && c.chooser === caster.playerID,
        ),
      "pay_unless prompt queued for the caster",
    );
    const prompt = (prompted.pending_choices ?? []).find((c) => c.kind === "pay_unless");
    expect(prompt?.pay_cost).toBe("{2}");

    // Caster's browser shows the pay dialog; the opponent's doesn't.
    await expect(caster.page.getByRole("button", { name: /^Pay \{2\}$/ })).toBeVisible({
      timeout: 5000,
    });
    await expect(opponent.page.getByRole("button", { name: /^Pay \{2\}$/ })).toHaveCount(0);

    await caster.page.getByRole("button", { name: /^Don't pay$/ }).click();

    const after = await admin.waitFor(
      (v) => {
        const t = findCardOnBattlefield(v, CARDS.Treasure);
        return t !== null && t.controller === opponent.playerID;
      },
      "Treasure created under the Tithe's controller",
    );
    expect(after.pending_choices ?? []).toHaveLength(0);
  });
});
