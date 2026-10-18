import { expect, test } from "@playwright/test";
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
    // #3264: a second legal target, so the Sage's pick below is asked
    // rather than answered for the caster.
    await adminMoveByName(
      admin,
      opponent.playerID,
      CARDS.GloriousAnthem,
      "library",
      "battlefield",
    );
    await admin.waitFor(
      (v) =>
        findCardOnBattlefield(v, CARDS.SolRing) !== null &&
        findCardOnBattlefield(v, CARDS.GloriousAnthem) !== null,
      "Sol Ring and Glorious Anthem on battlefield",
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

    // A target picked in the overlay. Reclamation Sage enters and the
    // board enters targeting mode: the trigger's target is chosen as it
    // goes on the stack, and its "you may" waits for it to resolve
    // (CR 603.5, ADR 0197).
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

  // ADR 0120 as amended 2026-10-18 (#4122): the overlay is a reading
  // surface. At each screen size the header names the seat, the
  // battlefield groups draw at one card size, nothing in it scrolls
  // sideways, and the hand stays in its strip. The opponent's Sol Ring
  // (artifacts) and Glorious Anthem (enchantments) are two back-row
  // groups that a table panel draws a step smaller than its creatures.
  test("lays the board out evenly, under a header naming the seat (#4122)", async ({
    browser,
    request,
  }, testInfo) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;
    const page = caster.page;

    for (const name of [CARDS.SolRing, CARDS.GloriousAnthem]) {
      await adminMoveByName(
        admin,
        opponent.playerID,
        name,
        "library",
        "battlefield",
      );
    }
    await admin.waitFor(
      (v) =>
        findCardOnBattlefield(v, CARDS.SolRing) !== null &&
        findCardOnBattlefield(v, CARDS.GloriousAnthem) !== null,
      "Sol Ring and Glorious Anthem on battlefield",
    );

    const table = page.getByRole("region", {
      name: "Opponent board",
      exact: true,
    });
    const avatar = table.locator(`[data-seat-id="${opponent.playerID}"]`);
    const expanded = page.getByRole("region", {
      name: "Opponent's board, expanded",
    });

    for (const [width, height] of [
      [1024, 768],
      [1440, 900],
      [1920, 1080],
    ]) {
      await page.setViewportSize({ width, height });
      await page.mouse.move(2, 2);
      await avatar.click();
      await expect(expanded).toBeVisible();
      await expect(expanded.locator(".head")).toContainText("Opponent");
      await expect(
        expanded.getByRole("button", { name: "Sol Ring", exact: true }),
      ).toBeVisible();
      await page.mouse.move(2, 2);
      // Let the rows fit and the fade finish.
      await page.waitForTimeout(600);

      const m = await expanded.evaluate((o) => {
        // Upright box heights: a tapped card's turn is a transform, so
        // offsetHeight is the same for it as for an untapped one.
        const heights = [
          ...o.querySelectorAll<HTMLElement>(
            ".grid-creatures [data-instance-id], .grid-middle [data-instance-id]",
          ),
        ].map((e) => e.offsetHeight);
        const scrolls = [...o.querySelectorAll<HTMLElement>("*")]
          .filter((e) => {
            const x = getComputedStyle(e).overflowX;
            return (
              (x === "auto" || x === "scroll") &&
              e.scrollWidth > e.clientWidth + 1
            );
          })
          .map((e) => e.className);
        const strip = o
          .querySelector(".grid-bottom")!
          .getBoundingClientRect().top;
        const handTops = [...o.querySelectorAll(".hand-zone .hand-slot")].map(
          (e) => e.getBoundingClientRect().top,
        );
        // Every battlefield tile takes its own clicks: nothing from the
        // strip (the hand's fan) is the top element over any of it.
        const covered = [
          ...o.querySelectorAll<HTMLElement>(
            ".grid-creatures [data-instance-id], .grid-middle [data-instance-id]",
          ),
        ]
          .filter((e) => {
            const r = e.getBoundingClientRect();
            return [0.5, 0.9].some((f) =>
              document
                .elementFromPoint(r.left + r.width / 2, r.top + r.height * f)
                ?.closest(".grid-bottom"),
            );
          })
          .map((e) => e.getAttribute("aria-label"));
        return { heights, scrolls, strip, handTops, covered };
      });
      expect(m.covered, `${width}×${height}`).toEqual([]);
      const label = `${width}×${height}`;
      expect(m.heights.length, label).toBeGreaterThanOrEqual(2);
      expect(
        Math.max(...m.heights) - Math.min(...m.heights),
        label,
      ).toBeLessThanOrEqual(1);
      expect(m.scrolls, label).toEqual([]);
      for (const top of m.handTops)
        expect(top, label).toBeGreaterThanOrEqual(m.strip - 1);

      await testInfo.attach(`expanded-${width}x${height}.png`, {
        body: await page.screenshot(),
        contentType: "image/png",
      });

      // Unpin, which closes it.
      await avatar.click();
      await expect(expanded).toHaveCount(0);
    }
  });
});
