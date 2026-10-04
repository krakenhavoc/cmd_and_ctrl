import { expect, test, type Page } from "@playwright/test";
import { CARDS } from "./s19-deck-fixture";
import {
  adminMoveByName,
  findCardOnBattlefield,
  setupS19Game,
  triggerOnStack,
  type S19Setup,
} from "./s19-helpers";

// #2208, ADR 0120: a seat's board drawn larger over the table. Hovering
// the opponent's avatar peeks at it, moving away closes it, clicking the
// avatar pins it, a card in it can be picked as a target exactly as on
// the table, and Escape closes it.
//
// The two-player S19 fixture: the opponent's deck has no creature, so the
// pick is Reclamation Sage's trigger choosing the opponent's Sol Ring, a
// board click through the same PlayerPanel router a creature target uses.
//
// The overlay's name, "Opponent's board, expanded", is a label contract
// (ADR 0120 §3): it must not contain "Opponent board", which other specs
// match as a substring to find the table's panel.

const dockOf = (page: Page) =>
  page.getByRole("region", { name: "actions", exact: true });

test.describe("expanded board (#2208)", () => {
  test.describe.configure({ mode: "serial" });
  let setup: S19Setup | null = null;

  test.afterEach(async () => {
    if (setup) {
      await setup.shutdown();
      setup = null;
    }
  });

  test("hover peeks, a click pins, a pick goes through it, Escape closes it", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;
    const page = caster.page;

    await adminMoveByName(
      admin,
      opponent.playerID,
      CARDS.SolRing,
      "library",
      "battlefield",
    );
    await admin.waitFor(
      (v) => findCardOnBattlefield(v, CARDS.SolRing) !== null,
      "Sol Ring on battlefield",
    );

    const table = page.getByRole("region", {
      name: "Opponent board",
      exact: true,
    });
    const avatar = table.locator(`[data-seat-id="${opponent.playerID}"]`);
    const expanded = page.getByRole("region", {
      name: "Opponent's board, expanded",
    });
    // The table's region is still exactly one while the overlay is open.
    const tableRegions = page.getByRole("region", { name: "Opponent board" });

    // Somewhere off the board to rest the pointer: the page's top-left
    // corner is the site header, nowhere near an avatar or the overlay.
    const away = () => page.mouse.move(2, 2);

    // Hover peeks.
    await away();
    await avatar.hover();
    await expect(expanded).toBeVisible();
    await expect(tableRegions).toHaveCount(1);
    await expect(
      expanded.getByRole("button", { name: "Sol Ring", exact: true }),
    ).toBeVisible();

    // Moving away closes it.
    await away();
    await expect(expanded).toHaveCount(0);

    // A click pins it: it stays when the pointer leaves.
    await avatar.click();
    await expect(expanded).toBeVisible();
    await expect(
      expanded.getByRole("button", { name: "Pin Opponent's expanded board" }),
    ).toHaveAttribute("aria-pressed", "true");
    await away();
    await page.waitForTimeout(800);
    await expect(expanded).toBeVisible();

    // A target picked in the overlay. Reclamation Sage enters, the caster
    // says Yes, and the board enters targeting mode.
    await adminMoveByName(
      admin,
      caster.playerID,
      CARDS.ReclamationSage,
      "library",
      "battlefield",
    );
    const sage = dockOf(page).getByRole("dialog", {
      name: /Reclamation Sage —/i,
    });
    await expect(sage).toBeVisible({ timeout: 20_000 });
    await sage.getByRole("button", { name: /^Yes$/ }).click();
    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "pick_target" && c.chooser === caster.playerID,
        ),
      "pick_target prompt queued",
    );
    await expect(
      page.getByRole("dialog", { name: /Select target for Reclamation Sage/i }),
    ).toBeVisible({ timeout: 20_000 });
    await expect(expanded).toBeVisible();
    await expanded
      .getByRole("button", { name: "Sol Ring", exact: true })
      .click();
    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).length === 0 &&
        triggerOnStack(v, CARDS.ReclamationSage) !== null,
      "Reclamation Sage trigger on the stack, its target picked in the overlay",
    );

    // Escape closes it.
    await away();
    await page.keyboard.press("Escape");
    await expect(expanded).toHaveCount(0);
    await expect(tableRegions).toHaveCount(1);
  });
});
