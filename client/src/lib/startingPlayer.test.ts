import { describe, expect, it } from "vitest";
import { firstTurnHeadline, openingRollText, openingRollWinner } from "./startingPlayer";
import type { LogEvent } from "./protocol";

const seats = [
  { seat: 0, name: "A" },
  { seat: 1, name: "B" },
  { seat: 2, name: "C" },
];

// ADR 0121 §3: the opening roll's lines, as the server writes them.
const rolled: LogEvent[] = [
  { seq: 1, kind: "roll", seat: 0, text: "A rolled a d20: 20", sides: 20, results: [20] },
  { seq: 2, kind: "roll", seat: 1, text: "B rolled a d20: 4", sides: 20, results: [4] },
  { seq: 3, kind: "roll", seat: 2, text: "C rolled a d20: 20", sides: 20, results: [20] },
  {
    seq: 4,
    kind: "opening_roll",
    seat: -1,
    label: "tie",
    seats: [0, 2],
    results: [20],
    text: "A and C tied with 20 and roll again",
  },
  { seq: 5, kind: "roll", seat: 0, text: "A rolled a d20: 7", sides: 20, results: [7] },
  { seq: 6, kind: "roll", seat: 2, text: "C rolled a d20: 18", sides: 20, results: [18] },
  {
    seq: 7,
    kind: "opening_roll",
    seat: 2,
    label: "won",
    results: [18],
    text: "C won the opening roll with 18",
  },
];

describe("openingRollWinner", () => {
  it("reads the starting_player entry and the winning roll", () => {
    const log: LogEvent[] = [
      ...rolled,
      {
        seq: 8,
        kind: "starting_player",
        seat: 2,
        target_seat: 2,
        text: "C chose to take the first turn",
      },
    ];
    const winner = openingRollWinner({ starting_seat: 2, seats, log });
    expect(winner).toEqual({ seat: 2, name: "C", result: 18 });
    expect(openingRollText(winner!)).toBe("C won the d20 roll with 18 and goes first");
  });

  it("names the seat the winner chose when it was someone else", () => {
    const log: LogEvent[] = [
      ...rolled,
      {
        seq: 8,
        kind: "starting_player",
        seat: 2,
        target_seat: 1,
        text: "C chose B to take the first turn",
      },
    ];
    const winner = openingRollWinner({ starting_seat: 1, seats, log });
    expect(winner).toEqual({ seat: 2, name: "C", result: 18, chosen: { seat: 1, name: "B" } });
    expect(openingRollText(winner!)).toBe("C won the d20 roll with 18 and chose B to go first");
  });

  it("has no winner while the roll is open, whatever starting_seat reads", () => {
    expect(
      openingRollWinner({
        starting_seat: 0,
        opening_roll: { rounds: [{ seats: [0, 1, 2], rolls: [] }] },
        seats,
        log: rolled.slice(0, 3),
      }),
    ).toBeNull();
  });

  it("never takes a bare d20 for the opening roll", () => {
    // A source-less d20 with no starting_player entry names nobody's
    // roll: the starting seat goes first, with no result.
    const log: LogEvent[] = [
      { seq: 1, kind: "roll", seat: 1, text: "B rolled a d20: 20", sides: 20, results: [20] },
    ];
    const winner = openingRollWinner({ starting_seat: 0, seats, log });
    expect(winner).toEqual({ seat: 0, name: "A" });
    expect(openingRollText(winner!)).toBe("A goes first");
  });

  it("still names the first player when a table never rolled (a replay from before the roll)", () => {
    const winner = openingRollWinner({ starting_seat: 1, seats });
    expect(winner).toEqual({ seat: 1, name: "B" });
    expect(openingRollText(winner!)).toBe("B goes first");
  });
});

describe("firstTurnHeadline", () => {
  it("names the seat that goes first, and says You for the viewer", () => {
    expect(firstTurnHeadline({ seat: 1, name: "Bob", result: 18 }, 0)).toBe("Bob goes first");
    expect(firstTurnHeadline({ seat: 1, name: "Bob", result: 18 }, 1)).toBe("You go first");
    expect(firstTurnHeadline({ seat: 1, name: "Bob" })).toBe("Bob goes first");
  });

  it("reads the chosen seat when the winner gave the first turn away", () => {
    const gave = { seat: 1, name: "Bob", result: 18, chosen: { seat: 2, name: "Carol" } };
    expect(firstTurnHeadline(gave, 0)).toBe("Carol goes first");
    expect(firstTurnHeadline(gave, 2)).toBe("You go first");
    expect(firstTurnHeadline(gave, 1)).toBe("Carol goes first");
  });
});
