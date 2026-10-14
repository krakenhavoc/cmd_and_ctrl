// @vitest-environment jsdom
//
// ADR 0165 §7: the stats page shows a signed-in person's own record,
// the same by deck and by commander, and recent games; an empty state
// when they have no finished game; and a sign-in line for a session
// that is not a person.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import Stats from "./Stats.svelte";
import { session, type Session } from "../lib/session";
import type { MyStats } from "../lib/myStats";
import { loadAppConfig, resetAppConfigForTests } from "../lib/env";
import { updateSettings } from "../lib/settings";
import { SHOP_DISCLOSURE } from "../lib/shopLinks";
import { cleanup, flushSync, render } from "../lib/test/render.svelte";

const realFetch = globalThis.fetch;
const USER = "11111111-1111-4111-8111-111111111111";

function person(role: "identified" | "admin" = "identified"): Session {
  const expiresAt = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token: "tok",
    expiresAt,
    principal: {
      role,
      user_id: role === "identified" ? USER : undefined,
      name: "Ana",
      issued_at: new Date().toISOString(),
      expires_at: expiresAt,
    },
  } as Session;
}

const group = (name: string, games: number, wins: number, avg: number | null) => ({
  name,
  games,
  wins,
  losses: games - wins,
  draws: 0,
  avg_place: avg,
});

const FULL: MyStats = {
  games: 12,
  wins: 4,
  losses: 7,
  draws: 1,
  avg_place: 2.4,
  unfinished: 1,
  by_deck: [group("Atraxa Superfriends", 8, 3, 2.1)],
  by_commander: [group("Atraxa, Praetors' Voice", 8, 3, 2.1)],
  recent: [
    {
      game_id: "g1",
      name: "Friday Night",
      ended_at: 1789825000000,
      result: "win",
      place: 1,
      turns: 9,
      players: 4,
      deck: "Atraxa Superfriends",
      commanders: ["Atraxa, Praetors' Voice"],
    },
  ],
};

const EMPTY: MyStats = {
  games: 0,
  wins: 0,
  losses: 0,
  draws: 0,
  avg_place: null,
  unfinished: 0,
  by_deck: [],
  by_commander: [],
  recent: [],
};

function serve(body: MyStats, status = 200): ReturnType<typeof vi.fn> {
  const f = vi.fn(
    async () =>
      new Response(JSON.stringify(status === 200 ? body : { error: "nope" }), {
        status,
        headers: { "Content-Type": "application/json" },
      }),
  );
  globalThis.fetch = f as never;
  return f;
}

// The header probes other routes too; count only this page's.
function statsCalls(f: ReturnType<typeof vi.fn>): number {
  return f.mock.calls.filter((c) => c[0] === "/me/stats").length;
}

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

beforeEach(() => session.set(person()));
afterEach(() => {
  cleanup();
  session.set(null);
  globalThis.fetch = realFetch;
  vi.restoreAllMocks();
});

describe("the stats page", () => {
  it("shows games played, the record, average place, the breakdowns and recent games", async () => {
    const f = serve(FULL);
    const { container } = render(Stats as never, {} as never);
    await settle();
    expect(statsCalls(f)).toBe(1);
    const tiles = container.querySelector("[data-testid=stats-tiles]")!.textContent!;
    expect(tiles).toContain("12");
    expect(tiles).toContain("4-7-1");
    expect(tiles).toContain("2.4");
    const text = container.textContent!.replace(/\s+/g, " ");
    expect(text).toContain("By deck");
    expect(text).toContain("Atraxa Superfriends");
    expect(text).toContain("By commander");
    expect(text).toContain("Recent games");
    expect(text).toContain("Friday Night");
    expect(text).toContain("1st of 4");
    expect(text).toContain("9 turns");
    expect(text).toContain("1 table ended with no result");
    expect(text).not.toMatch(/\bC2\b/);
  });

  it("says to play a game when there are none yet", async () => {
    serve(EMPTY);
    const { container } = render(Stats as never, {} as never);
    await settle();
    expect(container.querySelector("[data-testid=stats-empty]")).not.toBeNull();
    expect(container.querySelector("[data-testid=stats-tiles]")).toBeNull();
  });

  it("asks a session that is not a person to sign in, without calling the server", async () => {
    session.set(person("admin"));
    const f = serve(EMPTY);
    const { container } = render(Stats as never, {} as never);
    await settle();
    expect(statsCalls(f)).toBe(0);
    expect(container.textContent).toContain("Sign in with Discord");
  });

  it("shows the server's error and a retry", async () => {
    serve(EMPTY, 500);
    const { container } = render(Stats as never, {} as never);
    await settle();
    expect(container.querySelector("[role=alert]")).not.toBeNull();
    expect(container.textContent).toContain("Try again");
  });
});

// ADR 0179 §7: Buy links on the most-cast cards, only with a shop.
describe("the stats page's shop links", () => {
  const WITH_CARDS: MyStats = {
    ...FULL,
    top_cards: [
      { name: "Sol Ring", casts: 31 },
      { name: "Arcane Signet", casts: 1 },
    ],
  };

  async function configure(shops: unknown[]): Promise<void> {
    resetAppConfigForTests();
    globalThis.fetch = vi.fn(
      async () => new Response(JSON.stringify({ env: "prod", shops })),
    ) as never;
    await loadAppConfig();
    flushSync();
  }

  afterEach(() => {
    resetAppConfigForTests();
    updateSettings("display", "shopLinks", true);
  });

  it("links each most-cast card to the shops, with the disclosure", async () => {
    await configure([{ id: "tcgplayer", template: "{url}&partner=c2" }]);
    serve(WITH_CARDS);
    const { container } = render(Stats as never, {} as never);
    await settle();
    const panel = container.querySelector("[data-testid=top-cards]")!;
    expect(panel.textContent).toContain("Sol Ring");
    expect(panel.textContent).toContain("31 casts");
    expect(panel.textContent).toContain("1 cast");
    expect(panel.querySelectorAll("a").length).toBe(2);
    expect(panel.querySelector("a")!.getAttribute("href")).toContain("tcgplayer.com");
    expect(panel.textContent).toContain(SHOP_DISCLOSURE);
    expect(panel.textContent).not.toMatch(/\$\d/);
  });

  it("renders nothing without a configured shop or when hidden", async () => {
    await configure([]);
    serve(WITH_CARDS);
    let r = render(Stats as never, {} as never);
    await settle();
    expect(r.container.querySelector("[data-testid=top-cards]")).toBeNull();
    cleanup();

    await configure([{ id: "tcgplayer", template: "{url}&partner=c2" }]);
    updateSettings("display", "shopLinks", false);
    serve(WITH_CARDS);
    r = render(Stats as never, {} as never);
    await settle();
    expect(r.container.querySelector("[data-testid=top-cards]")).toBeNull();
  });
});
