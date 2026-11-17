import { describe, expect, it } from "vitest";

import type { DraftCard } from "./draft";
import { ratingFooter, ratingHeading, ratingReasons } from "./draftReasons";
import { L } from "./labels";

const lands = (value: number, over = {}): DraftCard["rating"] => ({
  source: "17lands",
  value,
  tier: "B+",
  games: 12000,
  avg_pick: 4.2,
  avg_seen: 5.1,
  set: "blb",
  format: "PremierDraft",
  fetched: "2026-11-11",
  ...over,
});
const card = (over: Partial<DraftCard>): DraftCard => ({
  id: "x",
  name: "X",
  colors: ["B"],
  mana_value: 2,
  type_line: "Creature — Rat",
  rating: lands(58.1),
  ...over,
});

describe("rating reasons", () => {
  it("heads the box with the rating and its grade", () => {
    expect(ratingHeading(card({}))).toBe("Rating 58.1 (B+)");
    expect(ratingHeading(card({ rating: { source: "c2", value: 3.4 } }))).toBe("Rating 3.4");
    expect(ratingHeading(card({ rating: null }))).toBe("");
  });

  it("explains a removal spell: win rate against the set, and what it does", () => {
    const lines = ratingReasons(
      card({
        type_line: "Instant",
        oracle_text: "Destroy target nonblack creature.",
        rating: lands(60.4),
      }),
      57.2,
    );
    expect(lines).toEqual([
      "Wins 60.4% of games when drawn, +3.2 against this set's average.",
      L.draftReasonRemoval,
    ]);
  });

  it("explains a flier, and says a win rate below the set's with a minus", () => {
    const lines = ratingReasons(
      card({ keywords: ["Flying"], oracle_text: "Flying", rating: lands(54.0) }),
      55.4,
    );
    expect(lines).toEqual([
      "Wins 54.0% of games when drawn, -1.4 against this set's average.",
      L.draftReasonFlier,
    ]);
  });

  it("adds the two-colour line, last", () => {
    const lines = ratingReasons(
      card({
        colors: ["R", "W"],
        type_line: "Instant",
        oracle_text: "Lightning Helix deals 3 damage to any target and you gain 3 life.",
      }),
      55,
    );
    expect(lines).toEqual([
      "Wins 58.1% of games when drawn, +3.1 against this set's average.",
      L.draftReasonRemoval,
      L.draftReasonTwoColours,
    ]);
    expect(ratingReasons(card({ colors: ["W", "U", "B"], type_line: "Sorcery" }), 55)).toContain(
      L.draftReasonManyColours,
    );
  });

  it("without a 17Lands number says the rating is C2's own, and has no win rate", () => {
    const c = card({ rating: { source: "c2", value: 3.4 }, colors: [] });
    const lines = ratingReasons(c, 55);
    expect(lines).toEqual([L.draftReasonInHouse, L.draftReasonPlain]);
    expect(lines.join(" ")).not.toContain("Wins");
  });

  it("gives the win rate alone when the set's average is not known", () => {
    expect(ratingReasons(card({ type_line: "Sorcery", colors: [] }), undefined)).toEqual([
      "Wins 58.1% of games when drawn.",
    ]);
  });

  it("never says how early a card goes (ALSA waits on #5726)", () => {
    const all = [
      ...ratingReasons(card({ rating: lands(60, { avg_seen: 2.1 }) }), 55),
      ...ratingReasons(card({ rating: lands(60, { avg_seen: 9 }) }), 55),
      ...ratingFooter(card({})),
    ].join(" ");
    expect(all).not.toMatch(/gone by pick|comes back around|ALSA/);
  });

  it("keeps to four lines at most", () => {
    expect(
      ratingReasons(card({ keywords: ["Flying"], colors: ["W", "U"] }), 55).length,
    ).toBeLessThanOrEqual(4);
  });
});

describe("the footer", () => {
  it("gives the win rate, the mana value and the source with its date", () => {
    expect(ratingFooter(card({ mana_value: 3 }))).toEqual([
      "GIH WR 58.1%",
      "Mana value 3",
      "Source: 17Lands, 2026-11-11",
    ]);
  });
  it("names the in-house rating as the source when there is no 17Lands data", () => {
    expect(ratingFooter(card({ rating: { source: "c2", value: 3.4 } }))).toEqual([
      "Mana value 2",
      "Source: C2's own rating",
    ]);
  });
});
