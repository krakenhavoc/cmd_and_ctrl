import { expect, test, type Page, type Route } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import {
  overviewFixture,
  sharedFixture,
} from "../../client/src/lib/test/advancedStatsFixture";

// ADR 0217 slice 6 (#4880): the stats pages and the share link.
//
// The e2e stack has no Discord sign-in, so a signed-in person is a
// session in localStorage and the /me routes answer from the fixture;
// the shared view needs no session at all, and its route is stubbed the
// same way. What is under test is the client: the overview's sections and
// charts, their text equivalents, the filters, the opt-in while advanced
// stats are off, the share dialog, a revoked link, and the layout at
// 1440, 820 and 390 px in a dark and the light skin with no page scroll.

const TOKEN = "AbCdEfGhIjKlMnOpQrStUv";
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

test.describe("stats pages (ADR 0217 slice 6)", () => {
  test("a shared link opens signed out, read-only, at three widths and two skins", async ({
    page,
  }, info) => {
    let reads = 0;
    await page.route(`**/stats/shared/${TOKEN}*`, (route) => {
      reads++;
      return json(route, 200, sharedFixture());
    });
    await page.goto("/");
    await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
    await page.goto(`/#/stats/shared/${TOKEN}`);
    await expect(page).toHaveURL(new RegExp(`#/stats/shared/${TOKEN}$`));
    await expect(
      page.getByRole("heading", { level: 1, name: "Bea's stats" }),
    ).toBeVisible();
    const shared = page.getByTestId("shared-stats");
    for (const h of ["Results", "Play", "Mana", "Cards", "Colours", "Decks"]) {
      await expect(
        shared.getByRole("heading", { level: 2, name: h }),
      ).toBeVisible();
    }
    // Every chart is an image named by its caption, with a table behind it.
    const places = shared.getByRole("img", {
      name: /four-seat tables you finish 1st/,
    });
    await expect(places).toBeVisible();
    const firstTable = shared.getByText(L.showAsTable).first();
    await firstTable.click();
    await expect(shared.getByRole("table").first()).toBeVisible();
    // No card images signed out: they need a session.
    await expect(shared.locator(".card-grid img")).toHaveCount(0);
    await expect(shared).toContainText("Sol Ring");

    // The viewer can change the filters: one read each.
    const before = reads;
    await page
      .getByRole("group", { name: L.statsRangeFilter })
      .getByText(L.statsRange30, { exact: true })
      .click();
    await expect.poll(() => reads).toBeGreaterThanOrEqual(before + 1);

    for (const theme of [null, "light"] as const) {
      await setSkin(page, theme);
      for (const width of WIDTHS) {
        await page.setViewportSize({ width, height: 900 });
        await noPageScroll(page);
        await info.attach(`shared-${theme ?? "dark"}-${width}.png`, {
          body: await page.screenshot({ fullPage: true }),
          contentType: "image/png",
        });
      }
    }
  });

  test("a revoked link says so, not an error", async ({ page }) => {
    await page.route("**/stats/shared/**", (route) =>
      json(route, 404, {
        error: "this stats link does not exist or has been revoked",
      }),
    );
    await page.goto(`/#/stats/shared/${TOKEN}`);
    await expect(page.getByTestId("shared-stats-gone")).toContainText(
      "doesn't work",
    );
    await expect(page.getByRole("alert")).toHaveCount(0);
  });

  test("#/stats: the opt-in while off, then the overview, the filters and the share dialog", async ({
    page,
  }, info) => {
    let prefs = {
      available: true,
      advanced: false,
      choice: "default",
      default: false,
    } as Record<string, unknown>;
    const advancedReads: string[] = [];
    const shareCalls: string[] = [];
    await page.route(/\/me(\/[^?]*)?(\?.*)?$/, async (route) => {
      const url = new URL(route.request().url());
      const method = route.request().method();
      switch (url.pathname) {
        case "/me":
          return json(route, 200, {
            role: "player",
            user_id: "11111111-1111-4111-8111-111111111111",
            display_name: "Bea",
          });
        case "/me/stats":
          return json(route, 200, {
            ...sharedFixture().basic,
            recent: [],
            advanced: prefs.advanced,
          });
        case "/me/stats/prefs":
          if (method === "PUT") {
            const on =
              JSON.parse(route.request().postData() ?? "{}").advanced === true;
            prefs = {
              ...prefs,
              advanced: on,
              choice: on ? "on" : "off",
              set_by: "self",
            };
          }
          return json(route, 200, prefs);
        case "/me/stats/advanced":
          advancedReads.push(url.search);
          return prefs.advanced
            ? json(route, 200, overviewFixture())
            : json(route, 403, { error: "advanced stats are off for you" });
        case "/me/stats/share-link":
          shareCalls.push(method);
          if (method === "POST") {
            prefs = { ...prefs, share_link_at: Date.now() };
            return json(route, 201, {
              token: TOKEN,
              path: `#/stats/shared/${TOKEN}`,
              created_at: Date.now(),
            });
          }
          return route.fulfill({ status: 204 });
        default:
          return json(route, 404, { error: "not found" });
      }
    });
    // Before the app's first script runs: the session module reads it
    // once, at load.
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
    await page.goto("/#/stats");

    // Off: what advanced stats are and the switch, never an error.
    const optIn = page.getByTestId("advanced-stats-opt-in");
    await expect(optIn).toBeVisible();
    await expect(page.getByRole("alert")).toHaveCount(0);
    expect(advancedReads).toHaveLength(0);
    await optIn.getByRole("button", { name: L.advancedStatsOptIn }).click();

    // On: the overview, and the switch stays at the head of the page.
    const overview = page.getByTestId("advanced-overview");
    await expect(overview).toBeVisible();
    await expect(
      page.getByRole("checkbox", {
        name: new RegExp(`^${L.advancedStatsSwitch}`),
      }),
    ).toBeChecked();
    await expect(
      overview.getByRole("img", { name: /most common win is combat damage/ }),
    ).toBeVisible();

    await page
      .getByRole("group", { name: L.statsMakeupFilter })
      .getByText(L.statsMakeupPeople, { exact: true })
      .click();
    await expect.poll(() => advancedReads.at(-1)).toContain("makeup=people");

    for (const theme of [null, "light"] as const) {
      await setSkin(page, theme);
      for (const width of WIDTHS) {
        await page.setViewportSize({ width, height: 900 });
        await noPageScroll(page);
        await info.attach(`overview-${theme ?? "dark"}-${width}.png`, {
          body: await page.screenshot({ fullPage: true }),
          contentType: "image/png",
        });
      }
    }
    await setSkin(page, null);
    await page.setViewportSize({ width: 1440, height: 900 });

    // The share dialog: make a link, see it once, revoke it.
    await page.getByRole("button", { name: L.shareStats }).click();
    const dialog = page.getByRole("dialog", { name: L.shareStatsDialog });
    await expect(dialog).toContainText("never shows the other players");
    await dialog.getByRole("button", { name: L.makeShareLink }).click();
    await expect(dialog.getByTestId("share-link-url")).toHaveValue(
      new RegExp(`/#/stats/shared/${TOKEN}$`),
    );
    await info.attach("share-dialog-1440.png", {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    await dialog.getByRole("button", { name: L.revokeShareLink }).click();
    await expect(dialog.getByTestId("share-link-url")).toHaveCount(0);
    expect(shareCalls).toEqual(["POST", "DELETE"]);
    await page.keyboard.press("Escape");
    await expect(dialog).toHaveCount(0);
  });
});
