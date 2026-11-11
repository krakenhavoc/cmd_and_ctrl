// @vitest-environment jsdom
//
// #5440 (ADR 0222 §9): the My drafts list and the read-only page of one
// finished draft, drawn from GET /me/drafts as docs/lobby.md gives it.

import { afterEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import MyDrafts from "./MyDrafts.svelte";
import { L } from "../lib/labels";
import { cleanup, flushSync, render } from "../lib/test/render.svelte";

const realFetch = globalThis.fetch;
const ID = "6f1c2a9e-0000-4000-8000-000000000000";

function serve(routes: Record<string, { status?: number; body?: unknown }>): string[] {
  const asked: string[] = [];
  globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    asked.push(url);
    const a = routes[url] ?? {};
    return new Response(JSON.stringify(a.body ?? {}), {
      status: a.status ?? 200,
      headers: { "Content-Type": "application/json" },
    });
  }) as never;
  return asked;
}

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
});

const text = (el: Element | null | undefined): string =>
  (el?.textContent ?? "").replace(/\s+/g, " ").trim();

const RECORD = {
  id: ID,
  name: "Friday draft",
  set: "blb",
  set_name: "Bloomburrow",
  mode: "pick_one",
  seats: 8,
  best_of: 1,
  seat: 2,
  drafted_at: "2026-11-11T19:02:11Z",
  matches: [
    {
      round: 1,
      opponent_kind: "person",
      opponent_name: "Bea",
      result: "win",
      games_won: 1,
      games_lost: 0,
    },
    { round: 2, opponent_kind: "bot", result: "loss", games_won: 0, games_lost: 1 },
    { round: 3, opponent_kind: "bye", result: "", games_won: 0, games_lost: 0 },
  ],
  wins: 1,
  losses: 1,
  draws: 0,
};

describe("My drafts", () => {
  it("lists the caller's drafts with the set, the record and a link to each", async () => {
    serve({ "/me/drafts": { body: { drafts: [RECORD] } } });
    const { container: c } = render(MyDrafts as never, {} as never);
    await settle();
    expect(text(c.querySelector("h1"))).toBe(L.myDrafts);
    const rows = [...c.querySelectorAll('[data-testid="my-drafts"] li')];
    expect(rows).toHaveLength(1);
    expect(text(rows[0])).toContain("Friday draft");
    expect(text(rows[0])).toContain("Bloomburrow");
    expect(text(rows[0])).toContain("1–1");
    expect(rows[0].querySelector("a")!.getAttribute("href")).toBe(`#/drafts/${ID}`);
  });

  it("says so when there are none, with the way to a draft", async () => {
    serve({ "/me/drafts": { body: { drafts: [] } } });
    const { container: c } = render(MyDrafts as never, {} as never);
    await settle();
    expect(text(c)).toContain("No drafts yet.");
    expect(text(c)).toContain("Drafting without signing in keeps nothing.");
    expect([...c.querySelectorAll("a")].some((a) => a.getAttribute("href") === "#/draft")).toBe(
      true,
    );
  });

  it("shows one draft read-only: its rounds, its deck and its sideboard", async () => {
    const asked = serve({
      [`/me/drafts/${ID}`]: {
        body: {
          draft: {
            ...RECORD,
            basics: { Forest: 9, Island: 8 },
            picks: [
              {
                name: "Shoreline Looter",
                scryfall_id: "sf-1",
                colors: ["U"],
                mana_cost: "{1}{U}",
                mana_value: 2,
                type_line: "Creature",
                in_deck: true,
              },
              {
                name: "Fountainport Bell",
                scryfall_id: "sf-2",
                colors: [],
                mana_cost: "{1}",
                mana_value: 1,
                type_line: "Artifact",
                in_deck: false,
              },
            ],
          },
        },
      },
    });
    const { container: c } = render(MyDrafts as never, { id: ID } as never);
    await settle();
    expect(asked).toContain(`/me/drafts/${ID}`);
    expect(text(c.querySelector("h1"))).toBe("Friday draft");
    expect([...c.querySelectorAll('[data-testid="my-draft-rounds"] li')].map(text)).toEqual([
      "Round 1 Won 1–0 against Bea",
      "Round 2 Lost 0–1 to a bot",
      "Round 3 Sat the round out",
    ]);
    const deck = c.querySelector('[data-testid="my-draft-deck"]')!;
    expect(text(deck.querySelector("h2"))).toBe("Deck 18");
    expect(text(deck)).toContain("Basic lands added: 8 Island, 9 Forest.");
    expect(deck.querySelector("img")!.getAttribute("src")).toBe("/draft/image/sf-1?size=normal");
    const side = c.querySelector('[data-testid="my-draft-side"]')!;
    expect(text(side.querySelector("h2"))).toBe("Sideboard 1");
    expect(side.querySelector("img")!.alt).toBe("Fountainport Bell");
    // Read-only: nothing on the page changes the draft.
    expect(c.querySelectorAll("main button")).toHaveLength(0);
  });

  it("says a draft that is not the caller's is not here", async () => {
    serve({ [`/me/drafts/${ID}`]: { status: 404, body: { error: "no such draft" } } });
    const { container: c } = render(MyDrafts as never, { id: ID } as never);
    await settle();
    expect(text(c.querySelector('[role="alert"]'))).toContain("That draft isn’t here.");
  });
});
