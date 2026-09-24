import { describe, expect, it } from "vitest";
import { openingRollText, openingRollWinner } from "./startingPlayer";
import type { LogEvent } from "./protocol";

describe("openingRollWinner", () => {
  it("uses the winner's tie-break roll from the pregame public log", () => {
    const log: LogEvent[] = [
      { seq: 1, kind: "roll", seat: 0, text: "A rolled a d20: 20", sides: 20, results: [20] },
      { seq: 2, kind: "roll", seat: 1, text: "B rolled a d20: 4", sides: 20, results: [4] },
      { seq: 3, kind: "roll", seat: 2, text: "C rolled a d20: 20", sides: 20, results: [20] },
      { seq: 4, kind: "roll", seat: 0, text: "A rolled a d20: 7", sides: 20, results: [7] },
      { seq: 5, kind: "roll", seat: 2, text: "C rolled a d20: 18", sides: 20, results: [18] },
    ];
    const winner = openingRollWinner({
      starting_seat: 2,
      seats: [
        { seat: 0, name: "A" },
        { seat: 1, name: "B" },
        { seat: 2, name: "C" },
      ],
      log,
    });
    expect(winner).toEqual({ seat: 2, name: "C", result: 18 });
    expect(openingRollText(winner!)).toBe("C won the d20 roll with 18 and goes first");
  });

  it("still names the first player when an old snapshot has no roll log", () => {
    const winner = openingRollWinner({
      starting_seat: 1,
      seats: [
        { seat: 0, name: "A" },
        { seat: 1, name: "B" },
      ],
    });
    expect(winner).toEqual({ seat: 1, name: "B", result: undefined });
    expect(openingRollText(winner!)).toBe("B goes first");
  });
});
