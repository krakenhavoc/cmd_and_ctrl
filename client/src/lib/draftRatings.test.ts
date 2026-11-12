// @vitest-environment jsdom
//
// #5703 (ADR 0222 §7): the words and the saved choice behind the rating
// beside a draft card.

import { afterEach, describe, expect, it, vi } from "vitest";

import {
  anyLandsRating,
  anyRating,
  isLandsRating,
  loadShowRatings,
  RATINGS_PREF_KEY,
  ratingAria,
  ratingTierText,
  ratingTooltip,
  ratingValueText,
  saveShowRatings,
  type DraftRating,
} from "./draftRatings";

const lands: DraftRating = {
  source: "17lands",
  value: 58.34,
  tier: "B+",
  games: 12345,
  avg_pick: 6.25,
  avg_seen: 9.1,
  set: "trk",
  format: "PremierDraft",
  fetched: "2026-11-11",
};
const house: DraftRating = { source: "c2", value: 3.4 };

afterEach(() => {
  vi.restoreAllMocks();
  localStorage.clear();
});

describe("the badge's words", () => {
  it("shows a win rate in percent for 17Lands and points for the in-house rating", () => {
    expect(ratingValueText(lands)).toBe("58.3%");
    expect(ratingValueText(house)).toBe("3.4");
    expect(ratingTierText(lands)).toBe("B+");
    // The in-house rating has no grade: it is not on a distribution of win rates.
    expect(ratingTierText(house)).toBe("");
    expect(isLandsRating(lands)).toBe(true);
    expect(isLandsRating(house)).toBe(false);
    expect(isLandsRating(null)).toBe(false);
  });

  it("explains the 17Lands number: what it is, its grade, its sample and where it is from", () => {
    const t = ratingTooltip(lands, "trk");
    expect(t).toContain("17Lands");
    expect(t).toContain("58.3% win rate");
    expect(t).toContain("in hand");
    expect(t).toContain("Grade B+");
    expect(t).toContain("12,345 games");
    expect(t).toContain("Premier Draft");
    expect(t).toContain("fetched 2026-11-11");
    expect(t).toContain("pick 6.3");
    // The card's own set is not repeated; another set is named (a cube card, a reprint).
    expect(t).not.toContain("from TRK");
    expect(ratingTooltip(lands, "fra")).toContain("from TRK");
  });

  it("labels the in-house rating as ours and says it is not a win rate", () => {
    const t = ratingTooltip(house);
    expect(t).toContain("C2 rating 3.4");
    expect(t).toContain("own score");
    expect(t).toContain("not from games played");
    expect(ratingAria(house)).toBe("C2 rating 3.4");
    expect(ratingAria(lands)).toBe("17Lands win rate 58.3 percent, grade B+");
    expect(ratingAria({ ...lands, tier: undefined })).toBe("17Lands win rate 58.3 percent");
  });

  it("knows when the credit is owed", () => {
    expect(anyLandsRating([{ rating: house }, {}])).toBe(false);
    expect(anyLandsRating([{ rating: house }, { rating: lands }])).toBe(true);
    expect(anyRating([{}, { rating: null }])).toBe(false);
    expect(anyRating([{}, { rating: house }])).toBe(true);
  });
});

describe("the viewer's choice", () => {
  it("shows ratings by default and remembers a choice to hide them", () => {
    expect(loadShowRatings()).toBe(true);
    saveShowRatings(false);
    expect(localStorage.getItem(RATINGS_PREF_KEY)).toBe("hide");
    expect(loadShowRatings()).toBe(false);
    saveShowRatings(true);
    expect(localStorage.getItem(RATINGS_PREF_KEY)).toBeNull();
    expect(loadShowRatings()).toBe(true);
  });

  it("reads the same without storage: shown, and a failed save is silent", () => {
    vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("blocked");
    });
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("blocked");
    });
    vi.spyOn(Storage.prototype, "removeItem").mockImplementation(() => {
      throw new Error("blocked");
    });
    expect(loadShowRatings()).toBe(true);
    expect(() => saveShowRatings(false)).not.toThrow();
    expect(() => saveShowRatings(true)).not.toThrow();
  });
});
