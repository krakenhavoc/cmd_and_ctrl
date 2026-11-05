// @vitest-environment jsdom
//
// adminStats.render.test.ts — #3425, rendered: the admin Stats tab (a
// fixture overview, the backfill note, the empty state, an error and a
// 403 handed to the page), and #4880's stats record v2 blocks (coverage,
// win kinds, combo rules, makeup), with and without the fields.

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
    { name: "AI agent (claude-code)", kind: "agent", ...tally(1, 1, 0, 0, 1) },
    { name: "Deleted players", kind: "deleted", ...tally(2, 1, 1, 0, 2) },
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

const cells = (row: Element): string[] =>
  [...row.querySelectorAll("td")].map((td) => text(td).trim());

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
    expect(rows[2]).toContain("AI agent (claude-code)");
    expect(rows[2]).toContain("agent");
    expect(rows[3]).toContain("Deleted players");
    expect(rows[3]).toContain("deleted");
    expect(text(container)).toContain("Top commanders by games (human seats only)");
    expect(text(container)).toContain("Most-cast cards (human seats only)");

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

  it("shows stats record v2: coverage, how games were won, combo rules and makeup", async () => {
    body = {
      ...structuredClone(STATS),
      coverage: [
        { version: 1, games: 4, frames_complete: 0, frames_incomplete: 0 },
        { version: 2, games: 36, frames_complete: 34, frames_incomplete: 2 },
      ],
      win_kinds: [
        { kind: "commander_damage", games: 7 },
        { kind: "combat", games: 21 },
        { kind: "win_effect", games: 1 },
      ],
      combo_rules: [
        { kind: "loop", games: 2 },
        { kind: "repeats", games: 4 },
      ],
      makeup: { people_only: 22, with_bots: 15, bots_only: 3 },
    } satisfies AdminStatsResponse;
    const { container } = render(StatsView as never, { onforbidden: vi.fn() } as never);
    await settle();

    const coverage = container.querySelector("table[aria-label='record coverage']")!;
    const cov = [...coverage.querySelectorAll("tbody tr")].map(text);
    expect(cov).toHaveLength(2);
    expect(cov[0]).toContain("v1");
    expect(cov[1]).toContain("v2");
    expect(cov[1]).toContain("36");
    expect(cov[1]).toContain("34");
    expect(text(coverage)).toContain("Replay read whole");
    expect(text(coverage)).toContain("Replay not read whole");

    const won = container.querySelector("table[aria-label='how games were won']")!;
    const kinds = [...won.querySelectorAll("tbody tr")].map(cells);
    expect(kinds).toEqual([
      ["Commander damage", "7"],
      ["Combat damage", "21"],
      ["Win-the-game effect", "1"],
    ]);
    expect(text(won)).not.toContain("commander_damage");

    const rules = container.querySelector("table[aria-label='combo wins by rule']")!;
    const rr = [...rules.querySelectorAll("tbody tr")].map(cells);
    expect(rr).toEqual([
      ["Loop", "2"],
      ["Repeats", "4"],
    ]);

    const makeup = text(container.querySelector("[aria-label='table makeup']"));
    expect(makeup).toContain("People only 22");
    expect(makeup).toContain("People and bots 15");
    expect(makeup).toContain("Bots only 3");
  });

  it("treats v2 fields an older server leaves out, or sends empty, as empty", async () => {
    // STATS has none of the v2 fields: an older server.
    const { container } = render(StatsView as never, { onforbidden: vi.fn() } as never);
    await settle();
    expect(text(container)).toContain("Record coverage");
    expect(text(container)).toContain("No coverage reported.");
    expect(text(container)).toContain("No wins recorded.");
    expect(text(container)).toContain("No combo wins recorded.");
    expect(text(container)).toContain("No table makeup recorded.");
    expect(container.querySelector("table[aria-label='record coverage']")).toBeNull();
    expect(container.querySelector("[aria-label='table makeup']")).toBeNull();
    // The v1 lists still render.
    expect(container.querySelector("table[aria-label='recent games']")).not.toBeNull();
    cleanup();

    body = {
      ...structuredClone(STATS),
      coverage: [],
      win_kinds: [],
      combo_rules: [],
      makeup: { people_only: 0, with_bots: 0, bots_only: 0 },
    } satisfies AdminStatsResponse;
    const second = render(StatsView as never, { onforbidden: vi.fn() } as never);
    await settle();
    expect(text(second.container)).toContain("No coverage reported.");
    expect(text(second.container)).toContain("No wins recorded.");
    expect(text(second.container)).toContain("No combo wins recorded.");
    expect(text(second.container)).toContain("No table makeup recorded.");
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
