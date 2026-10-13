// tutorialSteps.test.ts — the middle steps (ADR 0076 §2.1, #1081; ADR
// 0125 §5.1): each step's anchor, predicate, detour and "cannot", against
// boards shaped like the practice table's, what each step teaches (§5.3),
// and one walk through all fourteen.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import {
  TUTORIAL_STEP_COUNT,
  anchorsOf,
  copyText,
  createTutorialRun,
  statusText,
  type CopyContext,
  type TutorialStep,
} from "./tutorial";
import {
  ATTACK,
  CAST_CREATURE,
  COMMANDER,
  HANDOFF,
  LAND_PILES,
  MOVE_ALONG,
  ON_THE_STACK,
  OPENING_ROLL,
  PLAY_LAND,
  READ_HAND,
  RIGHT_CLICK,
  TAP_LAND,
  TUTORIAL_STEPS,
  WATCH_BOT,
  WATCH_TIMEOUT_MS,
  WELCOME,
  abilityCardID,
  abilityCardLabel,
  abilityCardName,
} from "./tutorialSteps";
import { L } from "./labels";
import { HINTS } from "./hints";
import type { GameView } from "./protocol";
import {
  BOT,
  ME,
  auto,
  board,
  card,
  ctx,
  elves,
  forest,
  marwyn,
  memnite,
  walker,
  type Board,
} from "./test/tutorialBoards";

beforeEach(() => {
  vi.useFakeTimers();
});
afterEach(() => {
  vi.useRealTimers();
});

const keys: CopyContext = { helpKey: "?", settingsKey: ",", nextKey: "Space" };

/** A board whose viewer has `n` cards left in the library. */
const withLibrary = (v: GameView, n: number): GameView => {
  const seats = v.seats.map((s) =>
    s.id === ME ? { ...s, library: { ...s.library, count: n } } : s,
  );
  return { ...v, seats } as GameView;
};
const noKeys: CopyContext = { helpKey: "", settingsKey: "", nextKey: "" };

describe("the script", () => {
  it("is fourteen steps in ADR 0125 §5.1's order, ids and kinds", () => {
    expect(TUTORIAL_STEPS).toHaveLength(TUTORIAL_STEP_COUNT);
    expect(TUTORIAL_STEP_COUNT).toBe(14);
    expect(TUTORIAL_STEPS.map((s) => [s.n, s.id, s.kind])).toEqual([
      [1, "welcome", "opening"],
      [2, "opening-roll", "action"],
      [3, "read-hand", "action"],
      [4, "play-land", "action"],
      [5, "land-piles", "action"],
      [6, "tap-land", "action"],
      [7, "cast-creature", "action"],
      [8, "on-the-stack", "action"],
      [9, "commander", "action"],
      [10, "move-along", "action"],
      [11, "watch-bot", "action"],
      [12, "right-click", "action"],
      [13, "attack", "action"],
      [14, "handoff", "done"],
    ]);
  });

  it("teaches the hints ADR 0125 §5.1 lists, and only live table hints", () => {
    expect(
      TUTORIAL_STEPS.filter((s) => s.teaches?.length).map((s) => [s.n, [...s.teaches!]]),
    ).toEqual([
      [2, ["table.opening-roll"]],
      [8, ["table.stack"]],
      [9, ["table.commander"]],
      [10, ["table.dock"]],
      [12, ["table.right-click"]],
      [14, ["table.more", "table.shortcuts"]],
    ]);
    const live = new Map(HINTS.map((h) => [h.id, h]));
    for (const s of TUTORIAL_STEPS) {
      for (const id of s.teaches ?? []) {
        expect(live.get(id)?.place, `${s.id} teaches ${id}`).toBe("table");
      }
    }
  });

  it("holds every step but the roll's own while the roll is open", () => {
    expect(TUTORIAL_STEPS.filter((s) => s.duringOpeningRoll).map((s) => s.id)).toEqual([
      "opening-roll",
    ]);
  });

  it("gives every middle step an anchor, copy and a way to finish", () => {
    for (const s of TUTORIAL_STEPS.slice(1, -1)) {
      expect(s.anchor, s.id).toBeDefined();
      for (const k of [keys, noKeys]) {
        expect(copyText(s.title, k), s.id).not.toBe("");
        expect(copyText(s.body, k), s.id).not.toBe("");
      }
      // A step finishes on its predicate, a hover, or (step 11) a timeout.
      expect(!!s.done || !!s.hover, s.id).toBe(true);
      if (s.kind === "action") expect(s.hint, s.id).toBeDefined();
    }
  });

  it("anchors to the labels the e2e suite asserts on, scoped to your board", () => {
    const v = board({ mine: [forest(), elves()] });
    expect(anchorsOf(OPENING_ROLL, ctx(board({ hand: [], roll: {} })))).toEqual([
      { label: "roll for the first turn" },
    ]);
    expect(anchorsOf(OPENING_ROLL, ctx(board({ hand: [], roll: { rolls: [[0, 9]] } })))).toEqual([
      { label: "opening roll" },
    ]);
    expect(anchorsOf(READ_HAND, ctx(v))).toEqual([{ label: "your hand" }]);
    expect(anchorsOf(PLAY_LAND, ctx(v))).toEqual([{ label: "your hand" }]);
    expect(anchorsOf(LAND_PILES, ctx(v))).toEqual([{ label: "lands", within: "your board" }]);
    expect(anchorsOf(TAP_LAND, ctx(v))).toEqual([{ label: "lands", within: "your board" }]);
    expect(anchorsOf(CAST_CREATURE, ctx(v))).toEqual([{ label: "your hand" }]);
    expect(anchorsOf(ON_THE_STACK, ctx(v))).toEqual([
      { label: L.stackPile.any },
      { label: L.pass, within: L.actions },
    ]);
    expect(anchorsOf(COMMANDER, ctx(v))).toEqual([
      { label: L.commandZone.any, within: "your board" },
    ]);
    expect(anchorsOf(MOVE_ALONG, ctx(v))).toEqual([{ label: "actions" }]);
    expect(anchorsOf(WATCH_BOT, ctx(v))).toEqual([{ label: "Skip to my turn", within: "actions" }]);
    expect(anchorsOf(ATTACK, ctx(v))).toEqual([
      { label: "creatures", within: "your board" },
      { seatID: BOT },
      { label: L.attackAll.any, within: L.actions },
    ]);
  });

  it("names the player's own next key, and copes with none", () => {
    expect(copyText(MOVE_ALONG.body, keys)).toBe(
      "The dock's Next button moves the game on a step, and so does Space. End turn plays out the rest of your turn for you.",
    );
    expect(copyText(MOVE_ALONG.body, noKeys)).not.toContain("so does");
    const leave = WATCH_BOT.first!(ctx(board()))!;
    expect(copyText(leave.body, keys)).toBe(
      "Skip to my turn switches itself off in your own main phase. Press Next (Space) once, then turn it on.",
    );
    expect(copyText(leave.body, noKeys)).toContain("Press Next once");
  });

  it("never sends the player to the ⋯ menu for Undo: it is on the dock's row now", () => {
    for (const s of TUTORIAL_STEPS) {
      const text = [s.title, s.body, s.hint, s.recover?.title, s.recover?.body, s.recover?.hint]
        .map((c) => copyText(c, keys))
        .join(" ");
      // Every sentence that names Undo says where it is: the dock.
      for (const m of text.match(/[^.]*\bUndo\b[^.]*/g) ?? []) {
        expect(m, s.id).toMatch(/in the dock beside Skip to my turn/);
        expect(m, s.id).not.toContain("⋯");
      }
    }
    expect(copyText(TAP_LAND.recover?.hint, keys)).toMatch(
      /Undo, in the dock beside Skip to my turn/,
    );
  });
});

// ADR 0118 §1 (amends ADR 0076 §2.1 step 1): the practice table forces
// strict payment on, so the welcome says a click taps the lands.
describe("step 1: welcome", () => {
  it("says a click on a card your lands can pay for taps them", () => {
    expect(WELCOME.body).toBe(
      "You are seated against a practice bot. Click a card your lands can pay for and the game taps them for you. You can undo, so nothing here can go wrong.",
    );
  });
});

// ADR 0125 §5.1, §5.2: the practice table opens the opening roll.
describe("step 2: roll for the first turn", () => {
  const rolling = (roll: Board["roll"]) => board({ hand: [], step: "untap", roll });

  it("is the one step the roll does not hold, and completes when the roll is over", () => {
    expect(OPENING_ROLL.duringOpeningRoll).toBe(true);
    expect(OPENING_ROLL.done!(ctx(rolling({})))).toBe(false);
    expect(
      OPENING_ROLL.done!(
        ctx(
          rolling({
            rolls: [
              [0, 19],
              [1, 4],
            ],
            chooser: 0,
          }),
        ),
      ),
    ).toBe(false);
    expect(OPENING_ROLL.done!(ctx(board({ step: "upkeep" })))).toBe(true);
    expect(OPENING_ROLL.done!(ctx(null))).toBe(false);
  });

  it("says to press Roll, with the d20 and the tie in its hint", () => {
    expect(copyText(OPENING_ROLL.title, keys)).toBe("Roll for the first turn");
    expect(copyText(OPENING_ROLL.body, keys)).toBe(
      "Every game starts with a roll for the first turn. Press Roll in the dock.",
    );
    expect(OPENING_ROLL.hint).toMatch(/a tie rolls again/);
  });

  it("points at Roll while you owe a die, and at the banner while the bot does", () => {
    expect(anchorsOf(OPENING_ROLL, ctx(rolling({})))).toEqual([{ label: L.rollForFirstTurn }]);
    const waiting = ctx(rolling({ rolls: [[0, 14]] }));
    expect(anchorsOf(OPENING_ROLL, waiting)).toEqual([{ label: L.openingRoll }]);
    expect(statusText(OPENING_ROLL, waiting)).toBe("Waiting for Practice Bot to roll");
    const botChooses = ctx(
      rolling({
        rolls: [
          [0, 3],
          [1, 18],
        ],
        chooser: 1,
      }),
    );
    expect(statusText(OPENING_ROLL, botChooses)).toBe("Practice Bot is choosing who goes first");
    // Your own roll and your own choice need no status: the card says it.
    expect(statusText(OPENING_ROLL, ctx(rolling({})))).toBeUndefined();
  });

  it("detours onto the chooser's sheet, or its confirm, when you win", () => {
    const won = ctx(
      rolling({
        rolls: [
          [0, 19],
          [1, 4],
        ],
        chooser: 0,
      }),
    );
    const d = OPENING_ROLL.first!(won)!;
    expect(d.id).toBe("choose-first");
    expect(copyText(d.title, keys)).toBe("You won the roll");
    expect(copyText(d.body, keys)).toBe(
      "Choose who goes first. For this game, take it yourself: press I go first.",
    );
    expect(d.anchor).toEqual([{ label: L.chooseFirstTurn }, { label: L.giveFirstTurn.any }]);
    expect(
      OPENING_ROLL.first!(
        ctx(
          rolling({
            rolls: [
              [0, 3],
              [1, 18],
            ],
            chooser: 1,
          }),
        ),
      ),
    ).toBeNull();
    expect(OPENING_ROLL.first!(ctx(rolling({})))).toBeNull();
  });
});

describe("step 3: read your hand", () => {
  it("completes on a 600ms rest, and a touch on the hand on a phone", () => {
    expect(READ_HAND.hover).toEqual({ ms: 600, event: "hand-hovered" });
    expect(READ_HAND.done).toBeUndefined();
  });

  // #2346: the opening hand is a stage over the whole table, so the hand
  // on the board cannot be rested on until it is kept.
  it("detours to the opening hand while it is still to be kept", () => {
    const d = READ_HAND.first!(ctx(board({ step: "upkeep", mulligan: { kept: false } })))!;
    expect(d.id).toBe("keep-hand");
    expect(d.anchor).toEqual({ label: "keep or mulligan your hand" });
    expect(READ_HAND.first!(ctx(board({ step: "upkeep", mulligan: { kept: true } })))).toBeNull();
    expect(READ_HAND.first!(ctx(board({ step: "upkeep" })))).toBeNull();
  });
});

describe("step 4: play a land", () => {
  it("completes when a land joins your board", () => {
    const before = board({ hand: [forest()] });
    expect(PLAY_LAND.done!(ctx(before))).toBe(false);
    const after = board({ mine: [forest()], hand: [], landsPlayed: 1 });
    expect(PLAY_LAND.done!(ctx(after, before))).toBe(true);
  });

  it("is already done if a land was played this turn before it began", () => {
    const v = board({ mine: [forest()], landsPlayed: 1 });
    expect(PLAY_LAND.done!(ctx(v, v))).toBe(true);
    // On the bot's turn a count left over from yours does not count.
    const theirs = board({ mine: [forest()], landsPlayed: 1, active: 1 });
    expect(PLAY_LAND.done!(ctx(theirs, theirs))).toBe(false);
  });

  it("detours to your main phase from the upkeep the table opens in", () => {
    const d = PLAY_LAND.first!(ctx(board({ step: "upkeep" })))!;
    expect(d.id).toBe("to-main");
    expect(d.anchor).toEqual({ label: "actions" });
    expect(copyText(d.body, keys)).toBe(
      "A land goes down in your main phase. Press Next (Space) in the dock until it reads Main.",
    );
    expect(PLAY_LAND.first!(ctx(board({ step: "precombat_main" })))).toBeNull();
    expect(PLAY_LAND.first!(ctx(board({ step: "postcombat_main" })))).toBeNull();
    // Your own spell on the stack is not sorcery speed either.
    expect(PLAY_LAND.first!(ctx(board({ stack: [walker()] })))?.id).toBe("to-main");
    expect(PLAY_LAND.first!(ctx(board({ active: 1 })))?.id).toBe("await-turn");
  });

  // Since the roll comes first, this step can begin while the opening
  // hand still waits on the player, when `next` does nothing for them.
  it("detours to the opening hand while it is still to be kept", () => {
    const d = PLAY_LAND.first!(ctx(board({ step: "upkeep", mulligan: { kept: false } })))!;
    expect(d.id).toBe("keep-hand");
    expect(d.anchor).toEqual({ label: "keep or mulligan your hand" });
    expect(copyText(d.title, keys)).toBe("First, keep your hand");
    expect(copyText(d.body, keys)).toBe(
      "Keep hand, under your opening hand, starts the game. A mulligan works too.",
    );
    // Kept, with the bot still deciding: on to the main phase.
    expect(PLAY_LAND.first!(ctx(board({ step: "upkeep", mulligan: { kept: true } })))?.id).toBe(
      "to-main",
    );
    expect(
      CAST_CREATURE.first!(ctx(board({ step: "upkeep", mulligan: { kept: false } })))?.id,
    ).toBe("keep-hand");
  });

  it("cannot happen with no land in hand", () => {
    expect(PLAY_LAND.cannot!(ctx(board({ hand: [elves()] })))).toBe("no land in hand");
    expect(PLAY_LAND.cannot!(ctx(board()))).toBeNull();
    expect(PLAY_LAND.cannot!(ctx(null))).toBeNull();
  });
});

describe("step 5: lands stack into piles", () => {
  it("completes on a rest over your lands row, pile or not", () => {
    expect(LAND_PILES.hover).toEqual({ ms: 600, event: "pile-hovered" });
    expect(copyText(LAND_PILES.hint, keys)).toMatch(/Next turn's Forest joins it/);
  });
});

describe("step 6: tap a land for mana", () => {
  it("completes when mana is in your pool", () => {
    expect(TAP_LAND.done!(ctx(board({ mine: [forest()] })))).toBe(false);
    expect(TAP_LAND.done!(ctx(board({ mine: [forest({ tapped: true })], pool: ["G"] })))).toBe(
      true,
    );
  });

  it("recovers from playing a land instead, and stays live", () => {
    const start = board({ mine: [forest()] });
    const played = board({ mine: [forest(), forest()] });
    expect(TAP_LAND.recover!.when(ctx(start, start))).toBe(false);
    expect(TAP_LAND.recover!.when(ctx(played, start))).toBe(true);
    expect(copyText(TAP_LAND.recover!.title, keys)).toBe("Close — that played a land");
  });

  it("cannot happen with no untapped land and nothing in the pool", () => {
    expect(TAP_LAND.cannot!(ctx(board({ mine: [forest({ tapped: true })] })))).toBe(
      "no untapped land",
    );
    expect(TAP_LAND.cannot!(ctx(board({ mine: [] })))).toBe("no untapped land");
    expect(TAP_LAND.cannot!(ctx(board({ mine: [forest()] })))).toBeNull();
  });
});

describe("step 7: cast a creature", () => {
  it("completes when a creature joins your board", () => {
    const before = board({ mine: [forest()] });
    expect(CAST_CREATURE.done!(ctx(before))).toBe(false);
    const after = board({ mine: [forest(), elves({ summoning_sick: true })] });
    expect(CAST_CREATURE.done!(ctx(after, before))).toBe(true);
    // A creature cast earlier this turn already taught it.
    expect(CAST_CREATURE.done!(ctx(after, after))).toBe(true);
    const settled = board({ mine: [forest(), elves()] });
    expect(CAST_CREATURE.done!(ctx(settled, settled))).toBe(false);
  });

  // ADR 0125 §5.1: done while the spell is on the stack, so step 8 can
  // teach the pile; the old "Now let it resolve" detour is step 8 now.
  it("completes the moment your creature spell is on the stack", () => {
    const before = board({ mine: [forest()] });
    expect(CAST_CREATURE.done!(ctx(board({ mine: [forest()], stack: [elves()] }), before))).toBe(
      true,
    );
    // Someone else's creature on the stack is not yours.
    const theirs = elves({ controller: BOT, owner: BOT });
    expect(CAST_CREATURE.done!(ctx(board({ mine: [forest()], stack: [theirs] }), before))).toBe(
      false,
    );
    // A non-creature spell is not the lesson either.
    expect(CAST_CREATURE.done!(ctx(board({ mine: [forest()], stack: [forest()] }), before))).toBe(
      false,
    );
  });

  it("detours to your main phase, and has no resolve detour any more", () => {
    expect(CAST_CREATURE.first!(ctx(board({ step: "draw" })))?.id).toBe("to-main");
    expect(CAST_CREATURE.first!(ctx(board()))).toBeNull();
    for (const b of [board({ stack: [elves()] }), board({ stack: [walker()] })]) {
      expect(CAST_CREATURE.first!(ctx(b))?.id).not.toBe("resolve");
    }
  });

  // #2992: it waits while a creature can still be cast this game.
  it("gives up only when no creature is left to cast this game", () => {
    const empty = board({ hand: [forest()] });
    expect(CAST_CREATURE.cannot!(ctx(empty))).toBe("no creature left to cast this game");
    // A library to draw from, or a commander to cast, keeps it waiting.
    expect(CAST_CREATURE.cannot!(ctx(withLibrary(empty, 40)))).toBeNull();
    expect(CAST_CREATURE.cannot!(ctx(board({ hand: [forest()], command: [marwyn()] })))).toBeNull();
    expect(CAST_CREATURE.cannot!(ctx(board({ hand: [forest()], stack: [elves()] })))).toBeNull();
    expect(CAST_CREATURE.cannot!(ctx(board()))).toBeNull();
    // The game is over: nothing more can happen.
    expect(CAST_CREATURE.cannot!(ctx({ ...board(), state: "ended" } as GameView))).toBe(
      "the game is over",
    );
  });

  it("waits for the next main phase when the server lists no cast for any creature in hand", () => {
    const dreadmaw = card("Colossal Dreadmaw", "Creature — Dinosaur");
    const e = elves();
    const withMoves = (b: Board, casts: string[]) =>
      ({
        ...board(b),
        legal_moves: [
          { kind: "pass", source: "00000000-0000-0000-0000-000000000000" },
          ...casts.map((source) => ({ kind: "cast", source })),
        ],
      }) as unknown as GameView;
    const stuck = ctx(withMoves({ hand: [dreadmaw] }, []));
    expect(CAST_CREATURE.cannot!(stuck)).toBeNull();
    const d = CAST_CREATURE.first!(stuck)!;
    expect(d.id).toBe("cast-next-turn");
    expect(d.anchor).toEqual({ label: "actions" });
    expect(copyText(d.body, keys)).toMatch(/press Next \(Space\) in the dock until your next turn/);
    expect(
      CAST_CREATURE.first!(ctx(withMoves({ hand: [dreadmaw, e] }, [e.instance_id]))),
    ).toBeNull();
    // Outside a main phase it is the main-phase detour, not this one.
    expect(CAST_CREATURE.first!(ctx(withMoves({ hand: [dreadmaw], step: "upkeep" }, [])))?.id).toBe(
      "to-main",
    );
    // No move list at all (the server owes no decision): no verdict.
    expect(CAST_CREATURE.first!(ctx(board({ hand: [dreadmaw] })))).toBeNull();
  });
});

// ADR 0125 §5.1: the stack pile, while the creature waits on it.
describe("step 8: your spell is on the stack", () => {
  const f = () => forest({ tapped: true });

  it("points at the stack pile and names your own next key", () => {
    expect(copyText(ON_THE_STACK.body, keys)).toBe(
      "Your creature waits on the stack, on the left, until both players pass. Press Pass (Space) in the dock and it resolves.",
    );
    expect(copyText(ON_THE_STACK.body, noKeys)).toContain("Press Pass in the dock");
  });

  it("completes when the stack empties with the creature on the battlefield", () => {
    const e = elves();
    const land = f();
    const waiting = board({ mine: [land], stack: [e] });
    expect(ON_THE_STACK.done!(ctx(waiting, waiting))).toBe(false);
    const arrived = board({ mine: [land, { ...e, summoning_sick: true }] });
    expect(ON_THE_STACK.done!(ctx(arrived, waiting))).toBe(true);
    expect(ON_THE_STACK.cannot!(ctx(waiting, waiting))).toBeNull();
  });

  // #2992: an empty stack waits for the next creature, not gives up.
  it("waits for the next creature when the stack was already empty as it began", () => {
    const resolved = board({ mine: [f(), elves({ summoning_sick: true })] });
    expect(ON_THE_STACK.done!(ctx(resolved, resolved))).toBe(false);
    expect(ON_THE_STACK.cannot!(ctx(resolved, resolved))).toBeNull();
    const d = ON_THE_STACK.first!(ctx(resolved, resolved))!;
    expect(d.id).toBe("cast-another");
    expect(d.anchor).toEqual({ label: "your hand" });
    // The next one goes on the stack: the step's own card again.
    const next = elves();
    const waiting = board({ mine: [f(), elves({ summoning_sick: true })], stack: [next] });
    expect(ON_THE_STACK.first!(ctx(waiting, resolved))).toBeNull();
    // It resolves: done.
    const arrived = board({
      mine: [f(), elves({ summoning_sick: true }), { ...next, summoning_sick: true }],
    });
    expect(ON_THE_STACK.done!(ctx(arrived, resolved))).toBe(true);
    // In the bot's turn it waits for the player's own.
    expect(ON_THE_STACK.first!(ctx(board({ active: 1, seq: 2 }), resolved))?.id).toBe("await-turn");
  });

  it("gives up only when no creature is left to cast this game", () => {
    const none = board({ hand: [forest()], mine: [f()] });
    expect(ON_THE_STACK.cannot!(ctx(none, none))).toBe("no creature left to cast this game");
    expect(ON_THE_STACK.cannot!(ctx(withLibrary(none, 10), none))).toBeNull();
  });

  it("gives up when the spell leaves the stack without arriving", () => {
    const land = f();
    const waiting = board({ mine: [land], stack: [elves()] });
    const countered = board({ mine: [land] });
    expect(ON_THE_STACK.done!(ctx(countered, waiting))).toBe(false);
    expect(ON_THE_STACK.cannot!(ctx(countered, waiting))).toBeNull();
    expect(ON_THE_STACK.first!(ctx(countered, waiting))?.id).toBe("cast-another");
  });
});

// ADR 0125 §5.1: the commander beside the hand, a hover step with no bus event.
describe("step 9: your commander", () => {
  it("completes on a rest over your commander, with no touch event", () => {
    expect(COMMANDER.hover).toEqual({ ms: 600 });
    expect(COMMANDER.hover?.event).toBeUndefined();
    expect(COMMANDER.done).toBeUndefined();
    expect(copyText(COMMANDER.body, keys)).toBe(
      "Your commander waits here, in the command zone beside your hand. Once your lands can pay for it, click it to cast it.",
    );
  });

  it("gives up with no commander in the zone", () => {
    expect(COMMANDER.cannot!(ctx(board()))).toBe("no commander in the command zone");
    expect(COMMANDER.cannot!(ctx(board({ command: [marwyn()] })))).toBeNull();
    expect(COMMANDER.cannot!(ctx(null))).toBeNull();
  });
});

describe("step 10: move the turn along", () => {
  it("completes when the step or the turn changes", () => {
    const main = board();
    expect(MOVE_ALONG.done!(ctx(main, main))).toBe(false);
    expect(MOVE_ALONG.done!(ctx(board({ step: "begin_combat" }), main))).toBe(true);
    expect(MOVE_ALONG.done!(ctx(board({ seq: 2, active: 1 }), main))).toBe(true);
    expect(MOVE_ALONG.done!(ctx(null, main))).toBe(false);
  });
});

describe("step 10's detour: one more creature (#3012)", () => {
  it("asks for a creature still castable this turn, by name, before the turn moves on", () => {
    const m = memnite();
    const v = board({
      mine: [forest({ tapped: true }), elves({ summoning_sick: true })],
      hand: [forest(), m],
      casts: [m.instance_id],
    });
    const d = MOVE_ALONG.first!(ctx(v))!;
    expect(d.id).toBe(`cast-too:${m.instance_id}`);
    expect(copyText(d.title, keys)).toBe("One more creature");
    expect(copyText(d.body, keys)).toBe(
      "You can still cast Memnite this turn. Cast it too: next turn one creature taps for mana while the other attacks.",
    );
    expect(d.anchor).toEqual({ label: "your hand" });
    // Nothing castable, no move list, or not your main phase: no detour.
    expect(MOVE_ALONG.first!(ctx(board({ hand: [m], casts: [] })))).toBeNull();
    expect(MOVE_ALONG.first!(ctx(board({ hand: [m] })))).toBeNull();
    expect(
      MOVE_ALONG.first!(ctx(board({ hand: [m], casts: [m.instance_id], step: "begin_combat" }))),
    ).toBeNull();
    expect(
      MOVE_ALONG.first!(ctx(board({ hand: [m], casts: [m.instance_id], active: 1 }))),
    ).toBeNull();
  });

  it("says to let it resolve while it waits on the stack", () => {
    const d = MOVE_ALONG.first!(ctx(board({ stack: [memnite()] })))!;
    expect(d.id).toBe("resolve-too");
    expect(copyText(d.body, keys)).toBe(
      "Press Pass (Space) in the dock and your creature resolves, as before.",
    );
    expect(d.anchor).toEqual([{ label: L.stackPile.any }, { label: L.pass, within: L.actions }]);
  });

  it("never holds the step: next still moves the turn on", () => {
    const m = memnite();
    const main = board({ hand: [m], casts: [m.instance_id] });
    expect(MOVE_ALONG.done!(ctx(board({ hand: [m], step: "begin_combat" }), main))).toBe(true);
  });
});

describe("step 11: let the bot play (autopass)", () => {
  it("completes once autopass is on and the bot's turn is running by itself", () => {
    const mine = board({ step: "begin_combat", seq: 1 });
    const theirs = board({ seq: 2, active: 1, priority: 1 });
    expect(WATCH_BOT.done!(ctx(mine))).toBe(false);
    // Switched on in your own turn: it passes the rest of it for you.
    expect(WATCH_BOT.done!(auto(mine))).toBe(false);
    expect(WATCH_BOT.done!(ctx(theirs, mine))).toBe(false);
    expect(WATCH_BOT.done!(auto(theirs, mine))).toBe(true);
  });

  it("detours out of your own main phase, where the safety belt clears the toggle", () => {
    expect(WATCH_BOT.first!(ctx(board({ step: "precombat_main" })))?.id).toBe("leave-main");
    expect(WATCH_BOT.first!(ctx(board({ step: "begin_combat" })))).toBeNull();
    expect(WATCH_BOT.first!(ctx(board({ step: "postcombat_main" })))).toBeNull();
    expect(WATCH_BOT.first!(ctx(board({ seq: 2, active: 1 })))).toBeNull();
  });

  it("gives up once the bot's turn has been pressed through by hand", () => {
    const mine = board({ step: "begin_combat", seq: 1 });
    expect(WATCH_BOT.cannot!(ctx(board({ seq: 2, active: 1 }), mine))).toBeNull();
    expect(WATCH_BOT.cannot!(ctx(board({ seq: 3, step: "upkeep" }), mine))).toBe(
      "the bot's turn is over",
    );
  });

  it("times out so a stalled bot never wedges it, and says when autopass is on", () => {
    expect(WATCH_BOT.timeoutMs).toBe(WATCH_TIMEOUT_MS);
    expect(statusText(WATCH_BOT, auto(board()))).toBe("Skip to my turn is on");
    expect(statusText(WATCH_BOT, ctx(board()))).toBeUndefined();
  });
});

// #3012: on turn two, a real activation of the mana creature cast on turn one.
describe("step 12: abilities live on right-click", () => {
  it("points at your newest ready mana creature, else one not ready yet, else a land", () => {
    const f = forest();
    const ready = elves();
    const sick = elves({ summoning_sick: true });
    const m = memnite();
    expect(abilityCardID(board({ mine: [f, ready, sick, m] }), ME)).toBe(ready.instance_id);
    const tapped = elves({ tapped: true });
    expect(abilityCardID(board({ mine: [f, tapped, sick] }), ME)).toBe(sick.instance_id);
    // No mana creature: an untapped land, the same row on a Forest.
    const f1 = forest({ tapped: true });
    const f2 = forest();
    expect(abilityCardID(board({ mine: [f1, f2, m] }), ME)).toBe(f2.instance_id);
    // The bot's permanents are never the anchor.
    expect(abilityCardID(board({ mine: [m], theirs: [elves()] }), ME)).toBeNull();
    expect(anchorsOf(RIGHT_CLICK, ctx(board({ mine: [f2] })))).toEqual([
      { cardID: f2.instance_id },
    ]);
    // Nothing with a menu: no anchor, so the step advances itself.
    expect(anchorsOf(RIGHT_CLICK, ctx(board({ mine: [m] })))).toEqual([]);
  });

  it("names the card and the row to choose, and copes with neither", () => {
    const v = board({ mine: [elves({ name: "Elvish Mystic" })] });
    expect(abilityCardName(v, ME)).toBe("Elvish Mystic");
    expect(abilityCardLabel(v, ME)).toBe("Add {G}");
    expect(
      copyText(RIGHT_CLICK.body, { ...keys, cardName: "Elvish Mystic", abilityLabel: "Add {G}" }),
    ).toBe(
      "Right-click a permanent to open its abilities; that is where you activate them. Try it on Elvish Mystic: choose Add {G}, and the mana goes into your pool to spend.",
    );
    expect(copyText(RIGHT_CLICK.body, { ...keys, cardName: "Elvish Mystic" })).toMatch(
      /choose its mana ability/,
    );
    expect(copyText(RIGHT_CLICK.body, keys)).toMatch(/one of your permanents that has a menu\.$/);
    expect(abilityCardName(board({ mine: [walker()] }), ME)).toBeUndefined();
    expect(abilityCardLabel(board({ mine: [walker()] }), ME)).toBeUndefined();
  });

  it("completes on the activation, not on the menu opening", () => {
    const e = elves();
    const start = board({ mine: [forest(), e] });
    expect(RIGHT_CLICK.done!(ctx(start, start))).toBe(false);
    expect(RIGHT_CLICK.done!({ ...ctx(start, start), event: "ability-menu-opened" })).toBe(false);
    // Add {G}: the elf taps and its mana is in the pool.
    const used = board({ mine: [forest(), { ...e, tapped: true }], pool: ["G"] });
    expect(RIGHT_CLICK.done!(ctx(used, start))).toBe(true);
    // Tapped by a cast's payment: used all the same.
    const paid = board({ mine: [forest(), { ...e, tapped: true }] });
    expect(RIGHT_CLICK.done!(ctx(paid, start))).toBe(true);
    // An attack taps it too, and is not the lesson.
    const attacking = board({ mine: [forest(), { ...e, tapped: true, attacking_target: BOT }] });
    expect(RIGHT_CLICK.done!(ctx(attacking, start))).toBe(false);
  });

  it("does not count a land while a mana creature is the lesson", () => {
    const f = forest();
    const e = elves();
    const start = board({ mine: [f, e] });
    const forestTapped = board({ mine: [{ ...f, tapped: true }, e], pool: ["G"] });
    expect(RIGHT_CLICK.done!(ctx(forestTapped, start))).toBe(false);
  });

  it("counts a land when there is nothing else, once its mana is in the pool", () => {
    const f = forest({ tapped: true });
    // Tapped on turn one as the step began; untapped since, and used again.
    const start = board({ mine: [f], active: 1, seq: 2 });
    const again = board({ mine: [f], seq: 3 });
    expect(RIGHT_CLICK.done!(ctx(again, start))).toBe(false);
    expect(RIGHT_CLICK.done!(ctx(board({ mine: [f], seq: 3, pool: ["G"] }), start))).toBe(true);
  });

  it("waits through the bot's turn, to your main phase, and for a ready card", () => {
    const id = (c: ReturnType<typeof ctx>) => RIGHT_CLICK.first!(c)?.id ?? null;
    const ready = () => elves();
    // The bot's turn, which step 11 hands over in.
    const watch = RIGHT_CLICK.first!(auto(board({ seq: 2, active: 1, mine: [ready()] })))!;
    expect(watch.id).toBe("watch-bot");
    expect(watch.anchor).toEqual({ label: "Practice Bot board" });
    const wait = RIGHT_CLICK.first!(ctx(board({ seq: 2, active: 1, mine: [ready()] })))!;
    expect(wait.id).toBe("await-turn");
    expect(copyText(wait.title, keys)).toBe("Wait for your turn");
    // Your upkeep and draw: on to the main phase, by hand, or by Skip to
    // my turn, which stops there itself; a press of next as it got there
    // would pass the main phase away, so that card asks for none.
    expect(id(ctx(board({ step: "upkeep", mine: [ready()] })))).toBe("to-main");
    const belt = RIGHT_CLICK.first!(auto(board({ step: "draw", mine: [ready()] })))!;
    expect(belt.id).toBe("main-next");
    expect(copyText(belt.body, keys)).not.toMatch(/next/i);
    // Your main phase, with a ready card: the step itself.
    expect(id(ctx(board({ mine: [ready()] })))).toBeNull();
    // Only a creature cast this turn: it is ready next turn.
    const d = RIGHT_CLICK.first!(ctx(board({ mine: [elves({ summoning_sick: true })] })))!;
    expect(d.id).toBe("ready-next-turn");
    expect(d.anchor).toEqual({ label: "actions" });
    expect(copyText(d.body, keys)).toMatch(
      /Press Next \(Space\) in the dock until your next turn\.$/,
    );
    // Skip to my turn left on past the main phase would pass the turn.
    expect(id(auto(board({ step: "begin_combat", mine: [ready()] })))).toBe("autopass-off");
  });

  it("gives up only when the game is over", () => {
    expect(RIGHT_CLICK.cannot!(ctx(board({ mine: [elves({ summoning_sick: true })] })))).toBeNull();
    expect(RIGHT_CLICK.cannot!(ctx(board({ mine: [] })))).toBeNull();
    expect(RIGHT_CLICK.cannot!(ctx({ ...board(), state: "ended" } as GameView))).toBe(
      "the game is over",
    );
  });

  // ADR 0117 §6: a left click acts only when an ability is ready, as it
  // is on turn two, so the hint says it works too.
  it("hints at right-click and the pip, and that a ready ability takes a left click too", () => {
    expect(RIGHT_CLICK.hint).toBe(
      "Right-click it, or tap its pip, then choose the row. Now that it is ready, a left click uses it too.",
    );
    expect(RIGHT_CLICK.recover).toBeUndefined();
  });
});

describe("step 13: attack", () => {
  const ready = () => elves();

  it("completes when one of your creatures is attacking", () => {
    const v = board({ step: "declare_attackers", mine: [ready()] });
    expect(ATTACK.done!(ctx(v))).toBe(false);
    const attacking = board({
      step: "declare_attackers",
      mine: [elves({ attacking_target: BOT, tapped: true })],
    });
    expect(ATTACK.done!(ctx(attacking))).toBe(true);
  });

  it("detours to your turn, to combat, or to next turn's combat", () => {
    const id = (b: Board) => ATTACK.first!(ctx(board({ mine: [ready()], ...b })))?.id ?? null;
    expect(id({ active: 1, seq: 2 })).toBe("await-turn");
    expect(id({ step: "upkeep" })).toBe("to-main");
    expect(id({ step: "begin_combat" })).toBe("to-attackers");
    expect(id({ step: "declare_attackers" })).toBeNull();
    expect(id({ step: "postcombat_main" })).toBe("next-combat");
    // Only a creature that has just arrived: it attacks next turn.
    expect(
      ATTACK.first!(
        ctx(board({ step: "declare_attackers", mine: [elves({ summoning_sick: true })] })),
      )?.id,
    ).toBe("next-combat");
    const main = ATTACK.first!(ctx(board({ step: "precombat_main", mine: [ready()] })))!;
    expect(copyText(main.body, keys)).toBe(
      "Play a land and cast what you like, then press Next (Space) until it reads Declare Attackers.",
    );
    const d = ATTACK.first!(ctx(board({ step: "upkeep", mine: [ready()] })))!;
    expect(copyText(d.body, keys)).toBe(
      "Press Next (Space) in the dock until it reads Main, then play a land and cast what you like.",
    );
    // Combat has begun: Main 1 is behind the player, so no Main copy.
    const c = ATTACK.first!(ctx(board({ step: "begin_combat", mine: [ready()] })))!;
    expect(copyText(c.body, keys)).toBe(
      "Press Next (Space) in the dock until it reads Declare Attackers.",
    );
  });

  it("gives each phrasing its own id, so the phase moving on republishes it (#3006)", () => {
    const at = (step: string) => ATTACK.first!(ctx(board({ step, mine: [ready()] })))!;
    const pre = at("upkeep");
    const main = at("precombat_main");
    expect(main.id).not.toBe(pre.id);
    expect(copyText(main.body, keys)).not.toMatch(/until it reads Main,/);
  });

  // #3012: step 12's mana is in the pool as this step begins.
  it("asks for step 12's mana to be spent first, on a creature it can cast", () => {
    const e = elves();
    const used = elves({ tapped: true });
    const m = memnite();
    const b = (extra: Board) =>
      board({ step: "precombat_main", mine: [used, m], hand: [forest(), e], ...extra });
    const d = ATTACK.first!(ctx(b({ pool: ["G"], casts: [e.instance_id] })))!;
    expect(d.id).toBe("spend-mana");
    expect(copyText(d.title, keys)).toBe("First, spend your mana");
    expect(copyText(d.body, keys)).toBe(
      "Your mana waits in your pool, by your portrait, until this step ends. Click a lit creature in your hand to spend it.",
    );
    expect(d.anchor).toEqual({ label: "your hand" });
    // Spent, or nothing it can cast: on to combat.
    expect(ATTACK.first!(ctx(b({ casts: [e.instance_id] })))?.id).toBe("to-combat");
    expect(ATTACK.first!(ctx(b({ pool: ["G"], casts: [] })))?.id).toBe("to-combat");
    // Memnite, cast on turn one, is the attacker while the elf is tapped.
    expect(
      ATTACK.first!(
        ctx(board({ step: "declare_attackers", mine: [used, m, elves({ summoning_sick: true })] })),
      ),
    ).toBeNull();
  });

  it("watches the bot's turn while autopass runs it", () => {
    const d = ATTACK.first!(auto(board({ seq: 2, active: 1, mine: [ready()] })))!;
    expect(d.id).toBe("watch-bot");
    expect(d.anchor).toEqual({ label: "Practice Bot board" });
    expect(copyText(d.body, keys)).toMatch(/It hands back at your main phase/);
  });

  it("asks for autopass off when it outlives your main phase", () => {
    const id = (b: Board) => ATTACK.first!(auto(board({ mine: [ready()], ...b })))?.id ?? null;
    // The safety belt's own window: upkeep and draw pass, main clears it.
    expect(id({ step: "upkeep" })).toBe("to-main");
    expect(id({ step: "precombat_main" })).toBe("to-combat");
    // Skip to my turn switched on after the main phase: it would pass the rest of the turn.
    const d = ATTACK.first!(auto(board({ step: "begin_combat", mine: [ready()] })))!;
    expect(d.id).toBe("autopass-off");
    expect(d.anchor).toEqual({ label: "Skip to my turn", within: "actions" });
    expect(id({ step: "declare_attackers" })).toBe("autopass-off");
  });

  // #2992: no creature yet is a wait, not a give-up.
  it("waits for a creature while one can still come", () => {
    const none = board({ mine: [forest()] });
    expect(ATTACK.cannot!(ctx(none))).toBeNull();
    const d = ATTACK.first!(ctx(none))!;
    expect(d.id).toBe("no-attacker");
    expect(d.anchor).toEqual({ label: "your hand" });
    // In the bot's turn, the turn detours come first.
    expect(ATTACK.first!(ctx(board({ mine: [forest()], active: 1, seq: 2 })))?.id).toBe(
      "await-turn",
    );
    expect(ATTACK.cannot!(ctx(board({ mine: [ready()] })))).toBeNull();
  });

  it("gives up only when no creature can attack this game", () => {
    expect(ATTACK.cannot!(ctx(board({ mine: [forest()], hand: [forest()] })))).toBe(
      "no creature left to attack with this game",
    );
    expect(ATTACK.cannot!(ctx(withLibrary(board({ mine: [forest()], hand: [] }), 5)))).toBeNull();
    expect(ATTACK.cannot!(ctx({ ...board({ mine: [ready()] }), state: "ended" } as GameView))).toBe(
      "the game is over",
    );
  });
});

// One player's first two turns, as the practice table plays them: the
// machine walks all fourteen steps on nothing but the board, the bus and
// the coach's hover reports, and reports what each completed step teaches.
describe("a whole tutorial", () => {
  it("walks steps 1 to 14 on the boards a first game produces", () => {
    const log = vi.fn();
    const taught: string[] = [];
    const e = elves({ name: "Elvish Mystic" });
    const m = memnite();
    const e2 = elves();
    const hand = [forest(), forest(), e, m, e2];
    const cmd = [marwyn()];
    let view = board({ hand: [], step: "untap", seq: 0, roll: {} });
    const run = createTutorialRun(TUTORIAL_STEPS, {
      viewerID: ME,
      view,
      log,
      onComplete: (s) => taught.push(...(s.teaches ?? [])),
    });
    const at = () => run.current().step.id;
    const move = (b: Board) => {
      view = board({ command: cmd, ...b });
      run.observe(view);
    };
    run.start();
    // The roll: Roll, the bot rolls, the player wins and takes the turn.
    expect(at()).toBe("opening-roll");
    expect(run.current().held).toBe(false);
    move({ hand: [], step: "untap", seq: 0, roll: { rolls: [[0, 17]] } });
    expect(anchorsOf(run.current().step, run.context())).toEqual([{ label: L.openingRoll }]);
    move({
      hand: [],
      step: "untap",
      seq: 0,
      roll: {
        rolls: [
          [0, 17],
          [1, 6],
        ],
        chooser: 0,
      },
    });
    expect(run.current().detour?.id).toBe("choose-first");
    // The deal: the opening hand waits to be kept.
    move({ step: "upkeep", hand, mulligan: { kept: false } });
    expect(at()).toBe("read-hand");
    // The opening hand is a stage over the table: keep it first.
    expect(run.current().detour?.id).toBe("keep-hand");
    move({ step: "upkeep", hand });
    expect(run.current().detour).toBeNull();
    run.hovered("read-hand");
    expect(at()).toBe("play-land");
    expect(run.current().detour?.id).toBe("to-main");
    move({ step: "draw", hand });
    move({ step: "precombat_main", hand });
    expect(run.current().detour).toBeNull();
    const f = forest();
    const left = [hand[1], m, e2];
    move({ mine: [f], hand: [hand[1], e, m, e2], landsPlayed: 1 });
    expect(at()).toBe("land-piles");
    run.hovered("land-piles");
    expect(at()).toBe("tap-land");
    const tf = { ...f, tapped: true };
    move({ mine: [tf], hand: [hand[1], e, m, e2], landsPlayed: 1, pool: ["G"] });
    expect(at()).toBe("cast-creature");
    move({ mine: [tf], hand: left, landsPlayed: 1, stack: [e] });
    // On the stack: step 8 shows the pile while it waits.
    expect(at()).toBe("on-the-stack");
    expect(anchorsOf(run.current().step, run.context())).toEqual([
      { label: L.stackPile.any },
      { label: L.pass, within: L.actions },
    ]);
    const sickE = { ...e, summoning_sick: true };
    move({ mine: [tf, sickE], hand: left, landsPlayed: 1 });
    // Turn one goes on: the commander, then Memnite before the turn moves on.
    expect(at()).toBe("commander");
    run.hovered("commander");
    expect(at()).toBe("move-along");
    move({ mine: [tf, sickE], hand: left, landsPlayed: 1, casts: [m.instance_id] });
    expect(run.current().detour?.id).toBe(`cast-too:${m.instance_id}`);
    move({ mine: [tf, sickE], hand: [hand[1], e2], landsPlayed: 1, stack: [m] });
    expect(run.current().detour?.id).toBe("resolve-too");
    const sickM = { ...m, summoning_sick: true };
    const turnOne = { mine: [tf, sickE, sickM], hand: [hand[1], e2], landsPlayed: 1 };
    move(turnOne);
    expect(at()).toBe("move-along");
    expect(run.current().detour).toBeNull();
    move({ ...turnOne, step: "begin_combat" });
    expect(at()).toBe("watch-bot");
    expect(run.current().coach).toBe("action");
    // The player turns autopass on: it passes the rest of their turn.
    run.observeClient({ autopass: true });
    expect(at()).toBe("watch-bot");
    move({ ...turnOne, step: "end" });
    // The bot's turn is running by itself: step 12 watches it.
    const theirs = { mine: [tf, sickE, sickM], hand: [hand[1], e2] };
    move({ ...theirs, seq: 2, active: 1, step: "upkeep" });
    expect(at()).toBe("right-click");
    expect(run.current().detour?.id).toBe("watch-bot");
    // Turn two: everything untaps, and nothing is summoning sick.
    const two = { mine: [f, e, m], hand: [hand[1], e2] };
    move({ ...two, seq: 3, step: "upkeep" });
    expect(run.current().detour?.id).toBe("main-next");
    // The safety belt clears the toggle at the player's main phase.
    move({ ...two, seq: 3, step: "precombat_main" });
    run.observeClient({ autopass: false });
    expect(run.current().detour).toBeNull();
    expect(anchorsOf(run.current().step, run.context())).toEqual([{ cardID: e.instance_id }]);
    // The menu opening is not the lesson; Add {G} is.
    run.observe(view, "ability-menu-opened");
    expect(at()).toBe("right-click");
    const usedE = { ...e, tapped: true };
    move({
      mine: [f, usedE, m],
      hand: [hand[1], e2],
      seq: 3,
      pool: ["G"],
      casts: [e2.instance_id],
    });
    expect(at()).toBe("attack");
    // Spend it: the second one-drop.
    expect(run.current().detour?.id).toBe("spend-mana");
    move({ mine: [f, usedE, m], hand: [hand[1]], seq: 3, stack: [e2] });
    expect(run.current().detour?.id).toBe("to-combat");
    const after = { mine: [f, usedE, m, { ...e2, summoning_sick: true }], hand: [hand[1]] };
    move({ ...after, seq: 3, step: "declare_attackers" });
    expect(run.current().detour).toBeNull();
    // Memnite attacks while the elf that made the mana stays tapped.
    move({
      ...after,
      seq: 3,
      step: "declare_attackers",
      mine: [
        f,
        usedE,
        { ...m, attacking_target: BOT, tapped: true },
        { ...e2, summoning_sick: true },
      ],
    });
    expect(run.current().step).toBe(HANDOFF);
    expect(taught).toEqual([
      "table.opening-roll",
      "table.stack",
      "table.commander",
      "table.dock",
      "table.right-click",
    ]);
    // Finish: the hand-off was read to the end.
    run.close();
    expect(taught.slice(-2)).toEqual(["table.more", "table.shortcuts"]);
    expect(log).not.toHaveBeenCalled();
  });

  it("teaches nothing a skipped step would have taught", () => {
    const taught: string[] = [];
    const run = createTutorialRun(TUTORIAL_STEPS, {
      viewerID: ME,
      view: board({ hand: [], step: "untap", seq: 0, roll: {} }),
      log: () => {},
      onComplete: (s) => taught.push(...(s.teaches ?? [])),
    });
    run.start();
    while (run.current().step !== HANDOFF) run.skipStep();
    expect(taught).toEqual([]);
  });

  // The gap PR 3 left (ADR 0125 §5.2): a step entered during the roll
  // lost its timeout to the hold, and never got it back.
  it("gives step 11 its whole timeout once the roll is over, if entered during it", () => {
    const log = vi.fn();
    const run = createTutorialRun(TUTORIAL_STEPS, {
      viewerID: ME,
      view: board({ hand: [], step: "untap", seq: 0, roll: {} }),
      log,
    });
    run.start();
    while (run.current().step.n < WATCH_BOT.n) run.skipStep();
    expect(run.current()).toMatchObject({ step: WATCH_BOT, held: true });
    vi.advanceTimersByTime(WATCH_TIMEOUT_MS * 2);
    expect(run.current().step).toBe(WATCH_BOT);
    run.observe(board({ step: "upkeep", mine: [elves()] }));
    expect(run.current().held).toBe(false);
    vi.advanceTimersByTime(WATCH_TIMEOUT_MS - 1);
    expect(run.current().step).toBe(WATCH_BOT);
    vi.advanceTimersByTime(1);
    expect(run.current().step).toBe(RIGHT_CLICK);
    expect(log).toHaveBeenCalledWith("tutorial: step watch-bot timed out waiting; advancing");
  });

  it("moves on by itself through a stalled bot turn", () => {
    const log = vi.fn();
    const run = createTutorialRun(TUTORIAL_STEPS, {
      viewerID: ME,
      view: board({ mine: [elves()] }),
      log,
    });
    run.start();
    while (run.current().step.n < WATCH_BOT.n) run.skipStep();
    expect(run.current().step).toBe(WATCH_BOT);
    vi.advanceTimersByTime(WATCH_TIMEOUT_MS);
    expect(run.current().step).toBe(RIGHT_CLICK);
    expect(log).toHaveBeenCalledWith("tutorial: step watch-bot timed out waiting; advancing");
  });

  it("skips the steps a hand with no land and no creature cannot do", () => {
    const log = vi.fn();
    const steps: TutorialStep[] = [WELCOME, PLAY_LAND, CAST_CREATURE, HANDOFF];
    const run = createTutorialRun(steps, {
      viewerID: ME,
      view: board({ hand: [card("Giant Growth", "Instant")] }),
      log,
    });
    run.start();
    expect(run.current().step).toBe(HANDOFF);
    expect(log.mock.calls.map((c) => c[0])).toEqual([
      "tutorial: step play-land cannot happen (no land in hand); advancing",
      "tutorial: step cast-creature cannot happen (no creature left to cast this game); advancing",
    ]);
    // The card says so (#2992): both skipped, with the first reason.
    expect(run.current().skipped).toEqual({ steps: [4, 7], why: "no land in hand" });
  });
});

// ADR 0125 §5.2: the practice table opens with the opening roll, and
// nothing is dealt until its winner chooses. No step may complete, give
// up or advance itself on that empty board (heldByOpeningRoll).
describe("while the opening roll is open", () => {
  const rolling = (): GameView =>
    ({
      ...board({ hand: [], step: "untap", seq: 0 }),
      opening_roll: { rounds: [{ seats: [0, 1], rolls: [] }], chooser: -1 },
      mulligans_open: true,
    }) as unknown as GameView;

  it("read-hand does not complete on the empty hand, by a rest or a phone's timer", () => {
    const log = vi.fn();
    const run = createTutorialRun([WELCOME, READ_HAND, PLAY_LAND, HANDOFF], {
      viewerID: ME,
      view: rolling(),
      log,
    });
    run.start();
    expect(run.current()).toMatchObject({ step: READ_HAND, held: true });
    run.hovered(READ_HAND.id);
    run.advance(READ_HAND.id, "cannot be hovered on this device");
    run.anchorMissing(READ_HAND.id);
    expect(run.current().step).toBe(READ_HAND);
    expect(log).not.toHaveBeenCalled();

    // The deal lifts the hold, and the step completes as it always has.
    run.observe(board({ hand: [forest(), elves()], step: "upkeep" }));
    expect(run.current().held).toBe(false);
    run.hovered(READ_HAND.id);
    expect(run.current().step).toBe(PLAY_LAND);
  });

  it("play-land does not give up for want of a land in an undealt hand", () => {
    const log = vi.fn();
    const run = createTutorialRun([WELCOME, PLAY_LAND, HANDOFF], {
      viewerID: ME,
      view: rolling(),
      log,
    });
    run.start();
    run.observe(rolling());
    expect(run.current().step).toBe(PLAY_LAND);
    expect(log).not.toHaveBeenCalled();
  });

  it("changes nothing once the roll has closed: a land in hand keeps the step, none gives it up", () => {
    const log = vi.fn();
    const run = createTutorialRun([WELCOME, PLAY_LAND, HANDOFF], {
      viewerID: ME,
      view: rolling(),
      log,
    });
    run.start();
    run.observe(board({ hand: [forest()], step: "upkeep" }));
    expect(run.current()).toMatchObject({ step: PLAY_LAND, held: false });
    expect(run.current().detour?.id).toBe("to-main");
    expect(log).not.toHaveBeenCalled();
    // The "+1" is measured from the dealt board.
    run.observe(board({ mine: [forest()], hand: [], landsPlayed: 1 }));
    expect(run.current().step).toBe(HANDOFF);

    const gaveUp = createTutorialRun([WELCOME, PLAY_LAND, HANDOFF], {
      viewerID: ME,
      view: rolling(),
      log,
    });
    gaveUp.start();
    gaveUp.observe(board({ hand: [elves()] }));
    expect(gaveUp.current().step).toBe(HANDOFF);
    expect(log).toHaveBeenCalledWith(
      "tutorial: step play-land cannot happen (no land in hand); advancing",
    );
  });
});
