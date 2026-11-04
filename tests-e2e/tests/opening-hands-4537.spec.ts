import { expect, test, type Page } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { ADMIN_TOKEN } from "./env";

// #4537: on a practice table at 1280×720 the attention strip's "Opening
// hands" roll call covered the first letters of the opponent's name
// ("ractice Bot"). The strip ended 74px left of the board's centre, and
// the opponent across the table has its name line centred there, up to
// 240px wide. This opens the table at the two sizes from the issue, waits
// for the roll call, and checks it meets no seat's name.

type Box = { left: number; right: number; top: number; bottom: number };

const overlaps = (a: Box, b: Box): boolean =>
  a.left < b.right && b.left < a.right && a.top < b.bottom && b.top < a.bottom;

async function toOpeningHands(page: Page): Promise<void> {
  await page.goto("/");
  await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
  await page.goto("/#/admin");
  await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
  await page.getByRole("button", { name: "log in" }).click();
  await expect(page).toHaveURL(/#\/lobby$/);
  await page.goto("/#/practice");
  await expect(page).toHaveURL(/#\/games\//, { timeout: 20_000 });
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
  // The opening-hand sheet covers the table; the roll call is the board's.
  await page.getByRole("button", { name: "View table" }).click();
  await expect(page.getByLabel(L.openingHandDecisions)).toBeVisible();
}

for (const size of [
  { width: 1280, height: 720 },
  { width: 1024, height: 768 },
]) {
  test(`the opening-hands roll call covers no seat's name at ${size.width}×${size.height}`, async ({
    browser,
  }) => {
    test.slow();
    const context = await browser.newContext({ viewport: size });
    const page = await context.newPage();
    try {
      await toOpeningHands(page);
      await page.mouse.move(size.width / 2, size.height - 4);
      const { strip, names } = await page.evaluate((label) => {
        const box = (e: Element) => {
          const r = e.getBoundingClientRect();
          return { left: r.left, right: r.right, top: r.top, bottom: r.bottom };
        };
        return {
          strip: box(document.querySelector(`[aria-label="${label}"]`)!),
          names: [
            ...document.querySelectorAll(".panel .identity .name-line"),
          ].map((e) => ({
            text: (e.textContent ?? "").trim(),
            box: box(e),
          })),
        };
      }, L.openingHandDecisions);
      expect(names.length, "seat name lines on the table").toBeGreaterThan(1);
      for (const n of names) {
        expect(overlaps(strip, n.box), `the roll call covers "${n.text}"`).toBe(
          false,
        );
      }
    } finally {
      await context.close();
    }
  });
}
