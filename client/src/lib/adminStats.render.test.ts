// @vitest-environment jsdom
//
// adminStats.render.test.ts — #3425, rendered: the admin Stats tab (a
// fixture overview, the backfill note, the empty state, an error and a
// 403 handed to the page).

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import StatsView from "./components/admin/StatsView.svelte";
import type { AdminStatsResponse } from "./adminViews";
import { session, type Session } from "./session";
import { cleanup, flushSync, render } from "./test/render.svelte";

const NOW = Date.UTC(2026, 9, 14, 12, 0, 0);
const GAME = "9c2f0e4c-1d1e-4c3b-9d7e-2f8a1b2c3d4e";

const tally = (games: number, wins: number, losses: number, draws: number, avg: number | null) => ({
  games,
  wins,
  losses,
  draws,
  avg_place: avg,
});

const STATS: AdminStatsResponse = {
  totals: { games: 3, live: 2, replay: 1, log_complete: 2, log_incomplete: 1, unrecorded: 2 },
  recent: [
    {
      game_id: GAME,
      name: "Friday night",
      ended_at: NOW - 60_000,
      turns: 21,
      rounds: 6,
      source: "live",
    },
  ],
  people: [
    { name: "Alice", kind: "account", ...tally(3, 2, 1, 0, 1.5) },
    { name: "Bot (heuristic)", kind: "bot", ...tally(2, 0, 2, 0, null) },
  ],
  commanders_by_games: [{ name: "Tymna / Thrasios", ...tally(3, 1, 1, 1, 2) }],
  commanders_by_wins: [{ name: "Atraxa", ...tally(2, 2, 0, 0, 1) }],
  cards: [{ name: "Sol Ring", casts: 9, games: 4 }],
};

let calls: string[];
let status: number;
let body: unknown;

beforeEach(() => {
  calls = [];
  status = 200;
  body = structuredClone(STATS);
  vi.useFakeTimers({ toFake: ["Date"] });
  vi.setSystemTime(NOW);
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string) => {
      calls.push(url);
      return new Response(JSON.stringify(status === 200 ? body : { error: "nope" }), { status });
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

describe("the Stats tab", () => {
  it("shows totals, the backfill note, and every list", async () => {
    const { container } = render(StatsView as never, { onforbidden: vi.fn() } as never);
    await settle();
    expect(calls).toEqual(["/admin/stats"]);
    const totals = text(container.querySelector("[aria-label='statistics totals']"));
    expect(totals).toContain("Games with a record 3");
    expect(totals).toContain("Log incomplete 1");
    expect(totals).toContain("Ended, no record 2");
    expect(totals).toContain("2 finished games have no record yet");

    const recent = container.querySelector("table[aria-label='recent games']")!;
    expect(text(recent)).toContain("Friday night");
    expect(recent.querySelector(`a[href="#/admin/games/${GAME}"]`)).not.toBeNull();
    expect(text(recent)).toContain("21");

    const people = container.querySelector("table[aria-label='records by person']")!;
    const rows = [...people.querySelectorAll("tbody tr")].map(text);
    expect(rows[0]).toContain("Alice");
    expect(rows[0]).toContain("2-1");
    expect(rows[0]).toContain("1.5");
    expect(rows[1]).toContain("Bot (heuristic)");
    expect(rows[1]).toContain("bot");

    expect(text(container.querySelector("table[aria-label='commanders by games']"))).toContain(
      "Tymna / Thrasios",
    );
    expect(text(container.querySelector("table[aria-label='commanders by wins']"))).toContain(
      "Atraxa",
    );
    expect(text(container.querySelector("table[aria-label='most-cast cards']"))).toContain(
      "Sol Ring",
    );
  });

  it("says nothing is recorded yet, with no backfill note, when there is nothing", async () => {
    body = {
      totals: { games: 0, live: 0, replay: 0, log_complete: 0, log_incomplete: 0, unrecorded: 0 },
      recent: [],
      people: [],
      commanders_by_games: [],
      commanders_by_wins: [],
      cards: [],
    } satisfies AdminStatsResponse;
    const { container } = render(StatsView as never, { onforbidden: vi.fn() } as never);
    await settle();
    expect(text(container)).toContain("No game statistics yet");
    expect(text(container)).not.toContain("backfill");
    expect(container.querySelector("table")).toBeNull();
  });

  it("shows a failed load, and hands a 403 to the page", async () => {
    status = 500;
    const forbidden = vi.fn();
    const { container } = render(StatsView as never, { onforbidden: forbidden } as never);
    await settle();
    expect(container.querySelector("[role='alert']")).not.toBeNull();
    expect(forbidden).not.toHaveBeenCalled();
    cleanup();

    status = 403;
    render(StatsView as never, { onforbidden: forbidden } as never);
    await settle();
    expect(forbidden).toHaveBeenCalledTimes(1);
  });
});
