import { afterEach, describe, expect, it, vi } from "vitest";

import {
  blockedSets,
  boosterName,
  cardsToGo,
  DRAFT_SETS_HASH,
  fetchDraftSets,
  PACKS_ARE_APPROXIMATIONS,
  readySets,
  releaseYear,
  type DraftSet,
} from "./draftSets";
import { L } from "./labels";
import { parseHash } from "./router";
import { isPublicRoute } from "./signedInHome";

function set(code: string, blockers: string[] = [], released = "2026-01-01"): DraftSet {
  return {
    code,
    name: code.toUpperCase(),
    released,
    booster: "play",
    cards: 260,
    draftable: blockers.length === 0,
    blockers,
  };
}

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
  vi.restoreAllMocks();
});

describe("the draftable-set list", () => {
  it("names the booster a set uses, from the label registry", () => {
    expect(boosterName("play")).toBe(L.draftPlayBooster);
    expect(boosterName("draft")).toBe(L.draftDraftBooster);
    expect(L.draftPlayBooster).toBe("Play booster");
    expect(L.draftDraftBooster).toBe("Draft booster");
    // An older server, or a kind this client does not know, shows nothing.
    expect(boosterName(undefined)).toBe("");
    expect(boosterName("collector")).toBe("");
  });

  it("says in the approximation note that both kinds of booster are used", () => {
    expect(PACKS_ARE_APPROXIMATIONS).toContain("play booster");
    expect(PACKS_ARE_APPROXIMATIONS).toContain("draft booster");
  });

  it("splits ready sets from blocked ones, the closest to ready first", () => {
    const sets = [set("new", ["A", "B", "C"]), set("mid"), set("old", ["A"]), set("tie", ["Z"])];
    expect(readySets(sets).map((s) => s.code)).toEqual(["mid"]);
    // old and tie both wait on one card: the server's order settles it.
    expect(blockedSets(sets).map((s) => s.code)).toEqual(["old", "tie", "new"]);
  });

  it("words the count of blocking cards", () => {
    expect(cardsToGo(1)).toBe("1 card to go");
    expect(cardsToGo(12)).toBe("12 cards to go");
  });

  it("reads a release year, and nothing from a missing or odd date", () => {
    expect(releaseYear("2024-08-02")).toBe("2024");
    expect(releaseYear(undefined)).toBe("");
    expect(releaseYear("soon")).toBe("");
  });

  it("says that packs are approximations of retail packs", () => {
    expect(PACKS_ARE_APPROXIMATIONS).toContain("Packs are approximations of retail packs.");
  });
});

describe("the route", () => {
  it("is #/draft/sets, and public", () => {
    expect(DRAFT_SETS_HASH).toBe("#/draft/sets");
    expect(parseHash(DRAFT_SETS_HASH)).toEqual({ name: "draftSets" });
    expect(isPublicRoute({ name: "draftSets" })).toBe(true);
  });

  it("leaves the rest of #/draft for the draft itself (#5438)", () => {
    expect(parseHash("#/draft")).toEqual({ name: "draftNew" });
    expect(parseHash("#/draft/abc")).toEqual({ name: "draft", id: "abc" });
    // "sets" is this page, never a draft's id.
    expect(parseHash("#/draft/sets?invite=x")).toEqual({ name: "draftSets" });
  });
});

describe("fetchDraftSets", () => {
  it("asks the public route without the session's fetch, and reads the list", async () => {
    const spy = vi.fn(
      async () =>
        new Response(
          JSON.stringify({
            sets: [set("aaa"), { ...set("bbb", ["X"]), blockers: null }],
            draftable: 1,
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
    );
    globalThis.fetch = spy as never;
    const got = await fetchDraftSets();
    expect(spy).toHaveBeenCalledTimes(1);
    expect((spy.mock.calls[0] as unknown[])[0]).toBe("/draft/sets");
    expect(got.sets.map((s) => s.code)).toEqual(["aaa", "bbb"]);
    // A missing list is an empty one, never a crash in the page.
    expect(got.sets[1].blockers).toEqual([]);
    expect(got.draftable).toBe(1);
  });

  it("surfaces the server's error", async () => {
    globalThis.fetch = vi.fn(
      async () => new Response(JSON.stringify({ error: "nope" }), { status: 500 }),
    ) as never;
    await expect(fetchDraftSets()).rejects.toThrow("nope");
  });
});
