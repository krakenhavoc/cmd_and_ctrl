import { describe, expect, it } from "vitest";

import { avgPlaceText, placeText, recordText, resultText, winRateText } from "./myStats";
import { routeRedirect } from "./signedInHome";
import { parseHash } from "./router";

describe("myStats", () => {
  it("writes a record, with draws only when there are some", () => {
    expect(recordText({ wins: 4, losses: 7, draws: 1 })).toBe("4-7-1");
    expect(recordText({ wins: 4, losses: 7, draws: 0 })).toBe("4-7");
  });

  it("rounds a win rate and has none without games", () => {
    expect(winRateText({ games: 3, wins: 1 })).toBe("33%");
    expect(winRateText({ games: 0, wins: 0 })).toBe("-");
  });

  it("formats average and exact places", () => {
    expect(avgPlaceText(2.44)).toBe("2.4");
    expect(avgPlaceText(null)).toBe("-");
    expect([1, 2, 3, 4, 11, 12, 13, 21].map(placeText)).toEqual([
      "1st",
      "2nd",
      "3rd",
      "4th",
      "11th",
      "12th",
      "13th",
      "21st",
    ]);
    expect(placeText(null)).toBe("-");
    expect(resultText("draw")).toBe("Draw");
  });

  it("routes #/stats, and sends a signed-out visitor to sign in", () => {
    expect(parseHash("#/stats")).toEqual({ name: "stats" });
    expect(routeRedirect({ name: "stats" }, null)).toBe("#/login");
  });
});

describe("the shared stats route (ADR 0217 §6)", () => {
  it("parses #/stats/shared/<token>, and #/stats stays the overview", () => {
    expect(parseHash("#/stats/shared/AbC_-123")).toEqual({
      name: "statsShared",
      token: "AbC_-123",
    });
    expect(parseHash("#/stats/shared")).toEqual({ name: "stats" });
    expect(parseHash("#/stats")).toEqual({ name: "stats" });
  });

  it("is public: a signed-out visitor is not sent to sign in", () => {
    expect(routeRedirect({ name: "statsShared", token: "t" }, null)).toBeNull();
    expect(routeRedirect({ name: "stats" }, null)).toBe("#/login");
    expect(routeRedirect({ name: "statsSharedDeck", token: "t", id: "d" }, null)).toBeNull();
    expect(routeRedirect({ name: "statsDeck", id: "d" }, null)).toBe("#/login");
  });
});

describe("the deck stats routes (ADR 0217 §8)", () => {
  it("parses #/stats/deck/<id> and the shared deck page", () => {
    expect(parseHash("#/stats/deck/3f2a-9")).toEqual({ name: "statsDeck", id: "3f2a-9" });
    expect(parseHash("#/stats/deck")).toEqual({ name: "stats" });
    expect(parseHash("#/stats/shared/AbC_-123/deck/3f2a-9")).toEqual({
      name: "statsSharedDeck",
      token: "AbC_-123",
      id: "3f2a-9",
    });
    // A shared view with no deck after "deck" is still the shared overview.
    expect(parseHash("#/stats/shared/AbC_-123/deck")).toEqual({
      name: "statsShared",
      token: "AbC_-123",
    });
  });
});
