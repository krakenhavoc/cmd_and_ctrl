import { expect, test, type Page, type Route } from "@playwright/test";
import { L } from "../../client/src/lib/labels";

// ADR 0220 slice 3 (#4880): the pod pages.
//
// The e2e stack has no Discord sign-in, so a signed-in person is a
// session in localStorage and the pod routes answer from this file, the
// way friends-3313 does it. What is under test is the client: #/pods
// (invites, pods, New pod), a pod's page (members, the friend picker, the
// invite link, Remove with its new-link box, Leave), the invite link
// page, the nav link, and no sideways scroll at 1440, 820 and 390 px.

const ME = "11111111-1111-4111-8111-111111111111";
const BEA = "22222222-2222-4222-8222-222222222222";
const CAL = "33333333-3333-4333-8333-333333333333";
const DEE = "44444444-4444-4444-8444-444444444444";
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

interface Seen {
  method: string;
  path: string;
  search: string;
  body: unknown;
}

async function stub(page: Page): Promise<Seen[]> {
  const seen: Seen[] = [];
  const now = Date.now();
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
    if (!p.includes("pod") && p !== "/me/friends") return json(route, 200, {});
    seen.push({ method, path: p, search: url.search, body });
    if (p === "/me/pods")
      return json(route, 200, {
        pods: [
          { id: POD, name: "Thursday night", member_count: 3, role: "owner" },
          { id: "b2", name: "Draft crew", member_count: 5, role: "member" },
        ],
        invites: [
          {
            pod_id: "c3",
            pod_name: "Friday",
            invited_by: { display_name: "Bea" },
            since: now,
          },
        ],
        suggestions: [],
        max_pods: 10,
        max_members: 12,
      });
    if (p === "/me/friends")
      return json(route, 200, {
        friends: [
          { user_id: BEA, display_name: "Bea", since: now },
          { user_id: DEE, display_name: "Dee", since: now },
        ],
        incoming: [],
        outgoing: [],
      });
    if (p === `/pods/${POD}` && method === "GET")
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
            joined_at: now - 9e8,
            sharing: true,
            owner: true,
          },
          {
            user_id: BEA,
            display_name: "Bea",
            joined_at: now - 5e8,
            sharing: true,
          },
          {
            user_id: CAL,
            display_name: "Cal",
            joined_at: now - 1e8,
            sharing: false,
          },
        ],
        invites: [],
        link: { live: true, expires_at: now + 5 * 86_400_000 },
      });
    if (p === `/pods/${POD}/link` && method === "POST")
      return json(route, 201, {
        token: "TOKTOKTOKTOKTOKTOKTOK",
        path: "#/pods/join/TOKTOKTOKTOKTOKTOKTOK",
        expires_at: now + 7 * 86_400_000,
      });
    if (p.startsWith("/pods/join/")) {
      if (p.endsWith("/dead"))
        return json(route, 404, { error: "This invite link can't be used." });
      if (method === "POST")
        return json(route, 200, {
          id: POD,
          name: "Thursday night",
          member_count: 4,
          role: "member",
        });
      return json(route, 200, {
        name: "Thursday night",
        owner_name: "Alice",
        member_count: 3,
        member: false,
      });
    }
    if (p === `/pods/${POD}/members/${CAL}` && req.url().includes("new_link"))
      return json(route, 200, {
        link: {
          token: "NEW",
          path: "#/pods/join/NEW",
          expires_at: now + 7 * 86_400_000,
        },
      });
    return route.fulfill({ status: 204 });
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

test.describe("pods (ADR 0220 slice 3)", () => {
  test("#/pods marks its nav link current, and lists invites and pods", async ({
    page,
  }, info) => {
    const seen = await stub(page);
    await page.goto("/#/pods");
    // The nav's Pods link is the current page.
    await expect(
      page.getByRole("link", { name: L.navPods, exact: true }).first(),
    ).toHaveAttribute("aria-current", "page");
    await expect(
      page.getByRole("heading", { level: 1, name: "Your pods" }),
    ).toBeVisible();
    await expect(page.locator("#p-invites").locator("xpath=..")).toContainText(
      "Bea invited you",
    );
    await expect(page.locator(".pod-row").first()).toContainText("Owner");
    await shoot(page, info, "pods");

    await page.getByRole("button", { name: L.acceptPodInvite }).click();
    await expect
      .poll(() => seen.some((s) => s.path === "/me/pod-invites/c3/accept"))
      .toBe(true);
  });

  test("a pod: the owner invites a friend, makes a link, removes with a new link, leaves", async ({
    page,
  }, info) => {
    const seen = await stub(page);
    await page.goto(`/#/pods/${POD}`);
    await expect(
      page.getByRole("heading", { level: 1, name: "Thursday night" }),
    ).toBeVisible();
    await expect(page.getByText("Not sharing", { exact: true })).toBeVisible();
    await shoot(page, info, "pod");

    await page.getByRole("button", { name: L.inviteToPod }).click();
    // Bea is a member; only Dee is offered.
    await expect(page.locator('input[name="friend"]')).toHaveCount(1);
    await page.getByLabel("Dee").check();
    await page.getByRole("button", { name: "Send invite" }).click();
    await expect
      .poll(() => seen.find((s) => s.path === `/pods/${POD}/invites`)?.body)
      .toEqual({ user_id: DEE });

    await page.getByRole("button", { name: L.makePodLink }).click();
    await expect(page.locator('input[aria-label="Invite link"]')).toHaveValue(
      /#\/pods\/join\/TOKTOK/,
    );
    await expect(
      page.getByRole("button", { name: L.copyPodLink }),
    ).toBeVisible();

    await page.getByRole("button", { name: "actions for Cal" }).click();
    await page.getByRole("menuitem", { name: L.removePodMember }).click();
    await expect(page.getByText("Also make a new invite link")).toBeVisible();
    await page
      .locator(".confirm")
      .getByRole("button", { name: L.removePodMember })
      .click();
    await expect
      .poll(() =>
        seen.some(
          (s) =>
            s.method === "DELETE" &&
            s.path === `/pods/${POD}/members/${CAL}` &&
            s.search === "?new_link=1",
        ),
      )
      .toBe(true);

    await page.getByRole("button", { name: L.leavePod }).click();
    await expect(page.getByText("Leave Thursday night?")).toBeVisible();
    await page.getByRole("button", { name: L.leavePod }).last().click();
    await expect(page).toHaveURL(/#\/pods$/);
  });

  test("the invite link: preview and Join, a dead link, and the signed-out line", async ({
    page,
  }, info) => {
    const seen = await stub(page);
    await page.goto("/#/pods/join/goodtoken");
    await expect(page.getByText("Owned by Alice")).toBeVisible();
    await shoot(page, info, "pod-join");
    await page.getByRole("button", { name: L.joinPod }).click();
    await expect(page).toHaveURL(new RegExp(`#/pods/${POD}$`));
    expect(
      seen.some(
        (s) => s.method === "POST" && s.path === "/pods/join/goodtoken",
      ),
    ).toBe(true);

    await page.goto("/#/pods/join/dead");
    await expect(
      page.getByText("This invite link can’t be used."),
    ).toBeVisible();
  });

  test("signed out, the invite link shows only a sign-in line", async ({
    page,
  }) => {
    await page.goto("/#/pods/join/anything");
    await expect(
      page.getByText("Sign in with Discord to join this pod"),
    ).toBeVisible();
    await expect(page.getByText("Thursday night")).toHaveCount(0);
  });
});
