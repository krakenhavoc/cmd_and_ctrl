// @vitest-environment jsdom
//
// moderation.render.test.ts — #2984, rendered: the admin Reports tab (a
// fixture list, dismissing one, a 403 handed to the page, the empty
// state) and the ⋯ menu's "Report a player…" (offered only with someone
// to report, sends the picked seat and the reason, shows a refusal, says
// thanks once sent).

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import GameMenu from "./components/board/GameMenu.svelte";
import ReportsList from "./components/admin/ReportsList.svelte";
import type { AdminReportsResponse } from "./adminViews";
import type { GameMenuOptions } from "./gameMenu";
import { session, type Session } from "./session";
import { _resetForTests as resetModals } from "./modalLayers";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const NOW = Date.UTC(2026, 9, 13, 12, 0, 0);
const GAME = "9c2f0e4c-1d1e-4c3b-9d7e-2f8a1b2c3d4e";
const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";

const REPORTS: AdminReportsResponse = {
  retention_days: 90,
  reports: [
    {
      id: "r1",
      created_at: NOW - 5 * 60_000,
      game_id: GAME,
      game_name: "Friday night",
      reporter: { player_id: "p1", name: "Alice", user_id: USER },
      reported: { player_id: "p2", name: "Bob" },
      reason: "slurs in chat",
      excerpt: [
        { author: "Alice", text: "hello", at: "2026-10-13T11:50:00Z" },
        { author: "Bob", text: "you are ********", at: "2026-10-13T11:51:00Z" },
      ],
    },
    {
      id: "r2",
      created_at: NOW - 3_600_000,
      game_id: GAME,
      game_name: "",
      reporter: { player_id: "p3", name: "Cat" },
      reported: { player_id: "p2", name: "Bob" },
      reason: "name",
      excerpt: [],
    },
  ],
};

interface Call {
  url: string;
  method: string;
}
let calls: Call[];
let status: number;
let reports: AdminReportsResponse;

function json(body: unknown, code = 200): Response {
  return new Response(JSON.stringify(body), { status: code });
}

beforeEach(() => {
  calls = [];
  status = 200;
  reports = structuredClone(REPORTS);
  resetModals();
  vi.useFakeTimers({ toFake: ["Date"] });
  vi.setSystemTime(NOW);
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      calls.push({ url, method });
      if (status !== 200) return json({ error: "admin only" }, status);
      if (url === "/admin/reports") return json(reports);
      if (url.startsWith("/admin/reports/") && method === "DELETE") {
        return new Response(null, { status: 204 });
      }
      return json({ error: "nope" }, 404);
    }),
  );
  const expiresAt = new Date(NOW + 86_400_000).toISOString();
  const s: Session = {
    token: "admin-tok",
    expiresAt,
    principal: {
      role: "admin",
      name: "admin",
      issued_at: new Date(NOW).toISOString(),
      expires_at: expiresAt,
    },
  };
  session.set(s);
});

afterEach(() => {
  cleanup();
  session.set(null);
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

const text = (el: Element | null | undefined): string =>
  (el?.textContent ?? "").replace(/\s+/g, " ");
const button = (c: ParentNode, name: string): HTMLButtonElement | undefined =>
  [...c.querySelectorAll<HTMLButtonElement>("button")].find((b) => b.textContent?.trim() === name);

describe("the Reports tab", () => {
  it("lists each report with the table, both sides, the reason and the chat excerpt", async () => {
    const forbidden = vi.fn();
    const { container } = render(ReportsList as never, { onforbidden: forbidden } as never);
    await settle();
    expect(calls).toEqual([{ url: "/admin/reports", method: "GET" }]);
    const cards = [...container.querySelectorAll("ul[aria-label='player reports'] > li")];
    expect(cards).toHaveLength(2);
    const first = text(cards[0]);
    expect(first).toContain("Friday night");
    expect(first).toContain("Alice");
    expect(first).toContain("Bob");
    expect(first).toContain("slurs in chat");
    expect(first).toContain("Recent chat (2)");
    expect(first).toContain("you are ********");
    // The reporter has an account (a link), the reported is a guest (a chip).
    expect(cards[0].querySelector('a[href="#/admin/accounts/' + USER + '"]')).not.toBeNull();
    expect(text(cards[0])).toContain("guest");
    expect(cards[0].querySelector(`a[href="#/admin/games/${GAME}"]`)).not.toBeNull();
    // A report with no recorded chat says so, and an unnamed table has a name.
    expect(text(cards[1])).toContain("No chat was recorded");
    expect(text(cards[1])).toContain("untitled table");
    // The page says how long a report lives, and that Dismiss is only the list.
    expect(text(container)).toContain("removed 90 days after");
    expect(forbidden).not.toHaveBeenCalled();
  });

  it("dismisses one report with DELETE and takes it off the list", async () => {
    const { container } = render(ReportsList as never, { onforbidden: vi.fn() } as never);
    await settle();
    click(container.querySelectorAll<HTMLButtonElement>(".actions button")[0]);
    await settle();
    expect(calls.at(-1)).toEqual({ url: "/admin/reports/r1", method: "DELETE" });
    const cards = container.querySelectorAll("ul[aria-label='player reports'] > li");
    expect(cards).toHaveLength(1);
    expect(text(cards[0])).not.toContain("slurs in chat");
  });

  it("says so when there is nothing to read", async () => {
    reports = { ...REPORTS, reports: [] };
    const { container } = render(ReportsList as never, { onforbidden: vi.fn() } as never);
    await settle();
    expect(text(container)).toContain("No reports.");
  });

  it("hands a 403 to the page instead of showing an error", async () => {
    status = 403;
    const forbidden = vi.fn();
    render(ReportsList as never, { onforbidden: forbidden } as never);
    await settle();
    expect(forbidden).toHaveBeenCalledTimes(1);
  });
});

// ---- the ⋯ menu's Report a player ----------------------------------------

function mountMenu(over: Partial<GameMenuOptions> = {}) {
  const sent: Array<[string, string]> = [];
  const r = render(
    GameMenu as never,
    {
      seated: true,
      canManage: false,
      spawnAvailable: false,
      myGames: false,
      voteOpen: false,
      onDraw: () => {},
      onUntapAll: () => {},
      onShuffle: () => {},
      onMulligan: () => {},
      onLifeHistory: () => {},
      onTableSettings: () => {},
      onSpawn: () => {},
      onMyGames: () => {},
      onBack: () => {},
      onConcede: () => {},
      onStartVote: () => {},
      reportTargets: [
        { id: "p2", name: "Bob" },
        { id: "p3", name: "Cat" },
      ],
      onReport: async (id: string, reason: string) => {
        sent.push([id, reason]);
      },
      ...over,
    } as never,
  );
  const trigger = r.container.querySelector<HTMLButtonElement>(
    'button[aria-label="more actions"]',
  )!;
  return { ...r, trigger, sent };
}

const menuItem = (c: ParentNode, name: string): HTMLElement | undefined =>
  [...c.querySelectorAll<HTMLElement>('[role="menuitem"]')].find((el) =>
    (el.textContent ?? "").replace(/\s+/g, " ").trim().startsWith(name),
  );

async function openReport(m: ReturnType<typeof mountMenu>): Promise<HTMLFormElement> {
  click(m.trigger);
  await settle();
  click(menuItem(m.container, "Report a player…")!);
  await settle();
  return m.container.querySelector<HTMLFormElement>("form[aria-label='report a player']")!;
}

function type(el: HTMLTextAreaElement | HTMLSelectElement, value: string): void {
  el.value = value;
  el.dispatchEvent(new Event("input", { bubbles: true }));
  el.dispatchEvent(new Event("change", { bubbles: true }));
  flushSync();
}

describe("Report a player in the ⋯ menu", () => {
  it("is offered to a seated player with someone to report", async () => {
    const m = mountMenu();
    click(m.trigger);
    await settle();
    expect(menuItem(m.container, "Report a player…")).toBeDefined();
  });

  it("is absent with nobody to report, for a spectator, and without a handler", async () => {
    for (const over of [{ reportTargets: [] }, { seated: false }, { onReport: undefined }]) {
      const m = mountMenu(over as Partial<GameMenuOptions>);
      click(m.trigger);
      await settle();
      expect(menuItem(m.container, "Report a player…")).toBeUndefined();
      cleanup();
    }
  });

  it("sends the picked player and the trimmed reason, then says thanks", async () => {
    const m = mountMenu();
    const form = await openReport(m);
    const select = form.querySelector("select")!;
    const reason = form.querySelector("textarea")!;
    const send = button(form, "send")!;
    // Nothing to send until the reason says something.
    expect(send.disabled).toBe(true);
    type(select, "p3");
    type(reason, "  slurs in chat  ");
    expect(send.disabled).toBe(false);
    form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(m.sent).toEqual([["p3", "slurs in chat"]]);
    expect(text(m.container)).toContain("Report sent. Thank you.");
    expect(form.querySelector("textarea")).toBeNull();
  });

  it("keeps the form and shows the server's sentence when the report is refused", async () => {
    const { LobbyApiError } = await import("./session");
    const m = mountMenu({
      onReport: async () => {
        throw new LobbyApiError(429, "too many requests");
      },
    });
    const form = await openReport(m);
    type(form.querySelector("textarea")!, "rude");
    form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(text(m.container)).toMatch(/try again in a few minutes/i);
    expect(form.querySelector("textarea")!.value).toBe("rude");
    expect(text(m.container)).not.toContain("Report sent");
  });
});
