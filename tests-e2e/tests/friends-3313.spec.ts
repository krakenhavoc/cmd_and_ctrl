import { expect, test, type Page, type Route } from "@playwright/test";
import { L } from "../../client/src/lib/labels";

// ADR 0213 slice 3 (#3313): friends and blocks.
//
// The e2e stack has no Discord sign-in, so a signed-in person is a
// session in localStorage and the /me routes answer from this file, the
// way stats-pages-4880 does it. What is under test is the client: the
// header's Friends button and drawer (Requests, Friends, Played with),
// Accept and Add friend sending the right calls, the #/friends page's
// block confirm and a setting saved on change, History's With people and
// With friends, and no sideways scroll at 1440 and 820 px.

const ME = "11111111-1111-4111-8111-111111111111";
const BOB = "22222222-2222-4222-8222-222222222222";
const CARA = "33333333-3333-4333-8333-333333333333";
const EVE = "44444444-4444-4444-8444-444444444444";
const WIDTHS = [1440, 820];

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

interface Seen {
  method: string;
  path: string;
  body: unknown;
}

async function stub(page: Page): Promise<Seen[]> {
  const seen: Seen[] = [];
  const now = Date.now();
  await page.route(/\/me(\/[^?]*)?(\?.*)?$/, async (route) => {
    const req = route.request();
    const url = new URL(req.url());
    const method = req.method();
    const body = req.postData()
      ? JSON.parse(req.postData() ?? "{}")
      : undefined;
    seen.push({ method, path: url.pathname, body });
    const p = url.pathname;
    if (p === "/me")
      return json(route, 200, {
        role: "identified",
        user_id: ME,
        display_name: "Alice",
      });
    if (p === "/me/tablemates")
      return json(route, 200, {
        tablemates: [
          {
            user_id: BOB,
            display_name: "Bob",
            last_played_at: now,
            games_together: 3,
            last_table: "FNM",
          },
          {
            user_id: CARA,
            display_name: "Cara",
            last_played_at: now,
            friend: true,
          },
        ],
      });
    if (p === "/me/friends")
      return json(route, 200, {
        friends: [{ user_id: CARA, display_name: "Cara", since: now }],
        incoming: [{ user_id: EVE, display_name: "Eve", since: now }],
        outgoing: [],
        max_friends: 200,
        max_outgoing: 20,
      });
    if (p === "/me/friends/requests" && method === "POST")
      return json(route, 200, {
        outcome: "sent",
        user_id: BOB,
        display_name: "Bob",
      });
    if (p.endsWith("/accept")) return json(route, 200, { outcome: "friends" });
    if (p === "/me/blocks") return json(route, 200, { blocks: [] });
    if (p.startsWith("/me/blocks/")) return json(route, 200, { blocked: true });
    if (p === "/me/social")
      return json(route, 200, {
        presence: "friends",
        requests: method === "PUT" ? "nobody" : "played_with",
      });
    if (p === "/me/games")
      return json(route, 200, {
        games: [
          {
            id: "g-a",
            name: "With Cara",
            state: "ended",
            seat: 0,
            winner_seat: 1,
            created_at: now - 86_400_000,
            started_at: now - 86_400_000,
            ended_at: now - 86_400_000,
            archived_at: null,
            others: [{ seat: 1, name: "Cara", friend: true }],
          },
          {
            id: "g-b",
            name: "With Bob",
            state: "ended",
            seat: 0,
            winner_seat: 1,
            created_at: now - 2 * 86_400_000,
            started_at: now - 2 * 86_400_000,
            ended_at: now - 2 * 86_400_000,
            archived_at: null,
            others: [{ seat: 1, name: "Bob" }],
          },
        ],
      });
    // Anything else under /me (the session renewal, the last setup, the
    // deck library) answers empty: a real 401 to the fake token would
    // sign the page out.
    return json(route, 200, {});
  });
  await page.route(
    (url) => url.pathname === "/games",
    (route) =>
      route.request().method() === "GET"
        ? json(route, 200, { games: [] })
        : route.fallback(),
  );
  await page.addInitScript((user) => {
    localStorage.setItem(
      "cmdctrl.session",
      JSON.stringify({
        token: "e2e-fake",
        expiresAt: "2999-01-01T00:00:00Z",
        principal: { role: "identified", user_id: user, display_name: "Alice" },
      }),
    );
  }, ME);
  return seen;
}

test.describe("friends and blocks (ADR 0213 slice 3)", () => {
  test("the drawer: Requests, Friends, Played with; Accept and Add friend", async ({
    page,
  }, info) => {
    const seen = await stub(page);
    await page.goto("/#/lobby");
    await page
      .getByRole("button", { name: L.friendsButton, exact: true })
      .click();
    const drawer = page.getByRole("dialog", { name: L.friendsDrawer });
    await expect(drawer).toBeVisible();
    await expect(
      drawer.getByRole("heading", { name: /Requests/ }),
    ).toBeVisible();
    await expect(
      drawer.getByRole("heading", { name: L.playedWith }),
    ).toBeVisible();

    for (const width of WIDTHS) {
      await page.setViewportSize({ width, height: 900 });
      await noPageScroll(page);
      await info.attach(`drawer-${width}.png`, {
        body: await page.screenshot(),
        contentType: "image/png",
      });
    }
    await page.setViewportSize({ width: 1440, height: 900 });

    await drawer.getByRole("button", { name: L.addFriend }).click();
    await expect(drawer).toContainText("Friend request sent to Bob");
    expect(seen.find((s) => s.path === "/me/friends/requests")?.body).toEqual({
      user_id: BOB,
    });

    await drawer
      .getByRole("button", { name: L.acceptFriend, exact: true })
      .first()
      .click();
    await expect
      .poll(() =>
        seen.some((s) => s.path === `/me/friends/requests/${EVE}/accept`),
      )
      .toBe(true);

    await drawer.getByRole("link", { name: L.manageFriends }).click();
    await expect(page).toHaveURL(/#\/friends$/);
  });

  test("#/friends: a block asks first, a setting saves on change", async ({
    page,
  }, info) => {
    const seen = await stub(page);
    await page.goto("/#/friends");
    await expect(
      page.getByRole("heading", { level: 1, name: "Friends and blocks" }),
    ).toBeVisible();
    for (const width of WIDTHS) {
      await page.setViewportSize({ width, height: 900 });
      await noPageScroll(page);
      await info.attach(`friends-page-${width}.png`, {
        body: await page.screenshot({ fullPage: true }),
        contentType: "image/png",
      });
    }
    await page.setViewportSize({ width: 1440, height: 900 });

    const yours = page.locator("section", { has: page.locator("#f-friends") });
    await yours.getByRole("button", { name: L.blockPerson }).click();
    await expect(yours).toContainText(
      "It ends any friendship and request between you",
    );
    expect(
      seen.some((s) => s.method === "PUT" && s.path.startsWith("/me/blocks/")),
    ).toBe(false);
    await yours.getByRole("button", { name: L.blockPerson }).click();
    await expect
      .poll(() =>
        seen.some((s) => s.method === "PUT" && s.path === `/me/blocks/${CARA}`),
      )
      .toBe(true);

    await page
      .getByRole("radiogroup", { name: L.whoCanRequest })
      .getByRole("radio", { name: "Nobody" })
      .check();
    await expect
      .poll(
        () =>
          seen.find((s) => s.method === "PUT" && s.path === "/me/social")?.body,
      )
      .toEqual({ requests: "nobody" });
  });

  test("History: With people is anyone at the table, With friends a friend", async ({
    page,
  }) => {
    await stub(page);
    await page.goto("/#/lobby?show=history");
    const history = page.locator("#history");
    await history
      .getByRole("button", { name: L.historyWithPeople, exact: true })
      .click();
    await expect(history.locator("li.hrow")).toHaveCount(2);
    await history
      .getByRole("button", { name: L.historyWithFriends, exact: true })
      .click();
    await expect(history.locator("li.hrow")).toHaveCount(1);
    await expect(history.locator("li.hrow")).toContainText("With Cara");
  });
});
