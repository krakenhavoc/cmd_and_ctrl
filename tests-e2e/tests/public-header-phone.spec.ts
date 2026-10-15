import { expect, test, type Page } from "@playwright/test";

// #3877: the site header is one row at phone width on the public pages.
//
// The render test (client/src/lib/siteHeaderPhone.render.test.ts) covers
// the menu button's behaviour in jsdom, which applies no CSS. This is the
// layout in a real browser: at 320, 360, 390 and 430px the header is a
// single row about 60px tall, the page does not scroll sideways, every
// header control is a 44px tap target, and Sign in and the menu (with the
// nav links) stay reachable. A desktop width keeps the links in the row.

const PHONES = [320, 360, 390, 430];
const PAGES = [
  { name: "front page", hash: "#/" },
  { name: "catalog", hash: "#/catalog" },
  { name: "roadmap", hash: "#/roadmap" },
  { name: "about", hash: "#/about" },
  { name: "guide", hash: "#/guide" },
  { name: "privacy", hash: "#/privacy" },
  { name: "terms", hash: "#/terms" },
];

async function open(page: Page, width: number, hash: string): Promise<void> {
  await page.setViewportSize({ width, height: 800 });
  await page.goto(`/${hash}`);
  await expect(page.locator(".site-header .row")).toBeVisible();
}

const noSideScroll = (page: Page) =>
  page.evaluate(
    () =>
      document.documentElement.scrollWidth <=
      document.documentElement.clientWidth,
  );

test.describe("site header at phone width", () => {
  for (const width of PHONES) {
    for (const { name, hash } of PAGES) {
      test(`${name} at ${width}px: one row, no sideways scroll, 44px targets`, async ({
        page,
      }) => {
        await open(page, width, hash);
        expect(await noSideScroll(page)).toBe(true);

        const box = await page.locator(".site-header").boundingBox();
        expect(box!.height).toBeLessThanOrEqual(64);

        // Every visible control sits on one line and is at least 44px each way.
        const controls = await page
          .locator(".site-header a, .site-header button")
          .evaluateAll((els) =>
            els
              .filter((e) => (e as HTMLElement).offsetParent !== null)
              .filter((e) => !e.classList.contains("skip-link"))
              .map((e) => {
                const r = e.getBoundingClientRect();
                return {
                  label:
                    e.getAttribute("aria-label") ?? e.textContent?.trim() ?? "",
                  top: r.top,
                  width: r.width,
                  height: r.height,
                  right: r.right,
                };
              }),
          );
        expect(controls.length).toBeGreaterThan(0);
        for (const c of controls) {
          expect(c.width, `${c.label} width`).toBeGreaterThanOrEqual(43.5);
          expect(c.height, `${c.label} height`).toBeGreaterThanOrEqual(43.5);
          expect(c.right, `${c.label} inside the viewport`).toBeLessThanOrEqual(
            width,
          );
        }
        const tops = new Set(controls.map((c) => Math.round(c.top)));
        expect(tops.size, "one row").toBe(1);
      });
    }
  }

  test("the menu opens the nav links, takes focus, and Escape closes it", async ({
    page,
  }) => {
    await open(page, 390, "#/roadmap");
    const menu = page.getByRole("button", { name: "Site menu" });
    const nav = page.getByRole("navigation", { name: "Site", exact: true });
    await expect(nav).toBeHidden();
    await expect(page.getByRole("link", { name: "Sign in" })).toBeVisible();

    await menu.click();
    await expect(menu).toHaveAttribute("aria-expanded", "true");
    await expect(nav).toBeVisible();
    for (const link of ["Decks", "Catalog", "Roadmap", "Guide", "About"]) {
      const box = await nav.getByRole("link", { name: link }).boundingBox();
      expect(box!.height, link).toBeGreaterThanOrEqual(44);
    }
    await expect(nav.getByRole("link", { name: "Decks" })).toBeFocused();
    expect(await noSideScroll(page)).toBe(true);

    await page.keyboard.press("Escape");
    await expect(nav).toBeHidden();
    await expect(menu).toBeFocused();

    await menu.click();
    await nav.getByRole("link", { name: "Catalog" }).click();
    await expect(page).toHaveURL(/#\/catalog/);
    await expect(nav).toBeHidden();
  });

  test("Help's menu stays on screen at 320px", async ({ page }) => {
    await open(page, 320, "#/");
    await page.getByRole("button", { name: "help" }).click();
    const box = await page.getByRole("menu", { name: "help" }).boundingBox();
    expect(box!.x).toBeGreaterThanOrEqual(0);
    expect(box!.x + box!.width).toBeLessThanOrEqual(320);
  });

  test("a desktop width keeps the links in the row and hides the menu button", async ({
    page,
  }) => {
    await open(page, 1440, "#/");
    await expect(page.getByRole("button", { name: "Site menu" })).toBeHidden();
    await expect(
      page
        .getByRole("navigation", { name: "Site", exact: true })
        .getByRole("link", { name: "Roadmap" }),
    ).toBeVisible();
    expect(await noSideScroll(page)).toBe(true);
  });
});
