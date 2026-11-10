// @vitest-environment jsdom
//
// ADR 0220 slice 5 (ADR 0217 slice 7): a pod's stats page
// (#/pods/<id>/stats), the sharing switch, and the words behind them.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import PodShareSwitch from "./components/PodShareSwitch.svelte";
import PodStats from "../routes/PodStats.svelte";
import { L } from "./labels";
import {
  boardCaption,
  headToHeadLine,
  podGamesLine,
  recordLine,
  winRateLine,
  type PodLeaderboard,
} from "./podStats";
import { setSession, type Session } from "./session";
import { STATS_PREFS_OFF, statsPrefs } from "./statsPrefs";
import { cleanup, flushSync, render } from "./test/render.svelte";

const ME = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const BEA = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const CAL = "7a1d3c10-0f6e-4d46-8d1b-3a1f4d7a9c22";
const POD = "a1111111-1111-4111-8111-111111111111";

function sess(): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: "tok",
    expiresAt: expires,
    principal: {
      role: "identified",
      user_id: ME,
      name: "Ann",
      issued_at: expires,
      expires_at: expires,
    },
  } as Session;
}

function reply(status: number, payload: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: "x",
    json: async () => payload,
    clone() {
      return this;
    },
  } as unknown as Response;
}

const tally = (games: number, wins: number) => ({
  games,
  wins,
  losses: games - wins,
  draws: 0,
  avg_place: 2.1,
});

let calls: { method: string; path: string; body?: unknown }[] = [];
let sharing = true;
let status = 200;

function stats() {
  return {
    id: POD,
    name: "Thursday",
    range: "all",
    games: 7,
    min_games: 5,
    members: [
      {
        user_id: ME,
        display_name: "Ann",
        you: true,
        owner: true,
        sharing,
        pod_games: tally(7, 4),
        overall: tally(9, 5),
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
    turn_order: [{ position: 1, wins: { n: 3, of: 5, rate: 0.6 }, avg_place: 2 }],
    head_to_head: [
      {
        a: { user_id: ME, display_name: "Ann" },
        b: { user_id: BEA, display_name: "Bea" },
        games: 7,
        a_wins: 4,
        b_wins: 3,
        others: 0,
      },
    ],
  };
}

function leaderboard() {
  return {
    id: POD,
    name: "Thursday",
    range: "all",
    games: 7,
    min_games: 5,
    rows: [
      { user_id: ME, display_name: "Ann", you: true, rank: 1, win_rate: 4 / 7, ...tally(7, 4) },
      { user_id: BEA, display_name: "Bea", rank: 2, win_rate: 3 / 7, ...tally(7, 3) },
    ],
    unranked: [{ user_id: CAL, display_name: "Cal", win_rate: 0.5, needs: 3, ...tally(2, 1) }],
  };
}

beforeEach(() => {
  calls = [];
  sharing = true;
  status = 200;
  setSession(sess());
  statsPrefs.set({ ...STATS_PREFS_OFF, available: true, share_pods: true });
  vi.stubGlobal(
    "fetch",
    vi.fn(async (path: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      const body = init.body ? JSON.parse(init.body as string) : undefined;
      calls.push({ method, path, body });
      if (path === "/me/stats/prefs" && method === "PUT")
        return reply(200, { ...STATS_PREFS_OFF, available: true, share_pods: body.share_pods });
      if (path === "/me/stats/prefs")
        return reply(200, { ...STATS_PREFS_OFF, available: true, share_pods: sharing });
      if (status !== 200) return reply(status, { error: "no such pod" });
      if (path.startsWith(`/pods/${POD}/stats`)) return reply(200, stats());
      if (path.startsWith(`/pods/${POD}/leaderboard`)) return reply(200, leaderboard());
      return reply(404, { error: "unexpected " + path });
    }),
  );
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
  statsPrefs.set(STATS_PREFS_OFF);
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

describe("pod stats words", () => {
  it("words a record, a rate and the pod's games", () => {
    expect(recordLine({ wins: 5, losses: 2, draws: 1 })).toBe("5-2-1");
    expect(winRateLine({ games: 8, wins: 5 })).toBe("63%");
    expect(winRateLine({ games: 0, wins: 0 })).toBe("-");
    expect(podGamesLine({ games: 1, range: "30d" })).toBe(
      "1 game in the last 30 days where two or more sharing members sat.",
    );
  });
  it("captions the board with the leader's count, or why it is empty", () => {
    const b = leaderboard() as unknown as PodLeaderboard;
    expect(boardCaption(b)).toBe("Ann leads with 57% over 7 games.");
    expect(boardCaption({ ...b, rows: [], games: 0 })).toMatch(/No pod games yet/);
    expect(boardCaption({ ...b, rows: [] })).toBe("Nobody has 5 pod games yet in this range.");
  });
  it("words a head to head", () => {
    expect(headToHeadLine(stats().head_to_head[0])).toBe("Ann 4, Bea 3 over 7 games");
  });
});

describe("#/pods/<id>/stats", () => {
  it("shows the board with counts, members with Not sharing and no numbers, and the figures", async () => {
    const r = render(PodStats, { id: POD });
    await settle();
    const root = r.container;
    const rows = [...root.querySelectorAll("[data-testid=pod-leaderboard] tbody tr")];
    expect(rows).toHaveLength(2);
    expect(rows[0].textContent).toContain("Ann (you)");
    expect(rows[0].textContent).toContain("4-3-0");
    expect(rows[0].textContent).toContain("7");
    expect(root.textContent).toContain("Cal has 2, 3 to go");
    const members = root.textContent ?? "";
    expect(members).toContain("Not sharing. Their numbers stay private to them.");
    expect(root.querySelector("[data-testid=pod-head-to-head]")?.textContent).toContain(
      "Ann 4, Bea 3 over 7 games",
    );
    expect(root.querySelectorAll("[data-testid=stats-chart]")).toHaveLength(2);
    // Sharing: no nudge to start.
    expect(root.querySelector("[data-testid=pod-stats-not-sharing]")).toBeNull();
    expect(calls.some((c) => c.path === `/pods/${POD}/stats?range=all`)).toBe(true);
  });

  it("nudges a viewer who is not sharing, who can still look", async () => {
    statsPrefs.set({ ...STATS_PREFS_OFF, available: true, share_pods: false });
    sharing = false;
    const r = render(PodStats, { id: POD });
    await settle();
    expect(r.container.querySelector("[data-testid=pod-stats-not-sharing]")).not.toBeNull();
    expect(r.container.querySelectorAll("[data-testid=pod-leaderboard] tbody tr")).toHaveLength(2);
  });

  it("changing the range reads both routes again", async () => {
    const r = render(PodStats, { id: POD });
    await settle();
    const radio = [
      ...r.container.querySelectorAll<HTMLInputElement>('input[name="stats-range"]'),
    ].find((i) => i.value === "30d");
    radio?.click();
    await settle();
    expect(calls.some((c) => c.path === `/pods/${POD}/stats?range=30d`)).toBe(true);
    expect(calls.some((c) => c.path === `/pods/${POD}/leaderboard?range=30d`)).toBe(true);
  });

  it("says a pod that is not there is gone", async () => {
    status = 404;
    const r = render(PodStats, { id: POD });
    await settle();
    expect(r.container.textContent).toContain("That pod isn’t there.");
  });
});

describe("the pod-sharing switch", () => {
  it("is named by its label, and PUTs share_pods alone", async () => {
    statsPrefs.set({ ...STATS_PREFS_OFF, available: true, share_pods: false });
    const r = render(PodShareSwitch, {});
    await settle();
    const box = r.container.querySelector<HTMLInputElement>("input[type=checkbox]");
    expect(r.container.textContent).toContain(L.sharePodsSwitch);
    expect(box?.checked).toBe(false);
    box?.click();
    await settle();
    expect(calls.find((c) => c.method === "PUT")?.body).toEqual({ share_pods: true });
    expect(box?.checked).toBe(true);
  });
});
