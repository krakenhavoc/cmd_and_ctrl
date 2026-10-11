import { expect, test, type Page } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { ADMIN_TOKEN } from "./env";
import { handCardCount, unreachableHandCards } from "./hand-reach";

// tutorial-layout: the tutorial's coach card must not cost the board
// any height (ADR 0076 §2.3, #1081 follow-up).
//
// The coach docks bottom-left on the practice table, and the viewer's
// own panel keeps a cell for it at the left of its bottom row. That cell
// used to take the card's height too, and the bottom row grew to fit it:
// at 1280×800 the creature row lost 40 of its 128px. Now the cell takes
// width from the hand row only, so the battlefield rows are the same
// height with the coach up as with it gone. This measures that on a real
// practice table at three desktop sizes.
//
// The width it takes is the hand's: the fan fits what is left of the
// row, so every card in the hand can still be reached by the pointer,
// with the coach up and with it gone (#2395: at 1440×900 the fan ran
// under the coach card and its leftmost card could not be clicked).

type Box = {
  top: number;
  bottom: number;
  left: number;
  right: number;
  height: number;
};

async function boxes(page: Page) {
  return page.evaluate(() => {
    const me = document.querySelector('[aria-label="your board"]');
    const r = (e: Element | null | undefined): Box | null => {
      if (!e) return null;
      const b = e.getBoundingClientRect();
      return {
        top: b.top,
        bottom: b.bottom,
        left: b.left,
        right: b.right,
        height: b.height,
      };
    };
    return {
      creatures: r(me?.querySelector(".grid-creatures")),
      middle: r(me?.querySelector(".grid-middle")),
      bottom: r(me?.querySelector(".grid-bottom")),
      handZone: r(me?.querySelector(".hand-zone")),
      coach: r(document.querySelector('[aria-label="tutorial coach"]')),
    };
  });
}

for (const size of [
  { width: 1280, height: 800 },
  { width: 1440, height: 900 },
  { width: 1920, height: 1080 },
]) {
  test(`the coach card takes no height from the board at ${size.width}×${size.height}`, async ({
    browser,
  }) => {
    test.slow();
    const context = await browser.newContext({ viewport: size });
    const page = await context.newPage();
    try {
      await page.goto("/");
      await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
      // The token form lives on #/admin only (ADR 0112 §2 item 8);
      // #/login is the signed-out visitor page.
      await page.goto("/#/admin");
      await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
      await page.getByRole("button", { name: "log in" }).click();
      await expect(page).toHaveURL(/#\/lobby$/);

      await page.goto("/#/practice");
      await expect(page).toHaveURL(/#\/games\//, { timeout: 20_000 });
      // The practice table opens the opening roll (ADR 0125 §5.2).
      // Press Roll, again after a tie, and take the first turn if this
      // page won; a bot that wins hands it to the player by itself.
      const keep = page.getByRole("button", { name: "Keep hand" });
      const roll = page
        .getByRole("dialog", { name: "roll for the first turn", exact: true })
        .getByRole("button", { name: "Roll", exact: true });
      const goFirst = page
        .getByRole("dialog", {
          name: "choose who takes the first turn",
          exact: true,
        })
        .getByRole("button", { name: "I go first", exact: true });
      await expect(roll).toBeVisible({ timeout: 15_000 });
      const deadline = Date.now() + 30_000;
      while (!(await keep.isVisible())) {
        if (Date.now() > deadline)
          throw new Error("the opening roll never reached the mulligan");
        if (await goFirst.isVisible()) await goFirst.click();
        else if ((await roll.isVisible()) && (await roll.isEnabled()))
          await roll.click();
        await page.waitForTimeout(200);
      }
      await keep.click();

      const coach = page.getByRole("complementary", { name: "tutorial coach" });
      await expect(coach).toBeVisible();
      // Let the dock and the coach publish their sizes.
      await expect(page.locator(".coach-spacer")).toHaveCount(1);
      await page.waitForTimeout(500);
      const up = await boxes(page);
      // #2395: every card of the opening hand shows, beside the coach.
      // The pointer is parked off the board, so the hand is at rest.
      await page.mouse.move(size.width / 2, 4);
      expect(await handCardCount(page)).toBeGreaterThan(0);
      expect(
        await unreachableHandCards(page),
        "hand cards the pointer cannot reach with the coach up",
      ).toEqual([]);

      await coach.getByRole("button", { name: "Skip tutorial" }).click();
      await expect(coach).toHaveCount(0);
      await expect(page.locator(".coach-spacer")).toHaveCount(0);
      await page.mouse.move(size.width / 2, 4);
      await page.waitForTimeout(500);
      const gone = await boxes(page);
      expect(
        await unreachableHandCards(page),
        "hand cards the pointer cannot reach with the coach gone",
      ).toEqual([]);

      // The battlefield rows and the bottom row are exactly as tall with
      // the coach as without it.
      for (const k of ["creatures", "middle", "bottom"] as const) {
        expect(up[k], k).not.toBeNull();
        expect(
          Math.abs(up[k]!.height - gone[k]!.height),
          `${k} height`,
        ).toBeLessThan(1);
        expect(Math.abs(up[k]!.top - gone[k]!.top), `${k} top`).toBeLessThan(1);
      }
      // The width came out of the hand row: the hand starts right of the card.
      expect(up.coach).not.toBeNull();
      expect(up.handZone!.left).toBeGreaterThanOrEqual(up.coach!.right);
      expect(gone.handZone!.left).toBeLessThan(up.handZone!.left);
    } finally {
      await context.close();
    }
  });
}

// #2990, #2409: nothing the tutorial anchors to sits under the coach card.
//
// The card is taller than the hand row, so it also rises over the left
// end of the lands row, where the first Forest lands and the corner
// piles sit. Steps 5 and 6 (the lands), 9 (a permanent, else a land) and
// 10 (the command zone) point at those; the card must not intersect any
// of them, nor the piles, with the coach up.

type Rect = { left: number; top: number; right: number; bottom: number };

const overlaps = (a: Rect, b: Rect): boolean =>
  a.left < b.right && b.left < a.right && a.top < b.bottom && b.top < a.bottom;

async function rectsOf(page: Page, selector: string) {
  return page.evaluate((sel) => {
    return [...document.querySelectorAll(sel)].map((e) => {
      const b = e.getBoundingClientRect();
      return { left: b.left, top: b.top, right: b.right, bottom: b.bottom };
    });
  }, selector);
}

const MINE = `[aria-label="${L.yourBoard}"]`;

for (const size of [
  { width: 1280, height: 720 },
  { width: 1366, height: 768 },
  { width: 1600, height: 900 },
]) {
  test(`the coach card covers no anchor of steps 5, 6, 9 and 10 at ${size.width}×${size.height}`, async ({
    browser,
  }) => {
    test.slow();
    const context = await browser.newContext({ viewport: size });
    const page = await context.newPage();
    try {
      await page.goto("/");
      await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
      await page.goto("/#/admin");
      await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
      await page.getByRole("button", { name: "log in" }).click();
      await expect(page).toHaveURL(/#\/lobby$/);
      await page.goto("/#/practice");
      await expect(page).toHaveURL(/#\/games\//, { timeout: 20_000 });

      const keep = page.getByRole("button", { name: "Keep hand" });
      const mulligan = page.getByRole("button", {
        name: "Mulligan",
        exact: true,
      });
      const roll = page
        .getByRole("dialog", { name: "roll for the first turn", exact: true })
        .getByRole("button", { name: "Roll", exact: true });
      const goFirst = page
        .getByRole("dialog", {
          name: "choose who takes the first turn",
          exact: true,
        })
        .getByRole("button", { name: "I go first", exact: true });
      await expect(roll).toBeVisible({ timeout: 15_000 });
      const deadline = Date.now() + 30_000;
      while (!(await keep.isVisible())) {
        if (Date.now() > deadline)
          throw new Error("the opening roll never reached the mulligan");
        if (await goFirst.isVisible()) await goFirst.click();
        else if ((await roll.isVisible()) && (await roll.isEnabled()))
          await roll.click();
        await page.waitForTimeout(200);
      }
      // A hand with a Forest, so there is a land to point at.
      const opening = page
        .getByRole("list", { name: "your opening hand" })
        .getByRole("listitem");
      const titles = () =>
        opening.evaluateAll((els) =>
          els.map((e) => e.getAttribute("title") ?? ""),
        );
      for (let i = 0; i < 6; i++) {
        const names = await titles();
        if (names.some((n) => n.startsWith("Forest"))) break;
        const before = names.join("|");
        await mulligan.click();
        await expect.poll(async () => (await titles()).join("|")).not.toBe(before);
      }
      await keep.click();

      const coach = page.getByRole("complementary", { name: "tutorial coach" });
      await expect(coach).toBeVisible();
      await expect(page.locator(".coach-spacer")).toHaveCount(1);
      await coach.getByRole("button", { name: "Start" }).click();

      // Play a Forest: it lands at the left end of the lands row.
      await page.mouse.move(size.width / 2, 4);
      const forest = page
        .getByLabel(L.yourHand, { exact: true })
        .getByRole("button", { name: /^Forest, .*playable land/ })
        .first();
      // Upkeep and draw come first: pass until the land is playable.
      const next = page
        .getByRole("region", { name: L.actions, exact: true })
        .getByRole("button", { name: L.next, exact: true });
      const until = Date.now() + 30_000;
      while (!(await forest.isVisible())) {
        if (Date.now() > until) break;
        if (await next.isEnabled().catch(() => false))
          await next.click().catch(() => undefined);
        await page.waitForTimeout(400);
      }
      await expect(forest).toBeVisible();
      await forest.evaluate((el) => (el as HTMLElement).click());
      await expect(
        page.locator(`${MINE} [aria-label="${L.lands}"] [data-instance-id]`).first(),
      ).toBeVisible();
      await page.mouse.move(size.width / 2, 4);
      await page.waitForTimeout(500);

      const check = async (what: string) => {
        const card = (await rectsOf(page, '[aria-label="tutorial coach"]'))[0];
        expect(card, "the coach card is up").toBeTruthy();
        const anchors: [string, Rect[]][] = [
          ["lands row", await rectsOf(page, `${MINE} [aria-label="${L.lands}"]`)],
          [
            "lands",
            await rectsOf(page, `${MINE} [aria-label="${L.lands}"] [data-instance-id]`),
          ],
          [
            "permanents",
            await rectsOf(page, `${MINE} [aria-label="${L.creatures}"] [data-instance-id]`),
          ],
          ["piles", await rectsOf(page, `${MINE} .pile-bar`)],
          ["command zone", await rectsOf(page, `${MINE} [aria-label*="command zone"]`)],
        ];
        for (const [name, rs] of anchors)
          for (const r of rs)
            expect(
              overlaps(card, r),
              `${what}: the coach card covers the ${name}`,
            ).toBe(false);
      };

      // Walk forward with Skip step to each step the issue names.
      const count = coach.locator(".coach-count");
      let checked = 0;
      for (let guard = 0; guard < 16; guard++) {
        const m = /^(\d+) \/ \d+$/.exec(
          ((await count.textContent()) ?? "").trim(),
        );
        const n = m ? Number(m[1]) : 0;
        if ([5, 6, 9, 10].includes(n)) {
          await check(`step ${n}`);
          checked++;
        }
        if (n >= 10) break;
        const skip = coach.getByRole("button", { name: "Skip step" });
        if (!(await skip.isVisible())) break;
        await skip.click();
        await page.waitForTimeout(300);
      }
      expect(checked, "steps checked").toBeGreaterThan(0);
    } finally {
      await context.close();
    }
  });
}
