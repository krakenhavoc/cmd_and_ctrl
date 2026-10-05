// hints.test.ts — the rules every hint is held to (ADR 0125 §3.2–3.4,
// §8): unique ids in the `<place>.<feature>` form, no retired id reused,
// known places, positive versions, the copy rules (a title of at most 32
// characters, a body of at most 140 ending in a full stop, no engine or
// project vocabulary), and an anchor for each hint over a fixture
// context.
//
// The rules run over every hint the client ships and over the test
// fixtures (lib/test/hints/), which never ship. Hints made to break a
// rule prove that each rule catches what it should.

import { describe, expect, it } from "vitest";

import { HINTS, collectHints } from "./index";
import { anchorOf, placeOfRoute, type Hint } from "./hint";
import { RETIRED_HINT_IDS } from "./retired";
import { BANNED_WORDS, BODY_MAX, TITLE_MAX, collectionProblems, hintProblems } from "./rules";
import { L } from "../labels";
import { FIXTURE_HINTS, hintContexts } from "../test/hintContexts";

const ALL: readonly Hint[] = [...HINTS, ...FIXTURE_HINTS];

const good: Hint = {
  id: "lobby.example",
  version: 1,
  place: "lobby",
  order: 0,
  anchor: { label: L.actions },
  title: "Start a table",
  body: "Name it and create it.",
};

describe("the hint collection", () => {
  it("collects the fixtures through the same glob the client uses", () => {
    expect(FIXTURE_HINTS.map((h) => h.id).sort()).toEqual(["admin.example", "table.example"]);
  });

  it("ships no test fixture", () => {
    const fixtureIDs = new Set(FIXTURE_HINTS.map((h) => h.id));
    expect(HINTS.filter((h) => fixtureIDs.has(h.id))).toEqual([]);
  });

  it("refuses a hint file without a default Hint", () => {
    expect(() => collectHints({ "/src/x/Broken.hint.ts": { default: { id: "lobby.x" } } })).toThrow(
      /Broken\.hint\.ts must have a default export/,
    );
    expect(() => collectHints({ "/src/x/Empty.hint.ts": {} })).toThrow(/Empty\.hint\.ts/);
  });
});

describe("every hint", () => {
  it("has a unique id, never a retired one", () => {
    const problems = collectionProblems(ALL, RETIRED_HINT_IDS);
    expect(problems, problems.join("\n")).toEqual([]);
  });

  it("follows the id, place, version and copy rules", () => {
    const problems = ALL.flatMap(hintProblems);
    expect(problems, problems.join("\n")).toEqual([]);
  });

  it("names an anchor in at least one of its contexts", () => {
    const problems: string[] = [];
    for (const h of ALL) {
      const named = hintContexts(h).some((c) => anchorOf(h, c) !== null);
      if (!named) problems.push(`${h.id}: its anchor names nothing in any fixture context`);
    }
    expect(problems, problems.join("\n")).toEqual([]);
  });
});

describe("the rules catch what they should", () => {
  it("accepts a good hint", () => {
    expect(hintProblems(good)).toEqual([]);
  });

  it("refuses an id not of the form <place>.<feature>, or not of its place", () => {
    expect(hintProblems({ ...good, id: "lobby.Bad_Name" as Hint["id"] })).not.toEqual([]);
    expect(hintProblems({ ...good, id: "lobby" as Hint["id"] })).not.toEqual([]);
    expect(hintProblems({ ...good, id: "decks.example" })).not.toEqual([]);
    expect(hintProblems({ ...good, place: "nowhere" as Hint["place"] })).not.toEqual([]);
  });

  it("refuses a version that is not a positive integer", () => {
    for (const version of [0, -1, 1.5, Number.NaN]) {
      expect(hintProblems({ ...good, version }), String(version)).not.toEqual([]);
    }
  });

  it("refuses a title over 32 characters and a body over 140", () => {
    expect(hintProblems({ ...good, title: "x".repeat(TITLE_MAX) })).toEqual([]);
    expect(hintProblems({ ...good, title: "x".repeat(TITLE_MAX + 1) })).not.toEqual([]);
    expect(hintProblems({ ...good, body: `${"x".repeat(BODY_MAX - 1)}.` })).toEqual([]);
    expect(hintProblems({ ...good, body: `${"x".repeat(BODY_MAX)}.` })).not.toEqual([]);
  });

  it("checks copy made from the key bindings under long rebinds too", () => {
    const body = (k: { nextKey?: string }) =>
      `${"Press it. ".repeat(12)}Then ${k.nextKey ?? ""} again.`;
    const problems = hintProblems({ ...good, body });
    expect(problems.some((p) => p.includes("over 140"))).toBe(true);
  });

  it("refuses a body that does not end in a full stop", () => {
    expect(hintProblems({ ...good, body: "No stop at the end" })).not.toEqual([]);
  });

  it("refuses engine and project vocabulary, as whole words", () => {
    for (const w of BANNED_WORDS) {
      const problems = hintProblems({ ...good, body: `The ${w} keeps this for you.` });
      expect(
        problems.some((p) => p.includes(`"${w}"`)),
        w,
      ).toBe(true);
    }
    // "Engineer" is not "engine"; the admin hint may say "database".
    expect(hintProblems({ ...good, body: "An engineer reads the database." })).toEqual([]);
  });

  it("refuses an id used twice, or a retired id used again", () => {
    expect(collectionProblems([good, good], [])).not.toEqual([]);
    expect(collectionProblems([good], ["lobby.example"])).not.toEqual([]);
  });

  it("refuses an action that leaves the site", () => {
    expect(
      hintProblems({ ...good, action: { label: "Go", href: "https://example.com" } }),
    ).not.toEqual([]);
    expect(
      hintProblems({ ...good, action: { label: "Start practice", href: "#/practice" } }),
    ).toEqual([]);
  });
});

describe("placeOfRoute", () => {
  it("maps the site pages and the table, and leaves the doors without hints", () => {
    expect(placeOfRoute("lobby")).toBe("lobby");
    expect(placeOfRoute("adminViews")).toBe("admin");
    expect(placeOfRoute("myGames")).toBe("my-games");
    expect(placeOfRoute("game")).toBe("table");
    for (const r of ["login", "adminLogin", "join", "reclaim", "practice", "oauthComplete"]) {
      expect(placeOfRoute(r), r).toBeNull();
    }
  });
});
