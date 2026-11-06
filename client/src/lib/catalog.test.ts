import { describe, it, expect } from "vitest";
import {
  activeChips,
  facetCounts,
  removeChip,
  catalogImageURL,
  filterCatalog,
  findSet,
  matchesColor,
  matchesQuery,
  searchSets,
  setOptionLabel,
  setSummary,
  setYear,
  supportedCount,
  toggle,
  EMPTY_FILTER,
  type CatalogEntry,
  type CatalogFilter,
  type CatalogSet,
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

describe("facetCounts", () => {
  const cards = [
    card({ oracle_id: "a", types: ["instant"], completeness: "full", color_identity: ["R"] }),
    card({ oracle_id: "b", types: ["creature"], completeness: "caveats", color_identity: ["R"] }),
    card({
      oracle_id: "c",
      types: ["artifact", "creature"],
      completeness: "unreviewed",
      color_identity: [],
    }),
  ];

  it("counts every value over the whole catalogue with no filter", () => {
    const n = facetCounts(cards, EMPTY_FILTER);
    expect(n.completeness).toEqual({ full: 1, caveats: 1, unreviewed: 1 });
    expect(n.types.creature).toBe(2);
    expect(n.types.artifact).toBe(1);
    expect(n.types.land).toBe(0);
  });

  it("applies the other facets but not its own", () => {
    const n = facetCounts(cards, filter({ colors: ["R"], types: ["creature"] }));
    // completeness: red creatures only -> b
    expect(n.completeness).toEqual({ full: 0, caveats: 1, unreviewed: 0 });
    // types: red cards -> a (instant), b (creature); own facet ignored
    expect(n.types.instant).toBe(1);
    expect(n.types.creature).toBe(1);
    expect(n.types.artifact).toBe(0);
  });
});

describe("activeChips and removeChip", () => {
  it("makes no chips for an empty filter", () => {
    expect(activeChips(EMPTY_FILTER)).toEqual([]);
  });

  it("makes one chip per active value", () => {
    const f = filter({
      query: " bolt ",
      colors: ["R", "C"],
      types: ["instant"],
      completeness: ["full"],
    });
    expect(activeChips(f).map((c) => `${c.kind}:${c.label}`)).toEqual([
      "query:“bolt”",
      "completeness:Complete",
      "colors:Red",
      "colors:Colourless",
      "types:instant",
    ]);
  });

  it("removes only the chip's own value", () => {
    const f = filter({ query: "bolt", colors: ["R", "U"], types: ["instant"] });
    const chips = activeChips(f);
    const red = chips.find((c) => c.kind === "colors" && c.value === "R")!;
    expect(removeChip(f, red).colors).toEqual(["U"]);
    const q = chips.find((c) => c.kind === "query")!;
    expect(removeChip(f, q).query).toBe("");
    expect(removeChip(f, q).types).toEqual(["instant"]);
  });
});

function setOf(over: Partial<CatalogSet> = {}): CatalogSet {
  return {
    code: "spm",
    name: "Marvel's Spider-Man",
    released: "2026-09-26",
    total: 188,
    full: 118,
    caveats: 2,
    unreviewed: 0,
    unsupported: 68,
    ...over,
  };
}

describe("the set filter", () => {
  const entries = [
    card({ oracle_id: "a", name: "In Spider-Man", sets: ["spm", "old"] }),
    card({ oracle_id: "b", name: "Only Old", sets: ["old"] }),
    card({ oracle_id: "c", name: "In Neither", sets: ["other"] }),
    card({ oracle_id: "d", name: "No Sets At All", sets: undefined }),
  ];

  it("shows every card when no set is picked", () => {
    expect(filterCatalog(entries, filter())).toHaveLength(4);
  });

  it("keeps only the cards printed in the set, whichever of their sets it is", () => {
    expect(filterCatalog(entries, filter({ set: "old" })).map((e) => e.name)).toEqual([
      "In Spider-Man",
      "Only Old",
    ]);
    expect(filterCatalog(entries, filter({ set: "spm" })).map((e) => e.name)).toEqual([
      "In Spider-Man",
    ]);
  });

  it("drops a card with no set list, and every card for an unknown set", () => {
    expect(filterCatalog(entries, filter({ set: "nope" }))).toEqual([]);
    expect(filterCatalog(entries, filter({ set: "other" })).map((e) => e.name)).toEqual([
      "In Neither",
    ]);
  });

  it("ANDs with the other facets", () => {
    const mixed = [
      card({ oracle_id: "a", name: "Red Full", sets: ["spm"], completeness: "full" }),
      card({ oracle_id: "b", name: "Red Unreviewed", sets: ["spm"], completeness: "unreviewed" }),
      card({ oracle_id: "c", name: "Elsewhere", sets: ["old"], completeness: "full" }),
    ];
    const got = filterCatalog(mixed, filter({ set: "spm", completeness: ["full"] }));
    expect(got.map((e) => e.name)).toEqual(["Red Full"]);
  });

  it("is applied to the facet counts, so a count says what ticking it would show", () => {
    const mixed = [
      card({ oracle_id: "a", sets: ["spm"], completeness: "full", types: ["creature"] }),
      card({ oracle_id: "b", sets: ["spm"], completeness: "caveats", types: ["instant"] }),
      card({ oracle_id: "c", sets: ["old"], completeness: "full", types: ["creature"] }),
    ];
    const counts = facetCounts(mixed, filter({ set: "spm" }));
    expect(counts.completeness).toEqual({ full: 1, caveats: 1, unreviewed: 0 });
    expect(counts.types.creature).toBe(1);
  });

  it("gives the set a chip that carries its name and clears with it", () => {
    const f = filter({ set: "spm", query: "web" });
    const chips = activeChips(f, [setOf()]);
    expect(chips.map((c) => `${c.kind}:${c.label}`)).toEqual([
      "set:Marvel's Spider-Man",
      "query:“web”",
    ]);
    const cleared = removeChip(f, chips[0]);
    expect(cleared.set).toBe("");
    expect(cleared.query).toBe("web");
  });

  it("falls back to the code when the set is not in the list", () => {
    expect(activeChips(filter({ set: "spm" }))[0].label).toBe("SPM");
  });
});

describe("set summaries", () => {
  it("states supported of total, the split, and the unsupported rest", () => {
    expect(setSummary(setOf())).toBe(
      "Marvel's Spider-Man: 120 of 188 supported — 118 complete, 2 with caveats; 68 not yet supported",
    );
  });

  it("names unreviewed cards too and leaves out empty buckets", () => {
    expect(
      setSummary(setOf({ full: 10, caveats: 0, unreviewed: 5, total: 20, unsupported: 5 })),
    ).toBe(
      "Marvel's Spider-Man: 15 of 20 supported — 10 complete, 5 unreviewed; 5 not yet supported",
    );
  });

  it("says so plainly when a set is fully supported", () => {
    expect(setSummary(setOf({ total: 10, full: 10, caveats: 0, unsupported: 0 }))).toBe(
      "Marvel's Spider-Man: 10 of 10 supported — 10 complete.",
    );
  });

  it("says so plainly when nothing is supported", () => {
    expect(setSummary(setOf({ total: 188, full: 0, caveats: 0, unsupported: 188 }))).toBe(
      "Marvel's Spider-Man: none of its 188 cards supported yet.",
    );
    expect(setSummary(setOf({ total: 1, full: 0, caveats: 0, unsupported: 1 }))).toBe(
      "Marvel's Spider-Man: its one card is not supported yet.",
    );
  });

  it("labels a picker option name first, then supported over total", () => {
    expect(supportedCount(setOf())).toBe(120);
    expect(setOptionLabel(setOf())).toBe("Marvel's Spider-Man (120/188)");
  });

  it("finds a set by code", () => {
    const sets = [setOf(), setOf({ code: "old", name: "Old" })];
    expect(findSet(sets, "old")?.name).toBe("Old");
    expect(findSet(sets, "zzz")).toBeUndefined();
  });
});

describe("searchSets, the set picker's filter (#4790)", () => {
  function set(code: string, name: string, released: string): CatalogSet {
    return { code, name, released, total: 1, full: 1, caveats: 0, unreviewed: 0, unsupported: 0 };
  }
  // Newest first, as the server sends them.
  const sets = [
    set("tmt", "Teenage Mutant Ninja Turtles", "2026-03-06"),
    set("mh3", "Modern Horizons 3", "2024-06-14"),
    set("mh2", "Modern Horizons 2", "2021-06-18"),
    set("mh1", "Modern Horizons", "2019-06-14"),
    set("csp", "Coldsnap", "2006-07-21"),
    set("hml", "Homelands", "1995-10-01"),
    set("ice", "Ice Age", "1995-06-03"),
  ];
  const codes = (q: string) => searchSets(sets, q).map((s) => s.code);

  it("returns every set in the order given for an empty query", () => {
    expect(codes("")).toEqual(sets.map((s) => s.code));
    expect(codes("   ")).toEqual(sets.map((s) => s.code));
    expect(searchSets(sets, "")).not.toBe(sets);
  });

  it("finds a set by its code", () => {
    expect(codes("hml")).toEqual(["hml"]);
    expect(codes("tmt")).toEqual(["tmt"]);
  });

  it("ignores case in the code and the name", () => {
    expect(codes("MH3")).toEqual(["mh3"]);
    expect(codes("HOMELANDS")).toEqual(["hml"]);
    expect(codes("  homeL ")).toEqual(["hml"]);
  });

  it("ranks an exact code first, then name prefixes, then other names, each newest first", () => {
    // "ice" is Ice Age's code and its name's start; Coldsnap does not contain it.
    expect(codes("ice")).toEqual(["ice"]);
    // "mh" is no set's code or name: the codes that start with it, newest first.
    expect(codes("mh")).toEqual(["mh3", "mh2", "mh1"]);
    // "modern" starts three names.
    expect(codes("modern")).toEqual(["mh3", "mh2", "mh1"]);
    // "horizons" is inside three names.
    expect(codes("horizons")).toEqual(["mh3", "mh2", "mh1"]);
    // "cs": no name has it, Coldsnap's code starts with it.
    expect(codes("cs")).toEqual(["csp"]);
    // "ho": Homelands starts with it, and it is inside the three Horizons names.
    expect(codes("ho")).toEqual(["hml", "mh3", "mh2", "mh1"]);
  });

  it("puts an exact code ahead of a name that starts with the query", () => {
    const withClash = [set("hom", "Homeward", "2020-01-01"), set("ho", "Old Hollow", "1990-01-01")];
    expect(searchSets(withClash, "ho").map((s) => s.code)).toEqual(["ho", "hom"]);
  });

  it("finds nothing for a query that matches no set", () => {
    expect(codes("zzz")).toEqual([]);
  });

  it("gives the year a set came out", () => {
    expect(setYear(set("hml", "Homelands", "1995-10-01"))).toBe("1995");
    expect(setYear({ ...set("x", "X", ""), released: undefined })).toBe("");
  });
});
