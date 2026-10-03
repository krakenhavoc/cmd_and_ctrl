import { expect, test, type Page } from "@playwright/test";
import { ADMIN_TOKEN } from "./env";

// tutorial-layout: the tutorial's coach card must not cost the board
// any height (ADR 0076 §2.3, #1081 follow-up).
//
// The coach docks bottom-left on the practice table, and the viewer's
// own panel keeps a cell for it at the left of its bottom row. That cell
// used to take the card's height too, and the bottom row grew to fit it:
// at 1280×800 the creature row lost 40 of its 128px. Now the cell takes
// width from the hand row only, so the battlefield rows are the same
// height with the coach up as with it gone. This measures that on a real
// practice table at two desktop sizes.

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
      const keep = page.getByRole("button", { name: "Keep hand" });
      await expect(keep).toBeVisible({ timeout: 15_000 });
      await keep.click();

      const coach = page.getByRole("complementary", { name: "tutorial coach" });
      await expect(coach).toBeVisible();
      // Let the dock and the coach publish their sizes.
      await expect(page.locator(".coach-spacer")).toHaveCount(1);
      await page.waitForTimeout(500);
      const up = await boxes(page);

      await coach.getByRole("button", { name: "Skip tutorial" }).click();
      await expect(coach).toHaveCount(0);
      await expect(page.locator(".coach-spacer")).toHaveCount(0);
      await page.waitForTimeout(500);
      const gone = await boxes(page);

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
