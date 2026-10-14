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
