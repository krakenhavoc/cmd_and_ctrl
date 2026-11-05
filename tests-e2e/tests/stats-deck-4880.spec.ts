import { expect, test, type Page, type Route } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import {
  deckFixture,
  overviewFixture,
} from "../../client/src/lib/test/advancedStatsFixture";

// ADR 0217 slice 6, part 2 (#4880): a deck's advanced stats page.
//
// As in stats-pages-4880.spec.ts the e2e stack has no Discord sign-in, so
// a signed-in person is a session in localStorage and the /me routes
// answer from the fixture; the shared deck page needs no session and its
// route is stubbed the same way. What is under test is the client: the
// page's sections, the mana curve over the land curve on one axis, the
// dead cards, the version picker, the way in from the overview, the opt-in
// while advanced stats are off, a shared deck page, and the layout at
// 1440, 820 and 390 px in a dark and the light skin with no page scroll.

const TOKEN = "AbCdEfGhIjKlMnOpQrStUv";
const DECK = deckFixture().deck.id;
const WIDTHS = [1440, 820, 390];

function json(route: Route, status: number, body: unknown): Promise<void> {
  return route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
  });
}

async function noPageScroll(page: Page): Promise<void> {
  const [scroll, client] = await page.evaluate(() => [
    document.documentElement.scrollWidth,
    document.documentElement.clientWidth,
  ]);
  expect(scroll, "the page scrolls sideways").toBeLessThanOrEqual(client);
}

async function setSkin(page: Page, theme: "light" | null): Promise<void> {
  await page.evaluate((t) => {
    if (t) document.documentElement.setAttribute("data-theme", t);
    else document.documentElement.removeAttribute("data-theme");
  }, theme);
}

async function signIn(page: Page): Promise<void> {
  await page.addInitScript(() =>
    localStorage.setItem(
      "cmdctrl.session",
      JSON.stringify({
        token: "e2e-fake",
        expiresAt: "2999-01-01T00:00:00Z",
        principal: {
          role: "identified",
          user_id: "11111111-1111-4111-8111-111111111111",
          display_name: "Bea",
        },
      }),
    ),
  );
}

async function shoot(page: Page, info: { attach: Function }, name: string) {
  for (const theme of [null, "light"] as const) {
    await setSkin(page, theme);
    for (const width of WIDTHS) {
      await page.setViewportSize({ width, height: 900 });
      await noPageScroll(page);
      await info.attach(`${name}-${theme ?? "dark"}-${width}.png`, {
        body: await page.screenshot({ fullPage: true }),
        contentType: "image/png",
      });
    }
  }
  await setSkin(page, null);
  await page.setViewportSize({ width: 1440, height: 900 });
}

test.describe("a deck's stats page (ADR 0217 slice 6, part 2)", () => {
  test("a shared deck page opens signed out, read-only, at three widths and two skins", async ({
    page,
  }, info) => {
    const reads: string[] = [];
    await page.route(`**/stats/shared/${TOKEN}/decks/**`, (route) => {
      reads.push(new URL(route.request().url()).search);
      return json(route, 200, { ...deckFixture(), display_name: "Bea" });
    });
    await page.goto("/");
    await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
    await page.goto(`/#/stats/shared/${TOKEN}/deck/${DECK}`);
    await expect(
      page.getByRole("heading", { level: 1, name: "Atraxa Superfriends" }),
    ).toBeVisible();
    await expect(page.getByText("Bea's deck stats")).toBeVisible();
    const deck = page.getByTestId("deck-stats");
    for (const h of ["Mana", "Cards", "Composition", "Play", "Commander"]) {
      await expect(
        deck.getByRole("heading", { level: 2, name: h }),
      ).toBeVisible();
    }
    // Read-only: no art, no card images (they need a session), a way back.
    await expect(deck.locator(".band")).toHaveCount(0);
    await expect(deck.locator(".card-grid img")).toHaveCount(0);
    await expect(deck).toContainText("Craterhoof Behemoth");
    await expect(page.getByTestId("deck-stats-back")).toHaveAttribute(
      "href",
      `#/stats/shared/${TOKEN}`,
    );

    // The viewer can change the filters: one read each.
    const before = reads.length;
    await page
      .getByRole("group", { name: L.statsRangeFilter })
      .getByText(L.statsRange30, { exact: true })
      .click();
    await expect.poll(() => reads.length).toBeGreaterThanOrEqual(before + 1);
    expect(reads.at(-1)).toContain("range=30d");

    await shoot(page, info, "shared-deck");
  });

  test("a revoked link, or a deck that is gone, says so, not an error", async ({
    page,
  }) => {
    await page.route(`**/stats/shared/${TOKEN}/decks/**`, (route) =>
      json(route, 404, { error: "this stats link does not exist" }),
    );
    await page.goto(`/#/stats/shared/${TOKEN}/deck/${DECK}`);
    await expect(page.getByTestId("deck-stats-gone")).toContainText(
      "doesn't work",
    );
    await expect(page.getByRole("alert")).toHaveCount(0);
  });

  test("the overview opens the deck page, which reads versions and filters", async ({
    page,
  }, info) => {
    const deckReads: string[] = [];
    let prefs = {
      available: true,
      advanced: true,
      choice: "on",
      set_by: "self",
      default: false,
    } as Record<string, unknown>;
    await page.route(/\/me(\/[^?]*)?(\?.*)?$/, async (route) => {
      const url = new URL(route.request().url());
      const method = route.request().method();
      if (url.pathname === `/me/stats/decks/${DECK}`) {
        deckReads.push(url.search);
        if (!prefs.advanced)
          return json(route, 403, { error: "advanced stats are off for you" });
        const d = deckFixture();
        const version = url.searchParams.get("version");
        if (version === "all") d.version = "all";
        return json(route, 200, d);
      }
      switch (url.pathname) {
        case "/me":
          return json(route, 200, {
            role: "player",
            user_id: "11111111-1111-4111-8111-111111111111",
            display_name: "Bea",
          });
        case "/me/stats":
          return json(route, 200, {
            games: 14,
            wins: 5,
            losses: 9,
            draws: 0,
            avg_place: 2.2,
            unfinished: 0,
            by_deck: [],
            by_commander: [],
            top_cards: [],
            recent: [],
            advanced: true,
          });
        case "/me/stats/prefs":
          if (method === "PUT") {
            prefs = { ...prefs, advanced: true, choice: "on" };
          }
          return json(route, 200, prefs);
        case "/me/stats/advanced":
          return json(route, 200, overviewFixture());
        default:
          return json(route, 404, { error: "not found" });
      }
    });
    // Card images need the real server and a Scryfall index; a 1x1 PNG
    // stands in so the art band and the card grids have something to draw.
    await page.route("**/cards/*/image*", (route) =>
      route.fulfill({
        status: 200,
        contentType: "image/png",
        body: Buffer.from(
          "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR4nGNgYGD4DwABBAEAwS2OUAAAAABJRU5ErkJggg==",
          "base64",
        ),
      }),
    );
    await signIn(page);
    await page.goto("/#/stats");
    // The Decks row names the deck as a link to its page.
    await page
      .getByTestId("stats-decks")
      .getByRole("link", { name: "Atraxa Superfriends" })
      .click();
    await expect(page).toHaveURL(new RegExp(`#/stats/deck/${DECK}$`));
    await expect(
      page.getByRole("heading", { level: 1, name: "Atraxa Superfriends" }),
    ).toBeVisible();
    expect(deckReads.at(-1)).toContain("version=current");

    // Mana: the curve and the land curve are one image on one axis, and the
    // same numbers are a table.
    const mana = page.getByTestId("deck-mana");
    await expect(
      mana.getByRole("img", { name: /land drop 4 by turn 4/ }),
    ).toBeVisible();
    await mana.getByText(L.showAsTable).first().click();
    await expect(mana.getByRole("table").first()).toContainText("7+");

    // Cards: dead cards with their afford tag in words.
    const cards = page.getByTestId("deck-cards");
    await expect(cards).toContainText("Dead in 4 of 5 games in hand (80%)");
    await expect(cards).toContainText("Couldn't afford");
    await expect(cards).toContainText("Could afford");
    // The commander's art leads the page.
    await expect(page.locator(".band img")).toHaveCount(1);
    await expect(cards.locator(".card-grid img").first()).toBeVisible();

    // The version picker lists the earlier list; choosing it is one read.
    const picker = page.getByLabel(L.statsVersionPicker);
    await expect(picker.locator("option")).toHaveCount(3);
    await picker.selectOption("0f0f0f0f0f0f0f0f");
    await expect
      .poll(() => deckReads.at(-1))
      .toContain("version=0f0f0f0f0f0f0f0f");
    await picker.selectOption("all");
    await expect.poll(() => deckReads.at(-1)).toContain("version=all");
    await expect(page.getByTestId("deck-version-note")).toContainText(
      "All time",
    );

    await page
      .getByRole("group", { name: L.statsMakeupFilter })
      .getByText(L.statsMakeupPeople, { exact: true })
      .click();
    await expect.poll(() => deckReads.at(-1)).toContain("makeup=people");

    await shoot(page, info, "deck");

    // The way back.
    await page.getByTestId("deck-stats-back").click();
    await expect(page).toHaveURL(/#\/stats$/);
  });

  test("while advanced stats are off the page offers the switch, never an error", async ({
    page,
  }) => {
    let on = false;
    await page.route(/\/me(\/[^?]*)?(\?.*)?$/, async (route) => {
      const url = new URL(route.request().url());
      switch (url.pathname) {
        case "/me":
          return json(route, 200, {
            role: "player",
            user_id: "11111111-1111-4111-8111-111111111111",
            display_name: "Bea",
          });
        case "/me/stats/prefs":
          if (route.request().method() === "PUT") on = true;
          return json(route, 200, {
            available: true,
            advanced: on,
            choice: on ? "on" : "off",
            default: false,
          });
        case `/me/stats/decks/${DECK}`:
          return on
            ? json(route, 200, deckFixture())
            : json(route, 403, { error: "advanced stats are off for you" });
        default:
          return json(route, 404, { error: "not found" });
      }
    });
    await signIn(page);
    await page.goto(`/#/stats/deck/${DECK}`);
    const optIn = page.getByTestId("advanced-stats-opt-in");
    await expect(optIn).toBeVisible();
    await expect(page.getByRole("alert")).toHaveCount(0);
    await optIn.getByRole("button", { name: L.advancedStatsOptIn }).click();
    await expect(page.getByTestId("deck-stats")).toBeVisible();
  });
});
