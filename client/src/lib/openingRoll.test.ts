import { describe, expect, it } from "vitest";

import {
  chooserLead,
  giveAwayQuestion,
  isOpeningDieLog,
  latestOpeningDice,
  nameList,
  openingChipText,
  openingRollAsk,
  openingRollModel,
  openingRollStatus,
  rollQuestion,
  startingChoices,
} from "./openingRoll";
import { rollFromLog } from "./dice";
import { passHint } from "./dockHint";
import { primeRandomEvents, trackRandomEvents } from "./reveals";
import type { GameView, LogEvent, OpeningRollView } from "./protocol";

const seats: { seat: number; name: string; eliminated?: boolean }[] = [
  { seat: 0, name: "Alice" },
  { seat: 1, name: "Bob" },
  { seat: 2, name: "Carol" },
  { seat: 3, name: "Dave" },
];

const view = (opening_roll?: OpeningRollView, s = seats) => ({ opening_roll, seats: s });

const ROUND_ONE: OpeningRollView = {
  rounds: [
    {
      seats: [0, 1, 2, 3],
      rolls: [
        { seat: 2, result: 17 },
        { seat: 0, result: 17 },
      ],
    },
  ],
};

// The ADR's own example: round 1 tied Alice and Carol on 17.
const REROLL: OpeningRollView = {
  rounds: [
    {
      seats: [0, 1, 2, 3],
      rolls: [
        { seat: 2, result: 17 },
        { seat: 0, result: 17 },
        { seat: 1, result: 9, by: 0 },
        { seat: 3, result: 4 },
      ],
    },
    { seats: [0, 2], rolls: [{ seat: 2, result: 11 }] },
  ],
};

const CHOSEN: OpeningRollView = {
  rounds: [
    ...REROLL.rounds.slice(0, 1),
    {
      seats: [0, 2],
      rolls: [
        { seat: 2, result: 11 },
        { seat: 0, result: 20 },
      ],
    },
  ],
  chooser: 0,
};

describe("openingRollModel", () => {
  it("is null when no roll is open", () => {
    expect(openingRollModel(view())).toBeNull();
  });

  it("lists every seat in round one: rolled, or rolling", () => {
    const m = openingRollModel(view(ROUND_ONE))!;
    expect(m.roundIndex).toBe(0);
    expect(m.waiting).toEqual([1, 3]);
    expect(m.chooser).toBeNull();
    expect(m.tieResult).toBeNull();
    expect(m.chips.map(openingChipText)).toEqual(["17", "rolling…", "17", "rolling…"]);
    // Nobody is outlined in round one, even level on 17 so far.
    expect(m.chips.some((c) => c.tied)).toBe(false);
  });

  it("outlines a reroll's tied leaders and greys the rest", () => {
    const m = openingRollModel(view(REROLL))!;
    expect(m.roundIndex).toBe(1);
    expect(m.tieResult).toBe(17);
    expect(m.waiting).toEqual([0]);
    expect(m.chips.map(openingChipText)).toEqual(["rolling…", "—", "11", "—"]);
    expect(m.chips.map((c) => c.tied)).toEqual([true, false, true, false]);
  });

  it("marks the chooser and the winning roll once one leader is left", () => {
    const m = openingRollModel(view(CHOSEN))!;
    expect(m.chooser).toBe(0);
    expect(m.winResult).toBe(20);
    expect(m.waiting).toEqual([]);
    expect(m.chips.map((c) => c.chooser)).toEqual([true, false, false, false]);
    expect(m.chips.some((c) => c.tied)).toBe(false);
  });

  it("puts an eliminated seat out of the round", () => {
    const m = openingRollModel(
      view(ROUND_ONE, [...seats.slice(0, 3), { ...seats[3], eliminated: true }]),
    )!;
    expect(m.chips[3].state).toBe("out");
  });
});

describe("openingRollAsk", () => {
  it("asks a seat that owes a die to roll, with the round's question", () => {
    expect(openingRollAsk(openingRollModel(view(ROUND_ONE)), seats, 1, false)).toEqual({
      kind: "roll",
      question: "Roll a d20. The highest roll chooses who goes first.",
      others: [],
    });
    expect(openingRollAsk(openingRollModel(view(REROLL)), seats, 0, false)).toEqual({
      kind: "roll",
      question: "You tied with 17. Roll again.",
      others: [],
    });
  });

  it("gives the host the seats left to roll for, their own excluded", () => {
    const ask = openingRollAsk(openingRollModel(view(ROUND_ONE)), seats, 1, true);
    expect(ask).toMatchObject({ kind: "roll", others: [3] });
  });

  it("leaves the host a wait with the button when they have nothing to roll", () => {
    expect(openingRollAsk(openingRollModel(view(ROUND_ONE)), seats, 0, true)).toEqual({
      kind: "wait",
      question: "Waiting for Bob and Dave to roll",
      others: [1, 3],
    });
  });

  it("asks nothing of a seat with nothing to roll, or of anyone but the chooser", () => {
    expect(openingRollAsk(openingRollModel(view(ROUND_ONE)), seats, 0, false)).toEqual({
      kind: "none",
    });
    expect(openingRollAsk(openingRollModel(view(CHOSEN)), seats, 2, true)).toEqual({
      kind: "none",
    });
    expect(openingRollAsk(openingRollModel(view(ROUND_ONE)), seats, null, true)).toEqual({
      kind: "none",
    });
  });

  it("asks the chooser to choose", () => {
    expect(openingRollAsk(openingRollModel(view(CHOSEN)), seats, 0, false)).toEqual({
      kind: "choose",
      result: 20,
    });
  });
});

describe("the words", () => {
  it("names the table's wait in the status line", () => {
    expect(openingRollStatus(view(ROUND_ONE))).toBe("Waiting for Bob and Dave to roll");
    expect(openingRollStatus(view(CHOSEN))).toBe("Alice is choosing who goes first");
    expect(openingRollStatus(view())).toBe("");
  });

  it("is what the dock's status line says while the roll is open", () => {
    const v = { ...view(CHOSEN), mulligans_open: true } as unknown as GameView;
    expect(passHint(v, false)).toBe("Alice is choosing who goes first");
  });

  it("lists the chooser's buttons in turn order, their own as 'I go first'", () => {
    const choices = startingChoices(
      [...seats.slice(0, 3), { ...seats[3], eliminated: true }].reverse(),
      2,
    );
    expect(choices.map((c) => c.label)).toEqual([
      "Alice goes first",
      "Bob goes first",
      "I go first",
    ]);
  });

  it("phrases the confirm, the lead and the lists", () => {
    expect(giveAwayQuestion("Bob")).toBe("Let Bob take the first turn?");
    expect(chooserLead(20)).toBe("You won the roll with 20. Choose who takes the first turn.");
    expect(nameList(["A"])).toBe("A");
    expect(nameList(["A", "B", "C"])).toBe("A, B and C");
    expect(rollQuestion(openingRollModel(view(ROUND_ONE))!)).toContain("Roll a d20");
  });
});

describe("opening dice in the log", () => {
  const die = (
    seq: number,
    seat: number,
    result: number,
    over: Partial<LogEvent> = {},
  ): LogEvent => ({
    seq,
    kind: "roll",
    seat,
    text: `rolled a d20: ${result}`,
    sides: 20,
    results: [result],
    ...over,
  });

  it("is a d20 with no card, before the first turn", () => {
    expect(isOpeningDieLog(die(1, 0, 5))).toBe(true);
    expect(isOpeningDieLog(die(1, 0, 5, { card_id: "ogre" }))).toBe(false);
    expect(isOpeningDieLog(die(1, 0, 5, { turn: 3 }))).toBe(false);
    expect(isOpeningDieLog(die(1, 0, 5, { kind: "flip" }))).toBe(false);
  });

  it("is marked `opening` for the dice layer", () => {
    expect(rollFromLog(die(1, 0, 5))?.source).toBe("opening");
    expect(rollFromLog(die(1, 0, 5, { card_id: "ogre" }))?.source).toBe("card");
  });

  it("raises no strip cue, while a card's roll still does", () => {
    const s = trackRandomEvents(
      primeRandomEvents([]),
      [die(1, 0, 5), die(2, 1, 9, { card_id: "ogre", turn: 4 })],
      0,
    );
    expect(s.cues.map((c) => c.log.seq)).toEqual([2]);
  });

  it("keeps each seat's latest die as the standings", () => {
    const log = [
      die(1, 0, 15),
      die(2, 1, 15),
      die(3, 2, 3),
      die(5, 1, 8),
      die(6, 0, 2, { card_id: "x" }),
    ];
    expect(latestOpeningDice(log).map((l) => l.seq)).toEqual([1, 5, 3]);
  });
});
