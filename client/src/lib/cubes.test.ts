// @vitest-environment jsdom
//
// #5441 (ADR 0222 §8): a signed-in person's saved cubes.

import { afterEach, describe, expect, it, vi } from "vitest";

import {
  createCube,
  cubeBlockerCount,
  cubeFits,
  cubeNeed,
  cubeStatus,
  cubeTooSmallText,
  CUBE_LIMITS,
  deleteCube,
  dropCards,
  fetchMyCubes,
  MY_CUBES_HASH,
  packSizeChoices,
  updateCube,
  type CubeCheck,
} from "./cubes";
import { draftNewCubeHash } from "./draft";
import { parseHash } from "./router";
import { isPublicRoute } from "./signedInHome";

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

const check = (over: Partial<CubeCheck> = {}): CubeCheck => ({
  cards: 360,
  unknown: [],
  not_full: [],
  repeated: [],
  too_large: false,
  ready: true,
  max_seats: 8,
  ...over,
});

function serve(status: number, body: unknown): { url: string; init: RequestInit }[] {
  const asked: { url: string; init: RequestInit }[] = [];
  globalThis.fetch = vi.fn(async (input: RequestInfo | URL, init: RequestInit = {}) => {
    asked.push({ url: String(input), init });
    return new Response(status === 204 ? null : JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    });
  }) as never;
  return asked;
}

describe("the size a draft needs", () => {
  it("is seats x 3 packs x pack size", () => {
    expect(cubeNeed(8, 15)).toBe(360);
    expect(cubeNeed(4, 15)).toBe(180);
    expect(cubeNeed(2, 9)).toBe(54);
  });

  it("fits a ready cube that is large enough, and no cube that is not ready", () => {
    expect(cubeFits({ check: check({ cards: 180 }) }, 4, 15)).toBe(true);
    expect(cubeFits({ check: check({ cards: 179 }) }, 4, 15)).toBe(false);
    expect(cubeFits({ check: check({ ready: false, not_full: ["X"] }) }, 2, 9)).toBe(false);
  });

  it("says how many cards the draft needs when the cube is too small", () => {
    const c = { name: "Pauper", check: check({ cards: 200 }) };
    expect(cubeTooSmallText(c, 4, 15)).toBe("");
    expect(cubeTooSmallText(c, 8, 15)).toBe(
      "Pauper has 200 cards. 8 seats with packs of 15 need 360: choose fewer seats or smaller packs.",
    );
  });

  it("offers every pack size the server takes", () => {
    const sizes = packSizeChoices(CUBE_LIMITS);
    expect(sizes[0]).toBe(9);
    expect(sizes[sizes.length - 1]).toBe(20);
    expect(sizes).toContain(15);
  });
});

describe("a cube's state in words", () => {
  it("counts each blocking card once and names the state", () => {
    const blocked = check({ ready: false, unknown: ["Zzz"], not_full: ["A", "B"], max_seats: 8 });
    expect(cubeBlockerCount(blocked)).toBe(3);
    expect(cubeStatus(blocked)).toBe("3 cards block it");
    expect(cubeStatus(check({ ready: false, repeated: ["A"] }))).toBe("1 card blocks it");
    expect(cubeStatus(check())).toBe("Ready for up to 8 seats");
    expect(cubeStatus(check({ cards: 40, max_seats: 0 }))).toBe("Too small to draft");
    expect(cubeStatus(check({ cards: 0, ready: false, max_seats: 0 }))).toBe("No cards");
    expect(cubeStatus(check({ cards: 2000, ready: false, too_large: true }))).toBe(
      "Too many cards",
    );
  });
});

describe("dropping the cards that block", () => {
  it("removes the named rows, with or without a printing, and keeps the rest", () => {
    const list = "1 Lightning Bolt\n2 Not Built (BLB) 12\n1 Counterspell (DMU) 45a\n1 Typo Card\n";
    expect(dropCards(list, ["not built", "Typo Card"])).toBe(
      "1 Lightning Bolt\n1 Counterspell (DMU) 45a",
    );
  });

  it("does not take a card whose name only starts the same", () => {
    expect(dropCards("1 Fire\n1 Fire // Ice\n", ["Fire"])).toBe("1 Fire // Ice");
  });
});

describe("the routes", () => {
  it("reads the list with the limits beside it", async () => {
    const asked = serve(200, { cubes: [], max_cubes: 20, max_cards: 1080, pack_size: 15 });
    const r = await fetchMyCubes();
    expect(asked[0].url).toBe("/me/cubes");
    expect(r.cubes).toEqual([]);
    expect(r.max_cards).toBe(1080);
    expect(r.min_pack_size).toBe(CUBE_LIMITS.min_pack_size);
  });

  it("creates from a pasted list or a link, never sending an empty field", async () => {
    const asked = serve(201, {
      cube: { id: "c1", name: "Pauper", singleton: true, check: check() },
    });
    await createCube({ name: " Pauper ", text: "1 Lightning Bolt\n", singleton: true });
    expect(asked[0].init.method).toBe("POST");
    expect(JSON.parse(String(asked[0].init.body))).toEqual({
      name: "Pauper",
      text: "1 Lightning Bolt",
      singleton: true,
    });
    await createCube({ name: "", url: "https://archidekt.com/decks/1", singleton: false });
    expect(JSON.parse(String(asked[1].init.body))).toEqual({
      url: "https://archidekt.com/decks/1",
      singleton: false,
    });
  });

  it("changes and deletes one cube by its id", async () => {
    const asked = serve(200, {
      cube: { id: "c 1", name: "Pauper", singleton: false, check: check() },
    });
    await updateCube("c 1", { singleton: false });
    expect(asked[0].url).toBe("/me/cubes/c%201");
    expect(asked[0].init.method).toBe("PUT");
    expect(JSON.parse(String(asked[0].init.body))).toEqual({ singleton: false });
    const del = serve(204, null);
    await deleteCube("c 1");
    expect(del[0].url).toBe("/me/cubes/c%201");
    expect(del[0].init.method).toBe("DELETE");
  });
});

describe("the hashes", () => {
  it("opens My cubes at #/cubes, for a signed-in person only", () => {
    expect(parseHash(MY_CUBES_HASH)).toEqual({ name: "myCubes" });
    expect(isPublicRoute({ name: "myCubes" })).toBe(false);
  });

  it("opens the create form with a cube chosen", () => {
    expect(draftNewCubeHash("abc")).toBe("#/draft?cube=abc");
    expect(parseHash("#/draft?cube=abc")).toEqual({ name: "draftNew", cube: "abc" });
    // The create form stays public: a visitor drafts a set.
    expect(isPublicRoute({ name: "draftNew", cube: "abc" })).toBe(true);
  });
});
