// @vitest-environment jsdom
//
// #5703 (ADR 0222 §7): the colour pairs of a set, as the page reads them.

import { afterEach, describe, expect, it, vi } from "vitest";

import {
  clearArchetypeCache,
  fetchArchetypes,
  pairHasContent,
  pairName,
  pairTitle,
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
