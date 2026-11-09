import { expect, test, type Page, type Route } from "@playwright/test";
import { L } from "../../client/src/lib/labels";

// ADR 0220 slice 5 = ADR 0217 slice 7 (#4880): a pod's stats page and the
// pod-sharing switch.
//
// The e2e stack has no Discord sign-in, so a signed-in person is a
// session in localStorage and the pod routes answer from this file, as
// pods-4880 does. What is under test is the client: #/pods/<id>/stats
// (the leaderboard with its counts, a member who is not sharing shown
// with no numbers, the range filter reading both routes again), the
// Pod stats link and the sharing switch on a pod's page, the switch in
// Settings, and no sideways scroll at 1440, 820 and 390 px.

const ME = "11111111-1111-4111-8111-111111111111";
const BEA = "22222222-2222-4222-8222-222222222222";
const CAL = "33333333-3333-4333-8333-333333333333";
const POD = "a1111111-1111-4111-8111-111111111111";
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

const tally = (games: number, wins: number) => ({
  games,
  wins,
  losses: games - wins,
  draws: 0,
  avg_place: 2.1,
});

interface Seen {
  method: string;
  path: string;
  search: string;
  body: unknown;
}

async function stub(page: Page, sharing: boolean): Promise<Seen[]> {
  const seen: Seen[] = [];
  let sharePods = sharing;
  const handler = async (route: Route) => {
    const req = route.request();
    const url = new URL(req.url());
    const method = req.method();
    const body = req.postData()
      ? JSON.parse(req.postData() ?? "{}")
      : undefined;
    const p = url.pathname;
    if (p === "/me")
      return json(route, 200, {
        role: "identified",
        user_id: ME,
        display_name: "Alice",
      });
    seen.push({ method, path: p, search: url.search, body });
    if (p === "/me/stats/prefs") {
      if (method === "PUT" && body && "share_pods" in body)
        sharePods = body.share_pods;
      return json(route, 200, {
        available: true,
        advanced: false,
        choice: "default",
        default: false,
        share_pods: sharePods,
      });
    }
    if (p === `/pods/${POD}/stats`)
      return json(route, 200, {
        id: POD,
        name: "Thursday night",
        range: url.searchParams.get("range") ?? "all",
        games: 7,
        min_games: 5,
        members: [
          {
            user_id: ME,
            display_name: "Alice",
            you: true,
            owner: true,
            sharing: sharePods,
            ...(sharePods
              ? { pod_games: tally(7, 4), overall: tally(9, 5) }
              : {}),
          },
          {
            user_id: BEA,
            display_name: "Bea",
            sharing: true,
            pod_games: tally(7, 3),
            overall: tally(8, 3),
          },
          { user_id: CAL, display_name: "Cal", sharing: false },
        ],
        win_kinds: [
          { kind: "combat", games: 5 },
          { kind: "combo", games: 2 },
        ],
        turn_order: [
          { position: 1, wins: { n: 3, of: 5, rate: 0.6 }, avg_place: 2 },
          { position: 2, wins: { n: 1, of: 5, rate: 0.2 }, avg_place: 2.6 },
        ],
        head_to_head: [
          {
            a: { user_id: ME, display_name: "Alice" },
            b: { user_id: BEA, display_name: "Bea" },
            games: 7,
            a_wins: 4,
            b_wins: 3,
            others: 0,
          },
        ],
      });
    if (p === `/pods/${POD}/leaderboard`)
      return json(route, 200, {
        id: POD,
        name: "Thursday night",
        range: url.searchParams.get("range") ?? "all",
        games: 7,
        min_games: 5,
        rows: [
          {
            user_id: BEA,
            display_name: "Bea",
            rank: 1,
            win_rate: 3 / 7,
            ...tally(7, 3),
          },
        ],
        unranked: [
          {
            user_id: CAL,
            display_name: "Cal",
            needs: 3,
            win_rate: 0.5,
            ...tally(2, 1),
          },
        ],
      });
    if (p === `/pods/${POD}`)
      return json(route, 200, {
        id: POD,
        name: "Thursday night",
        role: "owner",
        owner: { user_id: ME, display_name: "Alice" },
        member_count: 3,
        max_members: 12,
        members: [
          {
            user_id: ME,
            display_name: "Alice",
            joined_at: Date.now() - 9e8,
            sharing: sharePods,
            owner: true,
          },
          {
            user_id: BEA,
            display_name: "Bea",
            joined_at: Date.now() - 5e8,
            sharing: true,
          },
        ],
        invites: [],
        link: { live: false },
      });
    return json(route, 200, {});
  };
  await page.route(/\/(me|pods)(\/[^?]*)?(\?.*)?$/, handler);
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

async function shoot(
  page: Page,
  info: import("@playwright/test").TestInfo,
  name: string,
) {
  for (const width of WIDTHS) {
    await page.setViewportSize({ width, height: 900 });
    await noPageScroll(page);
    await info.attach(`${name}-${width}.png`, {
      body: await page.screenshot({ fullPage: true }),
      contentType: "image/png",
    });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}

test.describe("pod stats (ADR 0220 slice 5)", () => {
  test("the board shows counts, a member who is not sharing has no numbers, and the range reads again", async ({
    page,
  }, info) => {
    const seen = await stub(page, true);
    await page.goto(`/#/pods/${POD}/stats`);
    await expect(
      page.getByRole("heading", { level: 1, name: "Thursday night" }),
    ).toBeVisible();
    // The nav's Pods link stays current on a pod's stats page.
    await expect(
      page.getByRole("link", { name: L.navPods, exact: true }).first(),
    ).toHaveAttribute("aria-current", "page");

    const board = page.getByTestId("pod-leaderboard");
    await expect(board.locator("tbody tr")).toHaveCount(1);
    await expect(board.locator("tbody tr").first()).toContainText("Bea");
    await expect(board.locator("tbody tr").first()).toContainText("3-4-0");
    await expect(board).toContainText("Cal has 2, 3 to go");
    await expect(
      page.getByText("Not sharing. Their numbers stay private to them."),
    ).toBeVisible();
    await expect(page.getByTestId("pod-head-to-head")).toContainText(
      "Alice 4, Bea 3 over 7 games",
    );
    await expect(page.getByTestId("pod-stats-not-sharing")).toHaveCount(0);
    await shoot(page, info, "pod-stats");

    await page.getByLabel("30 days").check();
    await expect
      .poll(() =>
        seen.some(
          (s) => s.path === `/pods/${POD}/stats` && s.search === "?range=30d",
        ),
      )
      .toBe(true);
    expect(
      seen.some(
        (s) =>
          s.path === `/pods/${POD}/leaderboard` && s.search === "?range=30d",
      ),
    ).toBe(true);
  });

  test("a viewer who is not sharing is nudged, and the switch is one PUT of share_pods", async ({
    page,
  }, info) => {
    const seen = await stub(page, false);
    await page.goto(`/#/pods/${POD}/stats`);
    const nudge = page.getByTestId("pod-stats-not-sharing");
    await expect(nudge).toBeVisible();
    await shoot(page, info, "pod-stats-not-sharing");
    await nudge.getByLabel(L.sharePodsSwitch).check();
    await expect
      .poll(
        () =>
          seen.find((s) => s.method === "PUT" && s.path === "/me/stats/prefs")
            ?.body,
      )
      .toEqual({ share_pods: true });
    await expect(page.getByTestId("pod-stats-not-sharing")).toHaveCount(0);
  });

  test("a pod's page links to its stats and carries the same switch", async ({
    page,
  }) => {
    await stub(page, false);
    await page.goto(`/#/pods/${POD}`);
    await expect(
      page.getByRole("heading", { level: 1, name: "Thursday night" }),
    ).toBeVisible();
    await expect(page.getByLabel(L.sharePodsSwitch)).not.toBeChecked();
    await page.getByRole("link", { name: L.podStatsLink }).click();
    await expect(page).toHaveURL(new RegExp(`#/pods/${POD}/stats$`));
  });
});
