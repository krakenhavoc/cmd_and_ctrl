import { describe, it, expect } from "vitest";
import {
  catalogImageURL,
  filterCatalog,
  matchesColor,
  matchesQuery,
  toggle,
  EMPTY_FILTER,
  type CatalogEntry,
  type CatalogFilter,
} from "./catalog";

function card(over: Partial<CatalogEntry> = {}): CatalogEntry {
  return {
    oracle_id: "o-1",
    scryfall_id: "s-1",
    name: "Lightning Bolt",
    type_line: "Instant",
    oracle_text: "Lightning Bolt deals 3 damage to any target.",
    types: ["instant"],
    color_identity: ["R"],
    completeness: "full",
    ...over,
  };
}

function filter(over: Partial<CatalogFilter> = {}): CatalogFilter {
  return { ...EMPTY_FILTER, ...over };
}

describe("matchesColor", () => {
  it("matches on colour identity, not the card's own colours", () => {
    // A Signet's `colors` is empty but its identity is not — filtering
    // on colours would file every mana rock under colourless, which is
    // not how a Commander player thinks about them.
    const signet = card({ name: "Azorius Signet", colors: [], color_identity: ["W", "U"] });
    expect(matchesColor(signet, "W")).toBe(true);
    expect(matchesColor(signet, "U")).toBe(true);
    expect(matchesColor(signet, "C")).toBe(false);
  });

  it("treats C as 'no colour identity at all'", () => {
    expect(matchesColor(card({ color_identity: [] }), "C")).toBe(true);
    expect(matchesColor(card({ color_identity: undefined }), "C")).toBe(true);
    expect(matchesColor(card({ color_identity: ["R"] }), "C")).toBe(false);
  });
});

describe("matchesQuery", () => {
  it("is case-insensitive and matches name, type line and text", () => {
    expect(matchesQuery(card(), "BOLT")).toBe(true);
    expect(matchesQuery(card(), "instant")).toBe(true);
    expect(matchesQuery(card(), "3 damage")).toBe(true);
    expect(matchesQuery(card(), "counterspell")).toBe(false);
  });

  it("matches a face name the card's own name does not contain", () => {
    // Searching the land half of a modal double-faced card has to
    // find it; its printed name leads with the creature.
    const mdfc = card({
      name: "Akoum Warrior // Akoum Teeth",
      faces: [
        { index: 0, name: "Akoum Warrior", automated: false },
        { index: 1, name: "Akoum Teeth", type_line: "Land", automated: true },
      ],
    });
    expect(matchesQuery(mdfc, "akoum teeth")).toBe(true);
  });

  it("an empty or whitespace query matches everything", () => {
    expect(matchesQuery(card(), "")).toBe(true);
    expect(matchesQuery(card(), "   ")).toBe(true);
  });
});

describe("filterCatalog", () => {
  const bolt = card();
  const wrath = card({
    oracle_id: "o-2",
    name: "Wrath of God",
    type_line: "Sorcery",
    types: ["sorcery"],
    color_identity: ["W"],
    completeness: "caveats",
    caveats: ["Regeneration is not prevented."],
  });
  const solRing = card({
    oracle_id: "o-3",
    name: "Sol Ring",
    type_line: "Artifact",
    types: ["artifact"],
    color_identity: [],
    completeness: "unreviewed",
  });
  const all = [bolt, wrath, solRing];

  it("returns everything with an empty filter", () => {
    expect(filterCatalog(all, filter())).toHaveLength(3);
  });

  it("ORs selections within a facet", () => {
    const got = filterCatalog(all, filter({ colors: ["R", "W"] }));
    expect(got.map((c) => c.name)).toEqual(["Lightning Bolt", "Wrath of God"]);
  });

  it("ANDs across facets", () => {
    // "White cards that fully work" — Wrath is white but has caveats,
    // so nothing comes back.
    expect(filterCatalog(all, filter({ colors: ["W"], completeness: ["full"] }))).toHaveLength(0);
    expect(filterCatalog(all, filter({ colors: ["R"], completeness: ["full"] }))).toHaveLength(1);
  });

  it("filters by completeness, which is the point of the page", () => {
    expect(filterCatalog(all, filter({ completeness: ["full"] })).map((c) => c.name)).toEqual([
      "Lightning Bolt",
    ]);
    expect(filterCatalog(all, filter({ completeness: ["caveats"] })).map((c) => c.name)).toEqual([
      "Wrath of God",
    ]);
    // Unreviewed must be its own bucket. Folding it into "full" is
    // exactly the promise this page must not make.
    expect(filterCatalog(all, filter({ completeness: ["unreviewed"] })).map((c) => c.name)).toEqual(
      ["Sol Ring"],
    );
  });

  it("combines a text query with the facets", () => {
    expect(filterCatalog(all, filter({ query: "sol", types: ["artifact"] }))).toHaveLength(1);
    expect(filterCatalog(all, filter({ query: "sol", types: ["creature"] }))).toHaveLength(0);
  });
});

describe("toggle", () => {
  it("adds then removes", () => {
    expect(toggle<string>([], "W")).toEqual(["W"]);
    expect(toggle(["W"], "U")).toEqual(["W", "U"]);
    expect(toggle(["W", "U"], "W")).toEqual(["U"]);
  });
});

describe("catalogImageURL", () => {
  it("points at the public catalogue route, not the session-gated one", () => {
    expect(catalogImageURL(card())).toBe("/catalog/image/s-1?size=normal");
  });

  it("omits face=0 so it shares the existing cache entries", () => {
    expect(catalogImageURL(card(), "small", 0)).toBe("/catalog/image/s-1?size=small");
    expect(catalogImageURL(card(), "normal", 1)).toBe("/catalog/image/s-1?size=normal&face=1");
  });

  it("returns null when the dump has no printing for the card", () => {
    expect(catalogImageURL(card({ scryfall_id: undefined, missing: true }))).toBeNull();
  });
});
