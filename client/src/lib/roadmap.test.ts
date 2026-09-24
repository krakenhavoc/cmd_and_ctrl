import { describe, expect, it } from "vitest";

import {
  adrLabel,
  adrURL,
  catalogSearchHash,
  EMPTY_ROADMAP_FILTER,
  filterRoadmap,
  groupByStatus,
  impactLine,
  issueURL,
  matchesRoadmapQuery,
  nextItems,
  type RoadmapItem,
} from "./roadmap";

function item(over: Partial<RoadmapItem> & Pick<RoadmapItem, "slug">): RoadmapItem {
  return {
    name: over.slug,
    kind: "mechanic",
    status: "implemented",
    summary: "",
    ...over,
  };
}

const flying = item({
  slug: "flying",
  name: "Flying",
  kind: "keyword",
  summary: "Can't be blocked except by creatures with flying or reach.",
  examples: [{ name: "Serra Angel", oracle_id: "a" }],
});
const kicker = item({
  slug: "kicker",
  name: "Kicker",
  kind: "mechanic",
  status: "partial",
  summary: "An optional extra cost.",
  missing: "Counters for each time it was kicked are not placed.",
  partial_examples: [{ name: "Everflowing Chalice", caveat: "Enters with no counters." }],
});
const extraCombats = item({
  slug: "extra-combats",
  name: "Extra combats",
  kind: "seam",
  status: "missing",
  summary: "Additional combat phases.",
  missing: "No card can add a combat phase yet.",
  waiting: ["Relentless Assault", "Aggravated Assault"],
  unblocks: 28,
});
const all = [extraCombats, kicker, flying];

describe("filterRoadmap", () => {
  it("returns everything for the empty filter, in order", () => {
    expect(filterRoadmap(all, EMPTY_ROADMAP_FILTER)).toEqual(all);
  });

  it("ORs within a facet and ANDs across facets", () => {
    const f = { query: "", kinds: ["keyword" as const, "seam" as const], statuses: [] };
    expect(filterRoadmap(all, f).map((i) => i.slug)).toEqual(["extra-combats", "flying"]);
    const g = { ...f, statuses: ["missing" as const] };
    expect(filterRoadmap(all, g).map((i) => i.slug)).toEqual(["extra-combats"]);
  });

  it("filters by status alone", () => {
    const f = { query: "", kinds: [], statuses: ["partial" as const, "implemented" as const] };
    expect(filterRoadmap(all, f).map((i) => i.slug)).toEqual(["kicker", "flying"]);
  });

  it("searches names, sentences and every card name, case-insensitively", () => {
    const q = (query: string) =>
      filterRoadmap(all, { ...EMPTY_ROADMAP_FILTER, query }).map((i) => i.slug);
    expect(q("FLYING")).toEqual(["flying"]);
    expect(q("serra")).toEqual(["flying"]);
    expect(q("chalice")).toEqual(["kicker"]);
    expect(q("aggravated")).toEqual(["extra-combats"]);
    expect(q("combat phase")).toEqual(["extra-combats"]);
    expect(q("   ")).toHaveLength(3);
    expect(q("nothing like this")).toEqual([]);
  });

  it("does not match an absent optional field", () => {
    expect(matchesRoadmapQuery(item({ slug: "bare" }), "undefined")).toBe(false);
  });
});

describe("groupByStatus", () => {
  it("puts each item in its section, sorted by name", () => {
    const g = groupByStatus([
      item({ slug: "b", name: "banding", status: "missing" }),
      item({ slug: "z", name: "Ward" }),
      ...all,
      item({ slug: "a", name: "Afflict" }),
    ]);
    expect(g.implemented.map((i) => i.name)).toEqual(["Afflict", "Flying", "Ward"]);
    expect(g.partial.map((i) => i.name)).toEqual(["Kicker"]);
    expect(g.missing.map((i) => i.name)).toEqual(["banding", "Extra combats"]);
  });

  it("always has all three sections", () => {
    expect(groupByStatus([])).toEqual({ implemented: [], partial: [], missing: [] });
  });
});

describe("nextItems", () => {
  it("keeps the server's order and drops unknown slugs", () => {
    const got = nextItems({ next: ["kicker", "gone", "extra-combats"], items: all });
    expect(got.map((i) => i.slug)).toEqual(["kicker", "extra-combats"]);
  });
});

describe("impactLine", () => {
  it("names both measurements, separately", () => {
    expect(impactLine(extraCombats)).toBe("Unblocks 28 cards · 2 waiting");
    expect(impactLine(item({ slug: "x", unblocks: 1 }))).toBe("Unblocks 1 card");
    expect(impactLine(item({ slug: "y", waiting: ["A"] }))).toBe("1 waiting");
    expect(impactLine(flying)).toBe("");
  });
});

describe("links", () => {
  it("points at the public repository", () => {
    expect(issueURL(1386)).toBe("https://github.com/krakenhavoc/cmd_and_ctrl/issues/1386");
    expect(adrURL("0092-public-roadmap-and-site-portal.md")).toBe(
      "https://github.com/krakenhavoc/cmd_and_ctrl/blob/develop/docs/decisions/0092-public-roadmap-and-site-portal.md",
    );
  });

  it("labels an ADR by its number", () => {
    expect(adrLabel("0045-combat-restrictions.md")).toBe("ADR 0045");
    expect(adrLabel("notes.md")).toBe("notes");
  });

  it("builds a catalogue search link the router can read back", () => {
    expect(catalogSearchHash("Borrowing 100,000 Arrows")).toBe(
      "#/catalog?q=Borrowing%20100%2C000%20Arrows",
    );
  });
});
