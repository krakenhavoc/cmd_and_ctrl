import { describe, expect, it } from "vitest";
import {
  keepCostText,
  mulliganCostText,
  mulliganDecider,
  mulliganWaitingText,
  nextMulliganBottom,
  freeMulliganCount,
} from "./mulliganTurn";

describe("the London mulligan's cost (ADR 0168)", () => {
  it("says what a keep costs, and nothing when it is free", () => {
    expect(keepCostText(0)).toBeNull();
    expect(keepCostText(1)).toBe("Keep: put 1 card from this hand on the bottom of your library.");
    expect(keepCostText(2)).toBe("Keep: put 2 cards from this hand on the bottom of your library.");
  });

  it("counts the table's free mulligans: the rules' one, or the setting when it gives more", () => {
    expect(freeMulliganCount(2, undefined)).toBe(0);
    expect(freeMulliganCount(2, 0)).toBe(0);
    expect(freeMulliganCount(4, 0)).toBe(1);
    expect(freeMulliganCount(2, 1)).toBe(1);
    expect(freeMulliganCount(4, 3)).toBe(3);
    expect(freeMulliganCount(2, -1)).toBe(Infinity);
  });

  it("makes a mulligan free while the seat has free ones left", () => {
    expect(nextMulliganBottom(0, 0, 1)).toBe(0);
    expect(nextMulliganBottom(0, 0, 0)).toBe(1);
    expect(nextMulliganBottom(1, 0, 1)).toBe(1);
    expect(nextMulliganBottom(2, 1, 1)).toBe(2);
    expect(nextMulliganBottom(1, 1, 0)).toBe(2);
    expect(nextMulliganBottom(2, 0, 3)).toBe(0);
    expect(nextMulliganBottom(3, 0, 3)).toBe(1);
    expect(nextMulliganBottom(10, 0, Infinity)).toBe(0);
  });

  it("says how many free mulligans are left, or that they are unlimited", () => {
    expect(mulliganCostText(1, 0, 3)).toBe(
      "Mulligan: free (2 free mulligans left). Shuffle this hand away and draw 7.",
    );
    expect(mulliganCostText(2, 0, 3)).toBe(
      "Mulligan: free (1 free mulligan left). Shuffle this hand away and draw 7.",
    );
    expect(mulliganCostText(5, 0, Infinity)).toBe(
      "Mulligan: mulligans are free at this table. Shuffle this hand away and draw 7.",
    );
  });

  it("says what one more mulligan costs", () => {
    expect(mulliganCostText(0, 0, 1)).toBe(
      "Mulligan: your first mulligan is free. Shuffle this hand away and draw 7.",
    );
    expect(mulliganCostText(1, 0, 1)).toBe(
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
