import { expect, test, type Locator, type Page } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { ADMIN_TOKEN } from "./env";

// #2396: a hovered card stays hovered wherever the pointer rests on it.
//
// Cards lift on hover. When the lift moved the hovered box itself, a
// pointer resting near the card's bottom edge (or, on a tilted end card
// of the hand, near its outer edge) was no longer on it once it rose:
// the card lost the hover, dropped back under the pointer and rose
// again, over and over, and the hover zoom flickered with it. Now the
// box the pointer is on stays put and an inner element rises (the
// hand, the castable strip and the opening-hand stage), or the card
// grows from its own bottom edge (every Card).
//
// On a practice table, at two desktop sizes, this rests the pointer on
// the points that used to flicker and checks the hover holds: the
// opening-hand stage's cards (bottom edge and both sides), your
// commander beside the hand (bottom and top of what shows), the bot's
// commander, every hand card's bottom edge and the hand's two outer
// edges. A resting hand trembles, and a browser only re-reads :hover on
// a pointer move, so the pointer moves a third of a pixel and back
// between samples.

type Pt = { x: number; y: number };
type Where = "bottom" | "top" | "left" | "right";

/** A point of `el` that shows (is `el` or inside it, not under something else). */
async function pointOf(el: Locator, where: Where): Promise<Pt> {
  const pt = await el.evaluate((node, w) => {
    const r = node.getBoundingClientRect();
    const top = Math.max(r.top, 0);
    const bottom = Math.min(r.bottom, innerHeight) - 1;
    const on = (x: number, y: number) => {
      const hit = document.elementFromPoint(x, y);
      return !!hit && node.contains(hit);
    };
    if (w === "bottom" || w === "top") {
      for (let k = 0; k < bottom - top; k++) {
        const y = w === "bottom" ? bottom - k : top + 1 + k;
        for (const fx of [0.3, 0.2, 0.4, 0.1, 0.5]) {
          if (on(r.left + r.width * fx, y)) return { x: r.left + r.width * fx, y };
        }
      }
      return null;
    }
    // The outermost point that shows, on any row.
    let best: { x: number; y: number } | null = null;
    for (let y = top + 2; y < bottom; y += 2) {
      for (let k = 0; k < r.width; k++) {
        const x = w === "left" ? r.left + k : r.right - k;
        if (on(x, y)) {
          const p = { x: w === "left" ? x + 1 : x - 1, y };
          if (!best || (w === "left" ? p.x < best.x : p.x > best.x)) best = p;
          break;
        }
      }
    }
    return best;
  }, where);
  if (!pt) throw new Error(`no point of ${el} shows (${where})`);
  return pt;
}

class Rest {
  constructor(private readonly page: Page) {}

  private async park(): Promise<void> {
    const size = this.page.viewportSize()!;
    await this.page.mouse.move(size.width / 2, 4);
    // Long enough for a lift's 220ms to settle back.
    await this.page.waitForTimeout(350);
  }

  /**
   * Rest on `el` at `where` (found at rest, with the pointer parked)
   * and check that `held` (default `el`) is hovered throughout.
   */
  async check(what: string, el: Locator, where: Where, held: Locator = el): Promise<void> {
    await this.park();
    const pt = await pointOf(el, where);
    await this.page.mouse.move(pt.x, pt.y);
    // Let the lift play out before sampling.
    await this.page.waitForTimeout(300);
    const states: boolean[] = [];
    for (let i = 0; i < 12; i++) {
      await this.page.mouse.move(pt.x + (i % 2 ? 0.3 : 0), pt.y);
      await this.page.waitForTimeout(40);
      states.push(await held.evaluate((n) => n.matches(":hover")));
    }
    expect(
      states.every(Boolean),
      `${what} (${where} at ${Math.round(pt.x)},${Math.round(pt.y)}): hovered ${states.map((s) => (s ? 1 : 0)).join("")}`,
    ).toBe(true);
  }
}

for (const size of [
  { width: 1280, height: 720 },
  { width: 1920, height: 1080 },
]) {
  test(`a hovered card stays hovered wherever the pointer rests, at ${size.width}×${size.height}`, async ({
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

      // No coach: the table as every other game draws it.
      const coach = page.getByRole("complementary", { name: "tutorial coach" });
      await coach.getByRole("button", { name: "Skip tutorial" }).click();
      await expect(coach).toHaveCount(0);

      const keep = page.getByRole("button", { name: L.keepHand, exact: true });
      const roll = page
        .getByRole("dialog", { name: L.rollForFirstTurn, exact: true })
        .getByRole("button", { name: L.roll, exact: true });
      const goFirst = page
        .getByRole("dialog", { name: L.chooseFirstTurn, exact: true })
        .getByRole("button", { name: L.iGoFirst, exact: true });
      await expect(roll).toBeVisible({ timeout: 15_000 });
      const deadline = Date.now() + 30_000;
      while (!(await keep.isVisible())) {
        if (Date.now() > deadline) throw new Error("the opening roll never reached the mulligan");
        if (await goFirst.isVisible()) await goFirst.click();
        else if ((await roll.isVisible()) && (await roll.isEnabled())) await roll.click();
        await page.waitForTimeout(200);
      }

      const rest = new Rest(page);

      // The opening-hand stage: the two end cards and the middle one.
      const stage = page.getByRole("list", { name: "your opening hand" }).getByRole("listitem");
      await expect(stage).toHaveCount(7);
      for (const i of [0, 3, 6]) {
        for (const where of ["bottom", "left", "right"] as const) {
          await rest.check(`opening-hand card ${i}`, stage.nth(i), where);
        }
      }

      await keep.click();
      await expect(keep).toHaveCount(0, { timeout: 10_000 });

      // Your commander, beside the hand. The command zone's group is
      // what the tutorial's commander step reads for its rest.
      const me = page.getByRole("region", { name: L.yourBoard, exact: true });
      const zone = me.getByLabel(L.commandZone.any.stem).first();
      const commander = zone.locator("[data-instance-id]").first();
      await expect(commander).toBeVisible();
      for (const where of ["bottom", "top"] as const) {
        await rest.check("your commander", commander, where);
        await rest.check("your command zone", commander, where, zone);
      }

      // The bot's commander, in its CommandStrip.
      const theirs = page.locator(".command-strip [data-instance-id]").first();
      await expect(theirs).toBeVisible();
      await rest.check("the bot's commander", theirs, "bottom");

      // Every hand card's bottom edge, and the fan's two outer edges,
      // which a tilted end card moved off when the whole hand rose.
      const hand = page.getByLabel(L.yourHand, { exact: true });
      const cards = hand.locator(".hand-slot [data-instance-id]");
      const n = await cards.count();
      expect(n).toBeGreaterThan(1);
      for (let i = 0; i < n; i++) {
        await rest.check(`hand card ${i}`, cards.nth(i), "bottom");
      }
      await rest.check("the hand (first card's outer edge)", cards.first(), "left", hand);
      await rest.check("the hand (last card's outer edge)", cards.last(), "right", hand);
    } finally {
      await context.close();
    }
  });
}
