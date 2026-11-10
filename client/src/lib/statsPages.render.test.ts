// @vitest-environment jsdom
//
// statsPages.render.test.ts — ADR 0217 slice 6 on the client: the
// advanced overview's charts and their text equivalents, the share
// dialog, and #/stats with the setting off and on. The routes are a
// stubbed fetch.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

import AdvancedOverview from "./components/stats/AdvancedOverview.svelte";
import ShareStatsDialog from "./components/stats/ShareStatsDialog.svelte";
import Stats from "../routes/Stats.svelte";
import { L } from "./labels";
import { STATS_PREFS_OFF, statsPrefs, type StatsPrefs } from "./statsPrefs";
import { session, type Session } from "./session";
import { cleanup, click, render } from "./test/render.svelte";
import { overviewFixture } from "./test/advancedStatsFixture";

const person = (): Session =>
  ({
    token: "tok",
    expiresAt: "2999-01-01T00:00:00Z",
    principal: { role: "player", user_id: "u1", game_id: "g1", player_id: "p1" },
  }) as unknown as Session;

let prefs: StatsPrefs;
let calls: { url: string; method: string }[];

function json(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

const basic = {
  games: 14,
  wins: 5,
  losses: 9,
  draws: 0,
  avg_place: 2.2,
  unfinished: 0,
  by_deck: [{ name: "Atraxa", games: 14, wins: 5, losses: 9, draws: 0, avg_place: 2.2 }],
  by_commander: [],
  recent: [],
  top_cards: [],
};

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

beforeEach(() => {
  session.set(person());
  prefs = {
    available: true,
    advanced: false,
    choice: "default",
    default: false,
    share_pods: false,
  };
  calls = [];
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = typeof input === "string" ? input : input.toString();
      const method = init?.method ?? "GET";
      calls.push({ url, method });
      if (url.startsWith("/me/stats/prefs")) {
        if (method === "PUT") {
          const on = JSON.parse(String(init?.body)).advanced === true;
          prefs = { ...prefs, advanced: on, choice: on ? "on" : "off", set_by: "self" };
        }
        return json(200, prefs);
      }
      if (url.startsWith("/me/stats/advanced")) {
        return prefs.advanced ? json(200, overviewFixture()) : json(403, { error: "off" });
      }
      if (url.startsWith("/me/stats/share-link")) {
        if (method === "POST")
          return json(201, { token: "T".repeat(22), path: "#/stats/shared/x", created_at: 1000 });
        return new Response(null, { status: 204 });
      }
      if (url === "/me/stats") return json(200, { ...basic, advanced: prefs.advanced });
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

describe("the advanced overview", () => {
  it("names every chart by a caption that exists, and offers each as a table", () => {
    const r = render(AdvancedOverview, { overview: overviewFixture() });
    flushSync();
    const imgs = [...r.container.querySelectorAll("[role=img][aria-labelledby]")];
    expect(imgs.length).toBeGreaterThan(6);
    for (const img of imgs) {
      const cap = r.container.querySelector(`#${img.getAttribute("aria-labelledby")}`);
      expect(cap?.tagName, img.outerHTML.slice(0, 80)).toBe("FIGCAPTION");
      expect(cap?.textContent?.trim().length).toBeGreaterThan(10);
    }
    const summaries = [...r.container.querySelectorAll("summary")].map((s) =>
      s.textContent?.trim(),
    );
    expect(summaries.length).toBeGreaterThan(5);
    expect(new Set(summaries)).toEqual(new Set([L.showAsTable]));
    // Each section in ADR 0217 §8's order.
    const heads = [...r.container.querySelectorAll("h2")].map((h) => h.textContent);
    expect(heads).toEqual(["Results", "Play", "Mana", "Cards", "Colours", "Decks"]);
    // A rate shows its count beside it.
    expect(r.container.textContent).toContain("36%, 5 of 14 games");
    // The turn-order bar on 2 games has no rate.
    expect(r.container.textContent).toContain("2 games");
  });

  it("draws no card image in a shared view", () => {
    const o = overviewFixture();
    o.most_cast[0].scryfall_id = "11111111-1111-4111-8111-111111111111";
    const own = render(AdvancedOverview, { overview: o });
    flushSync();
    expect(own.container.querySelectorAll(".card-grid img")).toHaveLength(1);
    cleanup();
    const shared = render(AdvancedOverview, { overview: o, shared: true });
    flushSync();
    expect(shared.container.querySelectorAll(".card-grid img")).toHaveLength(0);
    expect(shared.container.textContent).toContain("Sol Ring");
  });
});

describe("the share dialog", () => {
  it("makes a link, shows it once, and revokes it", async () => {
    statsPrefs.set(prefs);
    const r = render(ShareStatsDialog, { onclose: () => {} });
    flushSync();
    const dialog = r.container.querySelector("[role=dialog]");
    expect(dialog?.getAttribute("aria-label")).toBe(L.shareStatsDialog);
    expect(dialog?.textContent).toContain("never shows the other players");
    const make = [...r.container.querySelectorAll("button")].find(
      (b) => b.textContent?.trim() === L.makeShareLink,
    );
    click(make!);
    await settle();
    const input = r.container.querySelector<HTMLInputElement>("[data-testid=share-link-url]");
    expect(input?.value).toMatch(/\/#\/stats\/shared\/T{22}$/);
    const revoke = [...r.container.querySelectorAll("button")].find(
      (b) => b.textContent?.trim() === L.revokeShareLink,
    );
    click(revoke!);
    await settle();
    expect(calls.map((c) => c.method)).toEqual(["POST", "DELETE"]);
    expect(r.container.querySelector("[data-testid=share-link-url]")).toBeNull();
  });
});

describe("#/stats", () => {
  it("explains advanced stats and offers the switch while they are off, never an error", async () => {
    const r = render(Stats, {});
    await settle();
    expect(r.container.querySelector("[data-testid=advanced-stats-opt-in]")).not.toBeNull();
    expect(r.container.querySelector("[role=alert]")).toBeNull();
    expect(r.container.querySelector("[data-testid=advanced-overview]")).toBeNull();
    expect(calls.some((c) => c.url.startsWith("/me/stats/advanced"))).toBe(false);
  });

  it("shows the overview once they are on, with the switch at the head", async () => {
    prefs = { ...prefs, advanced: true, choice: "on", set_by: "self" };
    const r = render(Stats, {});
    await settle();
    expect(r.container.querySelector("[data-testid=advanced-overview]")).not.toBeNull();
    const sw = r.container.querySelector<HTMLInputElement>("[data-testid=advanced-stats-switch]");
    expect(sw?.checked).toBe(true);
    expect(r.container.textContent).toContain(L.statsMakeupFilter);
    // A filter change is one read.
    const before = calls.filter((c) => c.url.startsWith("/me/stats/advanced")).length;
    const people = [
      ...r.container.querySelectorAll<HTMLInputElement>("input[name=stats-makeup]"),
    ].find((i) => i.value === "people");
    people!.click();
    await settle();
    const reads = calls.filter((c) => c.url.startsWith("/me/stats/advanced"));
    expect(reads.length).toBe(before + 1);
    expect(reads.at(-1)?.url).toContain("makeup=people");
    expect(r.container.querySelector("[role=status][aria-live=polite]")?.textContent).toContain(
      L.statsMakeupPeople.toLowerCase(),
    );
  });
});
