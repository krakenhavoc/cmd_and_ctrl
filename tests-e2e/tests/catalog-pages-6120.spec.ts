import { expect, test, type Page } from "@playwright/test";
import { L } from "../../client/src/lib/labels";

// #6120: the card catalogue shows one page at a time. 50 cards a page by
// default, with 25, 50, 75, 100 and 250 on offer; previous, next and the
// page numbers; "Showing 51–100 of N"; any filter, the search or a new
// page size goes back to page 1; and the page and size live in the URL,
// so back, forward and a shared link work, and the roadmap's ?q= link
// still lands on its card.
//
// /catalog answers from this file, so the test does not depend on which
// cards the e2e server's dump holds. The paging rules themselves are unit
// tested in client/src/lib/catalogPaging.test.ts; this is the page.

const TOTAL = 130;
const pad = (n: number) => String(n).padStart(3, "0");
const name = (n: number) => `Test Card ${pad(n)}`;

const CATALOG = {
  counts: { total: TOTAL, full: TOTAL, caveats: 0, unreviewed: 0, missing: 0 },
  cards: Array.from({ length: TOTAL }, (_, i) => ({
    oracle_id: `00000000-0000-4000-8000-000000000${pad(i + 1)}`,
    name: name(i + 1),
    type_line: i % 2 === 0 ? "Creature — Bear" : "Instant",
    oracle_text: "",
    types: [i % 2 === 0 ? "creature" : "instant"],
    color_identity: ["G"],
    completeness: "full",
  })),
};

async function stub(page: Page): Promise<void> {
  // Tips off, as in lobby.spec: a tip's card can sit over a control.
  await page.context().addInitScript(() => {
    try {
      if (localStorage.getItem("cmdctrl.settings.v1") === null) {
        localStorage.setItem(
          "cmdctrl.settings.v1",
          JSON.stringify({ help: { tipsOff: true } }),
        );
      }
    } catch {
      // storage unavailable: fall back to the defaults
    }
  });
  await page.route(
    (url) => url.pathname === "/catalog",
    (route) =>
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(CATALOG),
      }),
  );
}

const tiles = (page: Page) => page.locator("ul.grid > li");
const pager = (page: Page) =>
  page.getByRole("navigation", { name: L.cataloguePager, exact: true });
const count = (page: Page) => page.locator("p.count");

test.describe("the catalogue's pages (#6120)", () => {
  test.beforeEach(async ({ page }) => {
    await stub(page);
  });

  test("50 to a page, with next, previous, page numbers and back", async ({
    page,
  }) => {
    await page.goto("/#/catalog");
    await expect(tiles(page)).toHaveCount(50);
    await expect(tiles(page).first()).toContainText(name(1));
    await expect(count(page)).toHaveText(/Showing\s+1–50\s+of\s+130\s+cards/);

    const nav = pager(page);
    await expect(
      nav.getByRole("link", { name: L.cataloguePage(1), exact: true }),
    ).toHaveAttribute("aria-current", "page");
    // No Previous link on page 1.
    await expect(
      nav.getByRole("link", { name: L.cataloguePrevPage, exact: true }),
    ).toHaveCount(0);

    await nav
      .getByRole("link", { name: L.catalogueNextPage, exact: true })
      .click();
    await expect(page).toHaveURL(/#\/catalog\?page=2$/);
    await expect(tiles(page).first()).toContainText(name(51));
    await expect(count(page)).toHaveText(/Showing\s+51–100\s+of\s+130\s+cards/);

    await nav
      .getByRole("link", { name: L.cataloguePage(3), exact: true })
      .click();
    await expect(page).toHaveURL(/#\/catalog\?page=3$/);
    await expect(tiles(page)).toHaveCount(30);
    await expect(tiles(page).first()).toContainText(name(101));
    await expect(
      nav.getByRole("link", { name: L.catalogueNextPage, exact: true }),
    ).toHaveCount(0);
    // The current page is marked for a screen reader.
    await expect(
      nav.getByRole("link", { name: L.cataloguePage(3), exact: true }),
    ).toHaveAttribute("aria-current", "page");

    await page.goBack();
    await expect(page).toHaveURL(/#\/catalog\?page=2$/);
    await expect(tiles(page).first()).toContainText(name(51));
    await page.goBack();
    await expect(tiles(page).first()).toContainText(name(1));
    await page.goForward();
    await expect(tiles(page).first()).toContainText(name(51));
  });

  test("the pager works from the keyboard", async ({ page }) => {
    await page.goto("/#/catalog");
    await expect(tiles(page)).toHaveCount(50);
    const next = pager(page).getByRole("link", {
      name: L.catalogueNextPage,
      exact: true,
    });
    await next.focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/page=2/);
    await expect(tiles(page).first()).toContainText(name(51));
    // Previous and Next are both there on a middle page, and focus stayed on Next.
    await expect(next).toBeFocused();
    await expect(
      pager(page).getByRole("link", { name: L.cataloguePrevPage, exact: true }),
    ).toBeVisible();
    // On the last page Next is gone, so focus moves to the current page's
    // number rather than dropping to the document.
    await page.keyboard.press("Enter");
    await expect(tiles(page).first()).toContainText(name(101));
    await expect(next).toHaveCount(0);
    await expect(
      pager(page).getByRole("link", { name: L.cataloguePage(3), exact: true }),
    ).toBeFocused();
  });

  test("the page size is chosen, kept in the URL, and resets to page 1", async ({
    page,
  }) => {
    await page.goto("/#/catalog?page=2");
    await expect(tiles(page).first()).toContainText(name(51));

    const size = page.getByRole("combobox", {
      name: L.cataloguePageSize,
      exact: true,
    });
    await expect(size).toHaveValue("50");
    await expect(size.locator("option")).toHaveText([
      "25",
      "50",
      "75",
      "100",
      "250",
    ]);
    await size.selectOption("25");
    await expect(page).toHaveURL(/#\/catalog\?size=25$/);
    await expect(tiles(page)).toHaveCount(25);
    await expect(tiles(page).first()).toContainText(name(1));

    await size.selectOption("250");
    await expect(tiles(page)).toHaveCount(TOTAL);
    // One page: no pager at all.
    await expect(pager(page)).toHaveCount(0);
    await expect(count(page)).toHaveText(/^\s*130\s+cards/);
  });

  test("a shared link opens its page, and a page past the end shows the last", async ({
    page,
  }) => {
    await page.goto("/#/catalog?page=3&size=25");
    await expect(tiles(page)).toHaveCount(25);
    await expect(tiles(page).first()).toContainText(name(51));
    await expect(
      page.getByRole("combobox", { name: L.cataloguePageSize }),
    ).toHaveValue("25");

    await page.goto("/#/catalog?page=99");
    await expect(tiles(page).first()).toContainText(name(101));
    await expect(page).toHaveURL(/#\/catalog\?page=3$/);
  });

  test("a search or a filter goes back to page 1", async ({ page }) => {
    await page.goto("/#/catalog?page=2");
    await expect(tiles(page).first()).toContainText(name(51));

    await page
      .getByRole("textbox", { name: L.searchCatalogue, exact: true })
      .fill("Test Card");
    await expect(tiles(page).first()).toContainText(name(1));
    await expect(page).toHaveURL(/#\/catalog\?q=Test\+Card$/);

    await pager(page)
      .getByRole("link", { name: L.catalogueNextPage, exact: true })
      .click();
    await expect(tiles(page).first()).toContainText(name(51));
    await page.getByRole("button", { name: /^instant/i }).click();
    // 65 instants: page 1 of 2, starting at the first instant.
    await expect(tiles(page).first()).toContainText(name(2));
    await expect(count(page)).toHaveText(/Showing\s+1–50\s+of\s+65\s+cards/);
    await expect(page).not.toHaveURL(/page=/);
  });

  test("the roadmap's ?q= link still lands on its card", async ({ page }) => {
    await page.goto(`/#/catalog?q=${encodeURIComponent(name(120))}`);
    await expect(tiles(page)).toHaveCount(1);
    await expect(tiles(page).first()).toContainText(name(120));
    await expect(pager(page)).toHaveCount(0);
  });
});
