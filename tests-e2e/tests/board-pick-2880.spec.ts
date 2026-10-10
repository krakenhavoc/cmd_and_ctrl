import { expect, test, type Page } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { buildDeck } from "./s19-deck-fixture";
import {
  adminMoveByName,
  findCardOnBattlefield,
  findCardInPlayerGraveyard,
  resolveStack,
  setupS19Game,
  triggerOnStack,
  type S19Setup,
} from "./s19-helpers";

// #2880: a pending choice's permanents are picked by clicking them on
// the battlefield, and the choice is confirmed in the action dock.
//
// Fleshbag Marauder's "each player sacrifices a creature" asks the
// caster to choose between two creatures. The sheet lists them; the test
// folds it down, clicks the Grizzly Bears on the caster's own board, and
// presses the dock's Sacrifice. The opponent's half of the edict (their
// one creature, the Marauder) is answered from the admin socket.
//
// Tablet size (1024 × 768): the owner never plays smaller.

const BEARS = "Grizzly Bears";
const ELVES = "Llanowar Elves";
const MARAUDER = "Fleshbag Marauder";

const dockOf = (page: Page) => page.getByRole("region", { name: "actions", exact: true });

test.describe("#2880 choose permanents on the board", () => {
  let setup: S19Setup | null = null;

  test.afterEach(async () => {
    if (setup) {
      await setup.shutdown();
      setup = null;
    }
  });

  test("sacrifice a creature by clicking it on the board, then Sacrifice in the dock", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request, {
      casterDeck: buildDeck([BEARS, ELVES], "Forest", 97),
      opponentDeck: buildDeck([MARAUDER], "Plains", 98),
      viewport: { width: 1024, height: 768 },
    });
    const { admin, caster, opponent } = setup;

    await adminMoveByName(admin, caster.playerID, BEARS, "library", "battlefield");
    await adminMoveByName(admin, caster.playerID, ELVES, "library", "battlefield");
    await admin.waitFor(
      (v) => !!findCardOnBattlefield(v, BEARS) && !!findCardOnBattlefield(v, ELVES),
      "the caster's two creatures on the battlefield",
    );
    await adminMoveByName(admin, opponent.playerID, MARAUDER, "library", "battlefield");
    await admin.waitFor(
      (v) => triggerOnStack(v, MARAUDER) !== null,
      "Fleshbag Marauder's trigger on the stack",
    );
    await resolveStack(setup);

    const queued = await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "sacrifice_choice" && c.chooser === caster.playerID,
        ),
      "the caster's sacrifice prompt",
    );
    const bears = findCardOnBattlefield(queued, BEARS)!;
    const elves = findCardOnBattlefield(queued, ELVES)!;

    // The opponent sacrifices their only creature, the Marauder, from
    // the admin socket.
    const theirs = (queued.pending_choices ?? []).find(
      (c) => c.kind === "sacrifice_choice" && c.chooser === opponent.playerID,
    );
    if (theirs) {
      const marauder = findCardOnBattlefield(queued, MARAUDER)!;
      await admin.sendActionAsPlayer(opponent.playerID, "resolve_choice", {
        choice_id: theirs.id,
        card_ids: [marauder.instance_id],
      });
    }

    // The sheet opens in the dock, with Sacrifice waiting on a pick.
    const page = caster.page;
    const sheet = dockOf(page).getByRole("dialog").filter({
      has: page.getByRole("button", { name: L.sacrifice, exact: true }),
    });
    await expect(sheet).toBeVisible({ timeout: 20_000 });
    const sacrifice = sheet.getByRole("button", { name: L.sacrifice, exact: true });
    await expect(sacrifice).toBeDisabled();

    // Fold the sheet down: the board is in full view, and the dock's
    // Sacrifice stays in the bar.
    await sheet.getByRole("button", { name: "minimise" }).click();
    await expect(sheet.getByRole("button", { name: /^restore: /i })).toBeVisible();

    // The two creatures are highlighted on the caster's own board.
    const board = page.getByRole("region", { name: L.yourBoard, exact: true });
    const bearsTile = board.locator(`.card[data-instance-id="${bears.instance_id}"]`).first();
    const elvesTile = board.locator(`.card[data-instance-id="${elves.instance_id}"]`).first();
    await expect(bearsTile).toHaveClass(/\btargetable\b/);
    await expect(elvesTile).toHaveClass(/\btargetable\b/);

    // Click the Bears: picked on the board, and Sacrifice turns on.
    await bearsTile.click();
    await expect(bearsTile).toHaveClass(/\bpicked\b/);
    await expect(sacrifice).toBeEnabled();

    // A second click puts it back; a third picks it again.
    await bearsTile.click();
    await expect(sacrifice).toBeDisabled();
    await bearsTile.click();
    await expect(sacrifice).toBeEnabled();

    // The sheet shows the same pick.
    await sheet.getByRole("button", { name: /^restore: /i }).click();
    await expect(sheet).toContainText("1 / 1 selected");

    await sacrifice.click();

    const after = await admin.waitFor(
      (v) =>
        !(v.pending_choices ?? []).some(
          (c) => c.kind === "sacrifice_choice" && c.chooser === caster.playerID,
        ),
      "the caster's sacrifice answered",
    );
    expect(findCardOnBattlefield(after, BEARS)).toBeNull();
    expect(findCardOnBattlefield(after, ELVES)).not.toBeNull();
    expect(findCardInPlayerGraveyard(after, caster.playerID, BEARS)).not.toBeNull();
  });
});
