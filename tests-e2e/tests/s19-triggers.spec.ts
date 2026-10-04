import { expect, test, type Page } from "@playwright/test";
import { CARDS } from "./s19-deck-fixture";
import {
  adminMoveByName,
  findCardInPlayerGraveyard,
  findCardInPlayerHand,
  findCardOnBattlefield,
  playerByID,
  resolveStack,
  returnToLibrary,
  seedHandWithCard,
  setupS19Game,
  triggerOnStack,
  type JoinedPlayer,
  type S19Setup,
} from "./s19-helpers";

// S19 trigger e2e suite. Each test spins up a fresh 2-player game
// (≈10s setup) and exercises one S19 trigger end-to-end through the
// real WS path: admin moves a card onto the battlefield, the trigger
// harvester fires, the chooser's browser sees the prompt (or doesn't,
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
//
// ADR 0111 Delivery PR 5: a trigger's yes/no and a pay-unless are not a
// centred modal any more. They are answered inline in the action dock
// (region "actions"), inside one non-modal dialog that keeps the
// modal's name (the prompt's reason). The buttons keep their names
// (Yes, No, "Pay {2}", "Don't pay"), and are found inside the dock, so
// a copy anywhere else on the page would not be clicked by mistake.
// The "absent on the opponent's page" checks stay page-wide.
const dockOf = (page: Page) => page.getByRole("region", { name: "actions", exact: true });

// waitForPickTarget resolves once a pick_target prompt is queued for
// `chooser` (S20 sub-PR 2 — targeted triggers ask for their target on
// the board instead of auto-picking) AND that player's browser has
// actually entered targeting mode.
//
// Both halves matter. The admin snapshot is not the chooser's page:
// clicking a board card before the page has processed the delta is not
// a pick. With no targeting live, PlayerPanel's click rule (ADR 0117)
// does what the card does: nothing on a card the chooser doesn't
// control (unless it has an any-player ability, which opens its
// popover), and on their own card it may open the ability popover or
// tap it for mana. Either way the click is spent and the test then
// waits out the clock for a pick that never happened. Every
// board-click target pick has to gate on the
// targeting banner, which is the page's own proof it is ready.
//
// (Modal clicks don't need this: the button doesn't exist until the
// modal renders, so Playwright's own actionability wait covers it.
// A battlefield card is on screen the whole time, so there is
// nothing for it to wait on.)
async function waitForPickTarget(setup: S19Setup, chooser: JoinedPlayer, sourceName: string) {
  const view = await setup.admin.waitFor(
    (v) =>
      (v.pending_choices ?? []).some(
        (c) => c.kind === "pick_target" && c.chooser === chooser.playerID,
      ),
    "pick_target prompt queued",
  );
  // 20s, not the project's 10s default: three browser contexts, two
  // dev servers and four sockets share one self-hosted runner. A
  // player page was once seen 10-20s behind the admin socket, but that
  // was the hand seeding flooding it with snapshots of a 90-card hand
  // (#2253), which no longer happens. The wait is bounded well inside
  // the 90s test.slow() budget, so a genuinely stuck prompt still fails.
  await expect(
    chooser.page.getByRole("dialog", { name: new RegExp(`Select target for ${sourceName}`, "i") }),
  ).toBeVisible({ timeout: 20_000 });
  return view;
}

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
    // the pre-move hand size, so the baseline already counts it.
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
    //
    // 20s, not the project's 10s default (#1468): this assertion
    // reads the OPPONENT's own browser, whose WS broadcast is a hop
    // behind the admin snapshot we already confirmed the trigger
    // against above — the same cross-socket lag waitForPickTarget
    // documents higher up in this file, observed as a full 10s under
    // load on the shared self-hosted runner. Waiting longer, not
    // sleeping: this is still a polling `toBeVisible`, so a genuinely
    // stuck broadcast still fails, just past a bound that survives
    // load instead of racing it.
    await expect(caster.page.getByText(/Mulldrifter — draw two cards/i).first()).toBeVisible({
      timeout: 20_000,
    });
    await expect(
      opponent.page.getByText(/Mulldrifter — draw two cards/i).first(),
    ).toBeVisible({ timeout: 20_000 });

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
    // Anchor on the modal's accessible name (its <h2> is the
    // server's Question copy) rather than a bare page-wide text
    // scan: that also matched the stack overlay and the targeting
    // banner, and it raced the WS delta on a 5s budget.
    await expect(dockOf(caster.page).getByRole("dialog", { name: /Reclamation Sage —/i })).toBeVisible({ timeout: 20_000 });
    await expect(opponent.page.getByRole("dialog", { name: /Reclamation Sage —/i })).toHaveCount(
      0,
    );

    // Click "Yes" in the caster's dialog.
    await dockOf(caster.page)
      .getByRole("dialog", { name: /Reclamation Sage —/i })
      .getByRole("button", { name: /^Yes$/ })
      .click();

    // S20: "Yes" asks WHICH artifact or enchantment. The caster's
    // board enters targeting mode; the Sol Ring is a legal target
    // and gets clicked.
    await waitForPickTarget(setup, caster, CARDS.ReclamationSage);
    await caster.page
      .getByRole("region", { name: "Opponent board" })
      .getByRole("button", { name: "Sol Ring", exact: true })
      .click();

    // The pick puts the trigger on the stack; the Sol Ring survives
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

    await dockOf(caster.page).getByRole("button", { name: /^No$/ }).click();

    const after = await admin.waitFor(
      (v) => (v.pending_choices ?? []).length === 0,
      "prompt drained after No",
    );
    // Declined: no target prompt, nothing reaches the stack, Sol
    // Ring still on battlefield, not in graveyard.
    expect((after.pending_choices ?? []).some((c) => c.kind === "pick_target")).toBe(false);
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

    // S20: mandatory trigger → straight to the target pick (no
    // yes/no). Click the Sol Ring on the opponent's board.
    const picking = await waitForPickTarget(setup, caster, CARDS.AcidicSlime);
    expect((picking.pending_choices ?? []).some((c) => c.kind === "trigger_prompt")).toBe(false);
    await caster.page
      .getByRole("region", { name: "Opponent board" })
      .getByRole("button", { name: "Sol Ring", exact: true })
      .click();

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

    await expect(dockOf(caster.page).getByRole("dialog", { name: /Eternal Witness —/i })).toBeVisible({ timeout: 20_000 });

    await dockOf(caster.page)
      .getByRole("dialog", { name: /Eternal Witness —/i })
      .getByRole("button", { name: /^Yes$/ })
      .click();

    // S20: pick the card — open the caster's own graveyard browser
    // and click the Bolt.
    //
    // Two selectors here need care; both used to fail:
    //   * the targeting banner is ALSO role="dialog", so the zone
    //     browser has to be addressed by its own accessible name;
    //   * every card in a manageable zone renders three sibling
    //     "move <name> to <zone>" buttons, so an inexact name match
    //     resolves to four elements. exact:true picks the card.
    await waitForPickTarget(setup, caster, CARDS.EternalWitness);
    await caster.page
      .getByRole("region", { name: "your board" })
      .getByRole("button", { name: /^grave: / })
      .click();
    const graveBrowser = caster.page.getByRole("dialog", { name: /graveyard/i });
    await expect(graveBrowser).toBeVisible({ timeout: 20_000 });
    await graveBrowser.getByRole("button", { name: "Lightning Bolt", exact: true }).click();

    const staged = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).length === 0 &&
        triggerOnStack(v, CARDS.EternalWitness) !== null,
      "Eternal Witness trigger on the stack after Yes + pick",
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

    await dockOf(caster.page).getByRole("button", { name: /^No$/ }).click();

    const after = await admin.waitFor(
      (v) => (v.pending_choices ?? []).length === 0,
      "prompt drained after No",
    );
    // Lightning Bolt still in graveyard; nothing returned to hand.
    expect(findCardInPlayerGraveyard(after, caster.playerID, CARDS.LightningBolt)).not.toBeNull();
    expect(findCardInPlayerHand(after, caster.playerID, CARDS.LightningBolt)).toBeNull();
  });

  test("Solemn Simulacrum optional ETB → Yes, the caster picks the Island, it enters tapped", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster } = setup;

    const solemn = await seedHandWithCard(admin, caster.playerID, CARDS.SolemnSimulacrum);

    // The search below has to be able to OFFER the deck's single
    // Island, so it must be in the library, and the opening hand may
    // have drawn it. Putting it into the hand and back makes "exactly
    // one Island is in the library" a fact, not a hope about the
    // shuffle.
    await seedHandWithCard(admin, caster.playerID, CARDS.Island);
    await returnToLibrary(admin, caster.playerID, CARDS.Island, 1);
    const stocked = playerByID(admin.snapshot(), caster.playerID);
    expect(stocked.library.count).toBeGreaterThanOrEqual(4);
    expect(findCardInPlayerHand(admin.snapshot(), caster.playerID, CARDS.Island)).toBeNull();

    await admin.sendActionAsPlayer(caster.playerID, "move_card", {
      src: { kind: "hand", owner: caster.playerID },
      dst: { kind: "battlefield" },
      instance_id: solemn.instance_id,
    });

    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "Solemn prompt queued",
    );
    // Name the "you may" dialog exactly. Once the search opens there
    // are TWO dialogs whose accessible name starts "Solemn
    // Simulacrum —", and getByRole's name match is a substring by
    // default, so a loose pattern here would match either one.
    const solemnPrompt = dockOf(caster.page).getByRole("dialog", {
      name: "Solemn Simulacrum — search for a basic land?",
    });
    await expect(solemnPrompt).toBeVisible({ timeout: 20_000 });

    await solemnPrompt.getByRole("button", { name: /^Yes$/ }).click();

    const staged = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).length === 0 &&
        triggerOnStack(v, CARDS.SolemnSimulacrum) !== null,
      "Solemn trigger on the stack after Yes",
    );
    expect(findCardOnBattlefield(staged, CARDS.Forest)).toBeNull();
    expect(findCardOnBattlefield(staged, CARDS.Island)).toBeNull();

    await resolveStack(setup);

    // S22 (#272): answering "Yes" no longer finishes the job. The
    // trigger resolves into a SECOND prompt — the search chooser —
    // because the searcher, not the engine, decides which basic the
    // library gives up. Before #272 the engine took the first match
    // in library order; the test that predated it answered only the
    // "you may" half and then waited out the clock here.
    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "search_library" && c.chooser === caster.playerID,
        ),
      "search_library prompt queued for the caster",
    );
    // 20s for the same reason waitForPickTarget uses it: the admin
    // socket runs ahead of the player's page under runner load.
    // ADR 0111 PR 6: the search is a sheet in the action dock; its card
    // grid and its "Take" (the action bar's primary) share its dialog.
    const searchDialog = dockOf(caster.page).getByRole("dialog", {
      name: "Solemn Simulacrum — a basic land",
    });
    await expect(searchDialog).toBeVisible({ timeout: 20_000 });

    // Take the Island, not a Forest. Eighty-nine Forests and one Island
    // are on offer; picking the Island is what makes the assertion
    // below a statement about the CHOOSER rather than about deck
    // order. exact:true because "select Island" would otherwise
    // substring-match nothing useful today but is one card name away
    // from ambiguity.
    await searchDialog.getByRole("button", { name: "select Island", exact: true }).click();
    await searchDialog.getByRole("button", { name: /^Take$/ }).click();

    // The Island the caster picked — not the Forest the old
    // first-match-in-library-order engine would have taken — is on
    // the battlefield, tapped, under the caster's control.
    const after = await admin.waitFor(
      (v) => {
        const island = findCardOnBattlefield(v, CARDS.Island);
        return island !== null && island.controller === caster.playerID;
      },
      "the chosen Island on the battlefield under caster's control",
    );
    const island = findCardOnBattlefield(after, CARDS.Island);
    expect(island?.tapped).toBe(true);
    // The counter-assertion that carries the weight: no Forest came
    // along. A search that ignored the pick would have fetched one,
    // since Forests outnumber the Island eighty-nine to one.
    expect(findCardOnBattlefield(after, CARDS.Forest)).toBeNull();
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
    // Capture library count AFTER the move settles, BEFORE we
    // resolve the prompt: the decline path must not fetch anything
    // from the library.
    const promptView = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "trigger_prompt" && c.chooser === caster.playerID,
        ),
      "prompt queued",
    );
    const libraryAtPrompt = playerByID(promptView, caster.playerID).library.count;

    await dockOf(caster.page).getByRole("button", { name: /^No$/ }).click();

    const after = await admin.waitFor(
      (v) => (v.pending_choices ?? []).length === 0,
      "prompt drained",
    );
    const libraryAfter = playerByID(after, caster.playerID).library.count;
    // Decline doesn't fetch — the library count at decline time
    // should equal the library count when the prompt queued.
    expect(libraryAfter).toBe(libraryAtPrompt);
    // Neither basic reaches the battlefield, and no search chooser
    // ever opens — declining the "you may" ends the ability before
    // the search that would have asked.
    expect(findCardOnBattlefield(after, CARDS.Forest)).toBeNull();
    expect(findCardOnBattlefield(after, CARDS.Island)).toBeNull();
    expect((after.pending_choices ?? []).some((c) => c.kind === "search_library")).toBe(false);
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
    await expect(dockOf(caster.page).getByRole("button", { name: /^Yes$/ })).toBeVisible({ timeout: 20_000 });
    await expect(dockOf(caster.page).getByRole("button", { name: /^No$/ })).toBeVisible({ timeout: 20_000 });
    await expect(opponent.page.getByRole("dialog", { name: /Reclamation Sage —/i })).toHaveCount(
      0,
    );
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

    // The dock's prompt should render the same text — proves the
    // wire shape's `reason` field flows through to the dialog
    // without truncation or escape errors. The reason is the dock
    // request's dialog name (ADR 0111 PR 5), as it was the modal's.
    await expect(dockOf(caster.page).getByRole("dialog", { name: prompt!.reason! })).toBeVisible({ timeout: 20_000 });
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

    // Stage the Tithe before anything else: the only draw this test
    // makes is the one below that is meant to trigger it.
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
    await expect(dockOf(caster.page).getByRole("button", { name: /^Pay \{2\}$/ })).toBeVisible({ timeout: 20_000 });
    await expect(opponent.page.getByRole("button", { name: /^Pay \{2\}$/ })).toHaveCount(0);

    await dockOf(caster.page).getByRole("button", { name: /^Don't pay$/ }).click();

    const after = await admin.waitFor(
      (v) => {
        const t = findCardOnBattlefield(v, CARDS.Treasure);
        return t !== null && t.controller === opponent.playerID;
      },
      "Treasure created under the Tithe's controller",
    );
    expect(after.pending_choices ?? []).toHaveLength(0);
  });

  // ADR 0111 Delivery PR 6: a scry is a sheet that grows up out of the
  // action dock, not a centred modal. Its lanes are the sheet's body
  // and its Done is the action bar's primary, all inside one non-modal
  // dialog named by the prompt ("Scry 1"). It can be minimised to look
  // at the board and brought back without closing the prompt.
  test("Temple of Mystery's scry 1 is answered from the dock's sheet", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;

    const temple = await seedHandWithCard(admin, caster.playerID, CARDS.TempleOfMystery);
    await admin.sendActionAsPlayer(caster.playerID, "move_card", {
      src: { kind: "hand", owner: caster.playerID },
      dst: { kind: "battlefield" },
      instance_id: temple.instance_id,
    });
    await admin.waitFor(
      (v) => triggerOnStack(v, CARDS.TempleOfMystery) !== null,
      "Temple of Mystery's scry trigger on the stack",
    );
    await resolveStack(setup);
    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some((c) => c.kind === "scry" && c.chooser === caster.playerID),
      "scry prompt queued for the caster",
    );

    const scry = dockOf(caster.page).getByRole("dialog", { name: "Scry 1" });
    await expect(scry).toBeVisible({ timeout: 20_000 });
    await expect(scry).not.toHaveAttribute("aria-modal", "true");
    await expect(caster.page.locator(".prompt-backdrop")).toHaveCount(0);
    // Only the chooser is asked.
    await expect(opponent.page.getByRole("dialog", { name: "Scry 1" })).toHaveCount(0);

    // Minimise: the prompt stays open, folded to a restore chip.
    await scry.getByRole("button", { name: "minimise" }).click();
    const restore = scry.getByRole("button", { name: /^restore: /i });
    await expect(restore).toBeVisible();
    await expect(scry.getByRole("button", { name: /^Done$/ })).toBeVisible();
    await restore.click();

    // Put the one card on the bottom, then Done from the action bar.
    await scry.getByRole("button", { name: /on the bottom$/ }).click();
    await scry.getByRole("button", { name: /^Done$/ }).click();

    const after = await admin.waitFor(
      (v) => !(v.pending_choices ?? []).some((c) => c.kind === "scry"),
      "the scry is answered",
    );
    expect(findCardOnBattlefield(after, CARDS.TempleOfMystery)?.tapped).toBe(true);
    await expect(scry).toHaveCount(0, { timeout: 10_000 });
  });
});
