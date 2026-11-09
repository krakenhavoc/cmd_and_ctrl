// @vitest-environment jsdom
//
// statsDeckPage.render.test.ts — ADR 0217 slice 6 part 2: one deck's
// advanced stats page. The charts and their text equivalents, the
// version picker, the opt-in while the setting is off, the shared view's
// read-only copy and the overview's links to it. The routes are a
// stubbed fetch.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

import AdvancedOverview from "./components/stats/AdvancedOverview.svelte";
import DeckStats from "./components/stats/DeckStats.svelte";
import StatsDeck from "../routes/StatsDeck.svelte";
import { L } from "./labels";
import { STATS_PREFS_OFF, statsPrefs, type StatsPrefs } from "./statsPrefs";
import { session, type Session } from "./session";
import { cleanup, render } from "./test/render.svelte";
import { deckFixture, overviewFixture } from "./test/advancedStatsFixture";

const person = (): Session =>
  ({
    token: "tok",
    expiresAt: "2999-01-01T00:00:00Z",
    principal: { role: "player", user_id: "u1", game_id: "g1", player_id: "p1" },
  }) as unknown as Session;

let prefs: StatsPrefs;
let calls: string[];
let deckStatus: number;

function json(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

const DECK = deckFixture().deck.id;

beforeEach(() => {
  session.set(person());
  prefs = {
    available: true,
    advanced: true,
    choice: "on",
    set_by: "self",
    default: false,
    share_pods: false,
  };
  statsPrefs.set(prefs);
  calls = [];
  deckStatus = 200;
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = typeof input === "string" ? input : input.toString();
      calls.push(url);
      if (url.startsWith("/me/stats/prefs")) {
        if (init?.method === "PUT") {
          prefs = { ...prefs, advanced: true, choice: "on", set_by: "self" };
        }
        return json(200, prefs);
      }
      if (url.startsWith(`/me/stats/decks/${DECK}`))
        return deckStatus === 200 ? json(200, deckFixture()) : json(deckStatus, { error: "x" });
      if (url.startsWith("/stats/shared/tok/decks/"))
        return json(200, { ...deckFixture(), display_name: "Bea" });
      return json(404, { error: "not found" });
    }),
  );
});

afterEach(() => {
  cleanup();
  statsPrefs.set(STATS_PREFS_OFF);
  session.set(null);
  vi.unstubAllGlobals();
});

describe("the deck page's body", () => {
  it("lists its sections in ADR 0217 §8's order and names every chart by a real caption", () => {
    const r = render(DeckStats, { deck: deckFixture() });
    flushSync();
    const heads = [...r.container.querySelectorAll("h2")].map((h) => h.textContent);
    expect(heads).toEqual(["Mana", "Cards", "Composition", "Play", "Commander"]);
    const imgs = [...r.container.querySelectorAll("[role=img][aria-labelledby]")];
    expect(imgs.length).toBeGreaterThan(7);
    for (const img of imgs) {
      const cap = r.container.querySelector(`#${img.getAttribute("aria-labelledby")}`);
      expect(cap?.tagName, img.outerHTML.slice(0, 80)).toBe("FIGCAPTION");
      expect(cap?.textContent?.trim().length).toBeGreaterThan(10);
    }
    const summaries = new Set(
      [...r.container.querySelectorAll("summary")].map((s) => s.textContent?.trim()),
    );
    expect(summaries).toEqual(new Set([L.showAsTable]));
  });

  it("puts the mana curve and the land-drop curve on one axis, with a table of both", () => {
    const r = render(DeckStats, { deck: deckFixture() });
    flushSync();
    const fig = r.container.querySelector("#cap-deck-curves")!.closest("figure")!;
    expect(fig.querySelectorAll("svg")).toHaveLength(1);
    // The axis runs 0 … 7+ for both panels.
    const axis = [...fig.querySelectorAll("svg text.axis")].map((t) => t.textContent);
    expect(axis).toContain("7+");
    const rows = [...fig.querySelectorAll("tbody tr")].map((tr) =>
      [...tr.children].map((c) => c.textContent?.trim()),
    );
    expect(rows).toHaveLength(8);
    expect(rows[0]).toEqual(["0", "2", "-", "-"]);
    expect(rows[4][0]).toBe("4");
    expect(rows[4][1]).toBe("10");
    expect(rows[7][0]).toBe("7+");
    expect(fig.querySelector("figcaption")?.textContent).toMatch(/land drop 4 by turn 4/);
  });

  it("shows dead cards with their rate and a could/couldn't afford tag in words", () => {
    const r = render(DeckStats, { deck: deckFixture() });
    flushSync();
    const cards = [...r.container.querySelectorAll("[data-testid=deck-cards] .card-grid li")];
    expect(cards.length).toBe(2 + 3);
    const hoof = cards[0];
    expect(hoof.textContent).toContain("Craterhoof Behemoth");
    expect(hoof.textContent).toContain("Dead in 4 of 5 games in hand (80%)");
    expect(hoof.textContent).toContain("Couldn't afford");
    expect(hoof.querySelector("img")?.getAttribute("src")).toContain(
      "33333333-3333-4333-8333-333333333333",
    );
    expect(cards[1].textContent).toContain("Could afford");
    // No printing: the name stands alone.
    expect(cards[1].querySelector("img")).toBeNull();
  });

  it("compares composition against the other decks and credits the roles", () => {
    const r = render(DeckStats, { deck: deckFixture() });
    flushSync();
    const comp = r.container.querySelector("[data-testid=deck-composition]")!;
    expect(comp.textContent).toContain("Lands");
    expect(comp.textContent).toContain("others 36.3");
    expect(comp.textContent).toContain("Removal");
    expect(comp.textContent).toContain("Card roles from");
    expect(comp.querySelector("a[href='https://tagger.scryfall.com']")).not.toBeNull();
    expect(comp.textContent).toMatch(/2\s+cards in the list aren't in our card data/);
  });

  it("says roles are unavailable, and counts card types alone, without a roles file", () => {
    const d = deckFixture();
    d.deck.facts!.roles_available = false;
    d.deck.facts!.roles = null;
    d.other_decks!.roles = undefined;
    const r = render(DeckStats, { deck: d });
    flushSync();
    const comp = r.container.querySelector("[data-testid=deck-composition]")!;
    expect(comp.textContent).toContain("Card roles aren't available on this server");
    expect(comp.textContent).not.toContain("Card roles from");
    expect(comp.textContent).toContain("Creatures");
  });

  it("asks for games, never a zero chart, on a deck with none in the filter", () => {
    const d = deckFixture();
    d.record = { games: 0, wins: 0, losses: 0, draws: 0, avg_place: null };
    d.win_rate = { n: 0, of: 0, rate: null };
    d.dead_cards = [];
    d.most_cast = [];
    const r = render(DeckStats, { deck: d });
    flushSync();
    expect(r.container.textContent).toContain(
      "Play 3 games with this deck to see how often you miss",
    );
    expect(r.container.querySelector("[data-testid=deck-play] [role=img]")).toBeNull();
    // The list-based charts are still there.
    expect(r.container.querySelector("#cap-deck-curves")).not.toBeNull();
    expect(r.container.querySelector("#cap-deck-types")).not.toBeNull();
  });

  it("draws no card image in a shared view", () => {
    const own = render(DeckStats, { deck: deckFixture() });
    flushSync();
    expect(own.container.querySelectorAll(".card-grid img").length).toBeGreaterThan(0);
    cleanup();
    const shared = render(DeckStats, { deck: deckFixture(), shared: true });
    flushSync();
    expect(shared.container.querySelectorAll(".card-grid img")).toHaveLength(0);
    expect(shared.container.textContent).toContain("Craterhoof Behemoth");
  });
});

describe("#/stats/deck/<id>", () => {
  it("reads the deck since its last edit, and shows the version picker with earlier lists", async () => {
    const r = render(StatsDeck, { deckId: DECK });
    await settle();
    expect(
      calls.some((u) =>
        u.startsWith(`/me/stats/decks/${DECK}?makeup=all&range=all&version=current`),
      ),
    ).toBe(true);
    expect(r.container.querySelector("[data-testid=deck-stats]")).not.toBeNull();
    expect(r.container.querySelector("h1")?.textContent).toContain("Atraxa Superfriends");
    const select = r.container.querySelector<HTMLSelectElement>("[data-testid=deck-version]")!;
    expect(select.labels?.[0]?.textContent).toBe(L.statsVersionPicker);
    const options = [...select.options].map((o) => o.textContent?.trim());
    expect(options[0]).toBe("Since your last edit");
    expect(options[1]).toBe("All time");
    expect(options[2]).toMatch(/^Earlier list, .+ to .+ \(3 games\)$/);
    expect(options).toHaveLength(3);
    // Choosing a version is one read of that version.
    select.value = "0f0f0f0f0f0f0f0f";
    select.dispatchEvent(new Event("change", { bubbles: true }));
    await settle();
    expect(calls.at(-1)).toContain("version=0f0f0f0f0f0f0f0f");
    // The art of the commander leads the page; its credit is the art's own.
    expect(r.container.querySelector(".band img")?.getAttribute("src")).toContain(
      "22222222-2222-4222-8222-222222222222",
    );
  });

  it("changing the filters is one read", async () => {
    const r = render(StatsDeck, { deckId: DECK });
    await settle();
    const before = calls.filter((u) => u.startsWith("/me/stats/decks/")).length;
    [...r.container.querySelectorAll<HTMLInputElement>("input[name=stats-range]")]
      .find((i) => i.value === "30d")!
      .click();
    await settle();
    const reads = calls.filter((u) => u.startsWith("/me/stats/decks/"));
    expect(reads.length).toBe(before + 1);
    expect(reads.at(-1)).toContain("range=30d");
  });

  it("offers All time when nothing was played since the last edit", async () => {
    const d = deckFixture();
    d.no_games_since_edit = true;
    d.record = { games: 0, wins: 0, losses: 0, draws: 0, avg_place: null };
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = typeof input === "string" ? input : input.toString();
        calls.push(url);
        return url.includes("version=all")
          ? json(200, { ...deckFixture(), version: "all" })
          : json(200, d);
      }),
    );
    const r = render(StatsDeck, { deckId: DECK });
    await settle();
    const note = r.container.querySelector("[data-testid=deck-no-games-since-edit]");
    expect(note?.textContent).toMatch(/No games since your last edit on/);
    const btn = [...note!.querySelectorAll("button")].find(
      (b) => b.textContent?.trim() === L.statsAllDeckGames,
    );
    btn!.click();
    await settle();
    expect(calls.at(-1)).toContain("version=all");
    expect(r.container.querySelector("[data-testid=deck-no-games-since-edit]")).toBeNull();
  });

  it("offers the switch, never an error, while advanced stats are off", async () => {
    prefs = { ...prefs, advanced: false, choice: "off" };
    statsPrefs.set(prefs);
    deckStatus = 403;
    const r = render(StatsDeck, { deckId: DECK });
    await settle();
    expect(r.container.querySelector("[data-testid=advanced-stats-opt-in]")).not.toBeNull();
    expect(r.container.querySelector("[role=alert]")).toBeNull();
    expect(r.container.querySelector("[data-testid=deck-stats]")).toBeNull();
  });

  it("says so for a deck that is not one's own", async () => {
    deckStatus = 404;
    const r = render(StatsDeck, { deckId: DECK });
    await settle();
    expect(r.container.querySelector("[data-testid=deck-stats-gone]")?.textContent).toMatch(
      /no stats page/,
    );
  });
});

describe("the shared deck page", () => {
  it("is read-only: no sign-in, the owner's name, no art, and a way back to their stats", async () => {
    session.set(null);
    statsPrefs.set(STATS_PREFS_OFF);
    const r = render(StatsDeck, { deckId: DECK, token: "tok" });
    await settle();
    expect(calls.filter((u) => u.startsWith("/stats/") || u.startsWith("/me/"))).toEqual([
      `/stats/shared/tok/decks/${DECK}?makeup=all&range=all&version=current`,
    ]);
    expect(r.container.textContent).toContain("Bea's deck stats");
    expect(r.container.querySelector(".band")).toBeNull();
    expect(r.container.querySelectorAll(".card-grid img")).toHaveLength(0);
    expect(
      r.container
        .querySelector<HTMLAnchorElement>("[data-testid=deck-stats-back]")
        ?.getAttribute("href"),
    ).toBe("#/stats/shared/tok");
  });
});

describe("the overview's deck rows", () => {
  it("link to the deck page, or to the shared deck page in a shared view", () => {
    const own = render(AdvancedOverview, { overview: overviewFixture() });
    flushSync();
    const link = own.container.querySelector<HTMLAnchorElement>("table.decks th.name a")!;
    expect(link.getAttribute("href")).toBe(`#/stats/deck/${DECK}`);
    // A deck with no saved id is not a link.
    expect(own.container.querySelectorAll("table.decks th.name a")).toHaveLength(1);
    cleanup();
    const shared = render(AdvancedOverview, {
      overview: overviewFixture(),
      shared: true,
      token: "tok",
    });
    flushSync();
    expect(
      shared.container
        .querySelector<HTMLAnchorElement>("table.decks th.name a")
        ?.getAttribute("href"),
    ).toBe(`#/stats/shared/tok/deck/${DECK}`);
  });
});
