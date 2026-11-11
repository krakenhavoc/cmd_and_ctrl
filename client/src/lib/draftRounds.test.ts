// #5440 (ADR 0222 §6): reading a draft's rounds.

import { describe, expect, it } from "vitest";

import type { DraftMatch, DraftView } from "./draft";
import {
  matchLengthName,
  matchLine,
  opponentSeat,
  recordText,
  resultFor,
  roundsAnnouncement,
  seatName,
  standingRows,
  yourMatchLine,
  yourMatchState,
} from "./draftRounds";

const open: DraftMatch = { a: 2, b: 0, done: false, wins_a: 0, wins_b: 0 };

function view(over: Partial<DraftView> = {}): DraftView {
  return {
    id: "d",
    name: "Friday draft",
    set: "blb",
    mode: "pick_one",
    seats: 4,
    state: "done",
    seq: 1,
    you: 0,
    players: [
      { seat: 0, name: "Ian" },
      { seat: 1, name: "Bot 1", bot: true },
      { seat: 2, name: "Sam" },
      { seat: 3, name: "Bot 3", bot: true },
    ],
    phase: "rounds",
    best_of: 1,
    play: {
      round: 1,
      rounds: 3,
      matches: [[open, { a: 1, b: 3, done: true, winner: 3, wins_a: 0, wins_b: 1, coin: true }]],
      standings: [
        { seat: 0, wins: 1, losses: 0, draws: 0 },
        { seat: 1, wins: 0, losses: 1, draws: 0 },
        { seat: 2, wins: 1, losses: 1, draws: 0 },
        { seat: 3, wins: 1, losses: 0, draws: 1 },
      ],
      yours: open,
      table: "t",
    },
    ...over,
  };
}
const withMine = (mine: DraftMatch | undefined, table?: string): DraftView => {
  const v = view();
  return { ...v, play: { ...v.play!, yours: mine, table } };
};

describe("a match", () => {
  it("names its seats and says how it went", () => {
    const v = view();
    expect(seatName(v, 2)).toBe("Sam");
    expect(seatName(v, 7)).toBe("Seat 8");
    expect(matchLine(v, open)).toBe("Sam against Ian: playing");
    expect(matchLine(v, { a: 2, b: 0, done: true, winner: 0, wins_a: 1, wins_b: 2 })).toBe(
      "Ian beat Sam 2–1",
    );
    expect(matchLine(v, { a: 2, b: 0, done: true, wins_a: 1, wins_b: 1 })).toBe(
      "Sam and Ian drew 1–1",
    );
    expect(
      matchLine(v, { a: 1, b: 3, done: true, winner: 3, wins_a: 0, wins_b: 1, coin: true }),
    ).toBe("Bot 3 beat Bot 1 1–0, decided by a coin toss");
    expect(matchLine(v, { a: 1, bye: true, done: true, wins_a: 0, wins_b: 0 })).toBe(
      "Bot 1 sits this round out",
    );
  });

  it("knows the viewer's opponent and result", () => {
    expect(opponentSeat(open, 0)).toBe(2);
    expect(opponentSeat(open, 2)).toBe(0);
    expect(opponentSeat({ a: 0, bye: true, done: true, wins_a: 0, wins_b: 0 }, 0)).toBeUndefined();
    const won: DraftMatch = { ...open, done: true, winner: 0 };
    expect(resultFor(won, 0)).toBe("win");
    expect(resultFor(won, 2)).toBe("loss");
    expect(resultFor({ ...open, done: true }, 0)).toBe("draw");
    expect(resultFor(open, 0)).toBe("");
  });

  it("writes a record with its draws only when there are any", () => {
    expect(recordText({ wins: 2, losses: 1, draws: 0 })).toBe("2–1");
    expect(recordText({ wins: 1, losses: 1, draws: 1 })).toBe("1–1–1");
    expect(matchLengthName(3)).toBe("Best of three");
    expect(matchLengthName(undefined)).toBe("Best of one");
  });
});

describe("the viewer's match of the round in play", () => {
  it("is ready once it has a table, and opening until then", () => {
    expect(yourMatchState(view())).toBe("ready");
    expect(yourMatchLine(view())).toBe("You play Sam. Your table is ready.");
    expect(yourMatchState(withMine(open))).toBe("opening");
    expect(yourMatchLine(withMine(open))).toBe("You play Sam. Your table is being set.");
  });

  it("says a bot is a bot", () => {
    const v = withMine({ a: 0, b: 3, done: false, wins_a: 0, wins_b: 0 }, "t");
    expect(yourMatchLine(v)).toBe("You play Bot 3, a bot. Your table is ready.");
  });

  it("reads a finished match and a bye", () => {
    expect(yourMatchState(withMine({ ...open, done: true, winner: 0 }))).toBe("won");
    expect(yourMatchLine(withMine({ ...open, done: true, winner: 2 }))).toContain("Sam beat you.");
    expect(yourMatchState(withMine({ ...open, done: true }))).toBe("drew");
    const bye = withMine({ a: 0, bye: true, done: true, wins_a: 0, wins_b: 0 });
    expect(yourMatchState(bye)).toBe("bye");
    expect(yourMatchLine(bye)).toContain("You sit this round out");
  });

  it("has none once the draft is finished", () => {
    expect(yourMatchState(view({ phase: "finished" }))).toBe("none");
    expect(yourMatchLine(withMine(undefined))).toBe("");
  });
});

describe("the win-loss table", () => {
  it("orders by wins, then losses, then seat, and marks the viewer and the bots", () => {
    const rows = standingRows(view());
    expect(rows.map((r) => r.name)).toEqual(["Ian", "Bot 3", "Sam", "Bot 1"]);
    expect(rows[0].you).toBe(true);
    expect(rows[1].bot).toBe(true);
    expect(rows[2].bot).toBe(false);
    expect(standingRows(view({ play: undefined }))).toEqual([]);
  });
});

describe("what a listener hears", () => {
  it("speaks once for the build, a new round, a table and the end", () => {
    const building = view({ phase: "building", play: undefined });
    expect(roundsAnnouncement(null, building)).toBe("The picks are done. Build your deck.");
    expect(roundsAnnouncement(building, building)).toBe("");
    const r1 = view();
    expect(roundsAnnouncement(building, r1)).toBe(
      "Round 1 of 3. You play Sam. Your table is ready.",
    );
    expect(roundsAnnouncement(r1, r1)).toBe("");
    expect(roundsAnnouncement(withMine(open), r1)).toBe("Your table is ready.");
    const r2 = { ...r1, play: { ...r1.play!, round: 2 } };
    expect(roundsAnnouncement(r1, r2)).toContain("Round 2 of 3.");
    const done = view({ phase: "finished" });
    expect(roundsAnnouncement(r2, done)).toBe("The draft is finished.");
    expect(roundsAnnouncement(done, done)).toBe("");
  });
});
