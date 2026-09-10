import { expect, test } from "@playwright/test";
import { CARDS } from "./s19-deck-fixture";
import {
  adminMoveByName,
  findCardInPlayerGraveyard,
  setupS19Game,
  type S19Setup,
} from "./s19-helpers";

// zone-browser: the graveyard / exile / command-zone modal behind the
// pile chips. It is the only way to look at a pile in the redesigned
// table, and the S19 trigger suite leans on it to answer a
// "which card?" prompt — but nothing covered the browser itself, so a
// regression in it would surface as a confusing trigger failure five
// files away.
//
// Reuses the S19 harness: it is the one fixture that can put a named
// card into a named zone. Deliberately read-only — no move_card from
// the browser — so the test asserts UI behaviour and never depends on
// the engine's zone-move gates.

test.describe("zone browser", () => {
  test.describe.configure({ mode: "serial" });
  let setup: S19Setup | null = null;

  test.afterEach(async () => {
    if (setup) {
      await setup.shutdown();
      setup = null;
    }
  });

  test("own graveyard: header count, search filter, owner actions, Escape closes", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster } = setup;

    await adminMoveByName(admin, caster.playerID, CARDS.LightningBolt, "library", "graveyard");
    await adminMoveByName(admin, caster.playerID, CARDS.GloriousAnthem, "library", "graveyard");
    await admin.waitFor(
      (v) =>
        findCardInPlayerGraveyard(v, caster.playerID, CARDS.LightningBolt) !== null &&
        findCardInPlayerGraveyard(v, caster.playerID, CARDS.GloriousAnthem) !== null,
      "both cards in the caster's graveyard",
    );

    const board = caster.page.getByRole("region", { name: "your board" });
    const graveChip = board.getByRole("button", { name: /^grave: / });
    // The pile chip carries the count in its accessible name — that
    // is the label a screen reader and this test both read.
    await expect(graveChip).toHaveAccessibleName("grave: 2 cards");
    await graveChip.click();

    const dialog = caster.page.getByRole("dialog", { name: /graveyard/i });
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText("(2)");

    // Both cards render. A card in a browsable zone is a role=img
    // tile unless the modal is in target-picking mode, in which case
    // it becomes a button — see the S19 Eternal Witness test.
    await expect(dialog.getByRole("img", { name: CARDS.LightningBolt, exact: true })).toBeVisible();
    await expect(dialog.getByRole("img", { name: CARDS.GloriousAnthem, exact: true })).toBeVisible();

    // The owner gets the move cluster on every card.
    await expect(dialog.getByRole("button", { name: /^move .* to hand$/ })).toHaveCount(2);

    // Search filters by name.
    await dialog.getByLabel("search cards").fill("lightning");
    await expect(dialog.getByRole("img", { name: CARDS.LightningBolt, exact: true })).toBeVisible();
    await expect(dialog.getByRole("img", { name: CARDS.GloriousAnthem, exact: true })).toHaveCount(
      0,
    );
    await dialog.getByLabel("search cards").fill("nothing matches this");
    await expect(dialog).toContainText("Nothing matches");
    await dialog.getByLabel("search cards").fill("");
    await expect(dialog.getByRole("img", { name: CARDS.GloriousAnthem, exact: true })).toBeVisible();

    // Tabs swap zone without closing the modal.
    await dialog.getByRole("tab", { name: "exile" }).click();
    const exileDialog = caster.page.getByRole("dialog", { name: /exile/i });
    await expect(exileDialog).toBeVisible();
    await expect(exileDialog).toContainText("No cards in this zone.");

    await caster.page.keyboard.press("Escape");
    await expect(caster.page.getByRole("dialog", { name: /exile/i })).toHaveCount(0);
  });

  test("an opponent's graveyard is browsable but read-only", async ({ browser, request }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;

    await adminMoveByName(admin, opponent.playerID, CARDS.SolRing, "library", "graveyard");
    await admin.waitFor(
      (v) => findCardInPlayerGraveyard(v, opponent.playerID, CARDS.SolRing) !== null,
      "Sol Ring in the opponent's graveyard",
    );

    // Graveyards are public information — the caster can open the
    // opponent's pile from the opponent's own panel.
    await caster.page
      .getByRole("region", { name: "Opponent board" })
      .getByRole("button", { name: /^grave: / })
      .click();

    const dialog = caster.page.getByRole("dialog", { name: /graveyard/i });
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText("Opponent");
    await expect(dialog.getByRole("img", { name: CARDS.SolRing, exact: true })).toBeVisible();

    // …but not manageable: no move cluster, and the header says so.
    await expect(dialog.getByRole("button", { name: /^move / })).toHaveCount(0);
    await expect(dialog).toContainText("read only");

    await dialog.getByRole("button", { name: "close" }).click();
    await expect(caster.page.getByRole("dialog", { name: /graveyard/i })).toHaveCount(0);
  });
});
