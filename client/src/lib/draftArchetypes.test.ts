// @vitest-environment jsdom
//
// #5703 (ADR 0222 §7): the colour pairs of a set, as the page reads them.

import { afterEach, describe, expect, it, vi } from "vitest";

import type { DraftCard } from "./draft";
import {
  clearArchetypeCache,
  colourCounts,
  fetchArchetypes,
  PAIR_FIT_MIN,
  pairFit,
  pairHasContent,
  pairName,
  pairTitle,
  railColours,
  RAIL_MIN_PICKS,
  signpostCard,
} from "./draftArchetypes";

const realFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = realFetch;
  clearArchetypeCache();
  vi.restoreAllMocks();
});

function serve(body: unknown, status = 200): ReturnType<typeof vi.fn> {
  const spy = vi.fn(
    async () =>
      new Response(JSON.stringify(body), {
        status,
        headers: { "Content-Type": "application/json" },
      }),
  );
  globalThis.fetch = spy as never;
  return spy;
}

describe("colour pairs", () => {
  it("names a pair in words", () => {
    expect(pairName("WU")).toBe("white-blue");
    expect(pairTitle("BR")).toBe("Black-red");
    expect(pairName("WG")).toBe("white-green");
  });

  it("draws a signpost as a draft card by its printing", () => {
    const c = signpostCard({ id: "abc", oracle_id: "o", name: "Mentor", mana_cost: "{1}{W}{U}" });
    expect(c).toMatchObject({
      id: "abc",
      scryfall_id: "abc",
      name: "Mentor",
      mana_cost: "{1}{W}{U}",
    });
  });

  it("says which pairs have anything to show", () => {
    expect(pairHasContent({ colors: "WU", signposts: [] })).toBe(false);
    expect(pairHasContent({ colors: "WU", label: "Flyers", signposts: [] })).toBe(true);
    expect(pairHasContent({ colors: "WU", signposts: [{ id: "x", name: "N" }] })).toBe(true);
  });
});

describe("fetchArchetypes", () => {
  it("asks the public route once per set, and reads a list from the body", async () => {
    const spy = serve({
      set: "blb",
      name: "Bloomburrow",
      pairs: [{ colors: "WU", label: "Birds and flyers" }],
    });
    const a = await fetchArchetypes("BLB");
    expect(spy).toHaveBeenCalledTimes(1);
    expect(String(spy.mock.calls[0][0])).toBe("/draft/sets/blb/archetypes");
    expect(a.pairs[0]).toMatchObject({ colors: "WU", label: "Birds and flyers", signposts: [] });
    await fetchArchetypes("blb");
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it("refuses a code that is not a set code before it is put in a path", async () => {
    const spy = serve({});
    await expect(fetchArchetypes("../x")).rejects.toThrow();
    await expect(fetchArchetypes("")).rejects.toThrow();
    expect(spy).not.toHaveBeenCalled();
  });

  it("does not keep a failure", async () => {
    serve({ error: "not a draftable set" }, 404);
    await expect(fetchArchetypes("zzz")).rejects.toThrow("not a draftable set");
    const spy = serve({ set: "zzz", name: "Z", pairs: [] });
    await expect(fetchArchetypes("zzz")).resolves.toMatchObject({ set: "zzz" });
    expect(spy).toHaveBeenCalledTimes(1);
  });
});

// ---- The collapsed panel's pips and each pair's fit (#5848 §4) ----

describe("colourCounts / railColours", () => {
  const c = (id: string, colors?: string[]): DraftCard => ({ id, name: id, colors });

  it("counts a gold card once in each of its colours, and a colourless card or a land in none", () => {
    expect(colourCounts([c("a", ["W", "U"]), c("b", ["W"]), c("c", []), c("d")])).toEqual({
      W: 2,
      U: 1,
    });
  });

  it("lists the colours with three or more cards, most first, WUBRG on a tie", () => {
    const cards = [
      c("a", ["G"]),
      c("b", ["G"]),
      c("c", ["G"]),
      c("d", ["W"]),
      c("e", ["W"]),
      c("f", ["W"]),
      c("g", ["B"]),
      c("h", ["B"]),
      c("i", ["B"]),
      c("j", ["B"]),
      c("k", ["R"]),
      c("l", ["R"]),
    ];
    expect(railColours(cards)).toEqual(["B", "W", "G"]);
    expect(RAIL_MIN_PICKS).toBe(3);
  });

  it("is empty with no colour at three", () => {
    expect(railColours([c("a", ["U"]), c("b", ["U"])])).toEqual([]);
    expect(railColours([])).toEqual([]);
  });
});

describe("pairFit", () => {
  const c = (id: string, colors?: string[]): DraftCard => ({ id, name: id, colors });

  it("counts the coloured cards whose colours are all in the pair", () => {
    const cards = [
      c("a", ["W"]),
      c("b", ["U"]),
      c("c", ["W", "U"]),
      c("d", ["W", "B"]),
      c("e", ["B"]),
    ];
    expect(pairFit("WU", cards)).toBe(3);
    expect(pairFit("WB", cards)).toBe(3);
    expect(pairFit("RG", cards)).toBe(0);
  });

  it("leaves out colourless cards and lands, which fit every pair", () => {
    expect(pairFit("WU", [c("a"), c("b", []), c("c", ["U"])])).toBe(1);
    expect(PAIR_FIT_MIN).toBe(3);
  });
});
