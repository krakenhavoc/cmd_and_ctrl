import { describe, expect, it } from "vitest";
import {
  keepCostText,
  mulliganCostText,
  mulliganDecider,
  mulliganWaitingText,
  nextMulliganBottom,
} from "./mulliganTurn";

describe("the London mulligan's cost (ADR 0166)", () => {
  it("says what a keep costs, and nothing when it is free", () => {
    expect(keepCostText(0)).toBeNull();
    expect(keepCostText(1)).toBe("Keep: put 1 card from this hand on the bottom of your library.");
    expect(keepCostText(2)).toBe("Keep: put 2 cards from this hand on the bottom of your library.");
  });

  it("makes only the first mulligan free, and only when the table has one", () => {
    expect(nextMulliganBottom(0, 0, true)).toBe(0);
    expect(nextMulliganBottom(0, 0, false)).toBe(1);
    expect(nextMulliganBottom(1, 0, true)).toBe(1);
    expect(nextMulliganBottom(2, 1, true)).toBe(2);
    expect(nextMulliganBottom(1, 1, false)).toBe(2);
  });

  it("says what one more mulligan costs", () => {
    expect(mulliganCostText(0, 0, true)).toBe(
      "Mulligan: your first mulligan is free. Shuffle this hand away and draw 7.",
    );
    expect(mulliganCostText(1, 0, true)).toBe(
      "Mulligan: shuffle this hand away, draw 7, then put 1 card on the bottom when you keep.",
    );
  });
});

const seat = (name: string, extra: object = {}) => ({ name, ...extra });

describe("mulligan turn order", () => {
  const seats = [seat("Ann"), seat("Bo", { mulligan_turn: true }), seat("Cy"), seat("Di")];

  it("finds the one seat whose turn it is", () => {
    expect(mulliganDecider(seats)?.name).toBe("Bo");
  });

  it("finds nobody when no seat is marked", () => {
    expect(mulliganDecider([seat("Ann"), seat("Bo")])).toBeNull();
  });

  it("never names a seat that kept or left", () => {
    expect(mulliganDecider([seat("Ann", { mulligan_turn: true, hand_kept: true })])).toBeNull();
    expect(mulliganDecider([seat("Ann", { mulligan_turn: true, eliminated: true })])).toBeNull();
  });

  it("tells a waiting seat who is deciding", () => {
    expect(mulliganWaitingText(seats, false)).toBe("Waiting for Bo to decide");
  });

  it("tells the deciding seat nothing", () => {
    expect(mulliganWaitingText(seats, true)).toBeNull();
  });

  it("says nothing when nobody is deciding", () => {
    expect(mulliganWaitingText([seat("Ann")], false)).toBeNull();
  });
});
