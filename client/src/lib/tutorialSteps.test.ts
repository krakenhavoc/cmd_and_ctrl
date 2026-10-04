// tutorialSteps.test.ts — the nine middle steps (ADR 0076 §2.1, #1081):
// each step's anchor, predicate, detour and "cannot", against boards
// shaped like the practice table's, and one walk through all eleven.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import {
  TUTORIAL_STEP_COUNT,
  anchorsOf,
  copyText,
  createTutorialRun,
  statusText,
  type CopyContext,
  type StepContext,
  type TutorialStep,
} from "./tutorial";
import {
  ATTACK,
  CAST_CREATURE,
  HANDOFF,
  LAND_PILES,
  MOVE_ALONG,
  PLAY_LAND,
  READ_HAND,
  RIGHT_CLICK,
  TAP_LAND,
  TUTORIAL_STEPS,
  WATCH_BOT,
  WATCH_TIMEOUT_MS,
  WELCOME,
  abilityCardID,
} from "./tutorialSteps";
import type { CardView, GameView } from "./protocol";

beforeEach(() => {
  vi.useFakeTimers();
});
afterEach(() => {
  vi.useRealTimers();
});

const ME = "me";
const BOT = "bot";
const keys: CopyContext = { helpKey: "?", settingsKey: ",", nextKey: "Space" };
const noKeys: CopyContext = { helpKey: "", settingsKey: "", nextKey: "" };

let n = 0;
function card(name: string, type: string, extra: Partial<CardView> = {}): CardView {
  n += 1;
  return {
    instance_id: `${name.toLowerCase().replace(/\W+/g, "-")}-${n}`,
    name,
    owner: ME,
    controller: ME,
    type_line: type,
    ...extra,
  } as CardView;
}
const forest = (extra: Partial<CardView> = {}) =>
  card("Forest", "Basic Land — Forest", {
    mana_abilities: [{ index: 0, label: "Add {G}" }] as unknown as CardView["mana_abilities"],
    ...extra,
  });
const elves = (extra: Partial<CardView> = {}) =>
  card("Llanowar Elves", "Creature — Elf Druid", {
    mana_abilities: [{ index: 0, label: "Add {G}" }] as unknown as CardView["mana_abilities"],
    ...extra,
  });
const walker = (extra: Partial<CardView> = {}) =>
  card("Phyrexian Walker", "Artifact Creature — Phyrexian Construct", extra);

interface Board {
  step?: string;
  active?: number;
  priority?: number;
  seq?: number;
  mine?: CardView[];
  theirs?: CardView[];
  hand?: CardView[];
  pool?: string[];
  landsPlayed?: number;
  stack?: CardView[];
}

function board(b: Board = {}): GameView {
  const zone = (kind: string, cards: CardView[] = []) => ({ kind, count: cards.length, cards });
  const seat = (id: string, i: number, hand: CardView[] = []) => ({
    id,
    name: id === ME ? "Player" : "Practice Bot",
    seat: i,
    life: 40,
    library: zone("library"),
    hand: zone("hand", hand),
    graveyard: zone("graveyard"),
    command: zone("command"),
    mana_pool: id === ME ? (b.pool ?? []) : [],
    lands_played_this_turn: id === ME ? (b.landsPlayed ?? 0) : 0,
  });
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, 0, b.hand ?? [forest(), elves()]), seat(BOT, 1)],
    battlefield: zone("battlefield", [
      ...(b.mine ?? []),
      ...(b.theirs ?? []).map((c) => ({ ...c, owner: BOT, controller: BOT })),
    ]),
    stack: zone("stack", b.stack ?? []),
    exile: zone("exile"),
    stack_items: [],
    turn: {
      seq: b.seq ?? 1,
      number: 1,
      active_seat: b.active ?? 0,
      priority_holder: b.priority ?? b.active ?? 0,
      phase: "",
      step: b.step ?? "precombat_main",
    },
    mulligans_open: false,
  } as unknown as GameView;
}

const ctx = (
  view: GameView | null,
  start: GameView | null = view,
  event = null,
  autopass = false,
): StepContext => ({
  view,
  start,
  viewerID: ME,
  event,
  client: { autopass },
});
const auto = (view: GameView | null, start: GameView | null = view) => ctx(view, start, null, true);

describe("the script", () => {
  it("is eleven steps in ADR 0076 §2.1's order, ids and kinds", () => {
    expect(TUTORIAL_STEPS).toHaveLength(TUTORIAL_STEP_COUNT);
    expect(TUTORIAL_STEPS.map((s) => [s.n, s.id, s.kind])).toEqual([
      [1, "welcome", "opening"],
      [2, "read-hand", "action"],
      [3, "play-land", "action"],
      [4, "land-piles", "action"],
      [5, "tap-land", "action"],
      [6, "cast-creature", "action"],
      [7, "right-click", "action"],
      [8, "move-along", "action"],
      [9, "watch-bot", "action"],
      [10, "attack", "action"],
      [11, "handoff", "done"],
    ]);
  });

  it("gives every middle step an anchor, copy and a way to finish", () => {
    for (const s of TUTORIAL_STEPS.slice(1, -1)) {
      expect(s.anchor, s.id).toBeDefined();
      for (const k of [keys, noKeys]) {
        expect(copyText(s.title, k), s.id).not.toBe("");
        expect(copyText(s.body, k), s.id).not.toBe("");
      }
      // A step finishes on its predicate, a hover, or (step 9) a timeout.
      expect(!!s.done || !!s.hover, s.id).toBe(true);
      if (s.kind === "action") expect(s.hint, s.id).toBeDefined();
    }
  });

  it("anchors to the labels the e2e suite asserts on, scoped to your board", () => {
    const v = board({ mine: [forest(), elves()] });
    expect(anchorsOf(READ_HAND, ctx(v))).toEqual([{ label: "your hand" }]);
    expect(anchorsOf(PLAY_LAND, ctx(v))).toEqual([{ label: "your hand" }]);
    expect(anchorsOf(LAND_PILES, ctx(v))).toEqual([{ label: "lands", within: "your board" }]);
    expect(anchorsOf(TAP_LAND, ctx(v))).toEqual([{ label: "lands", within: "your board" }]);
    expect(anchorsOf(CAST_CREATURE, ctx(v))).toEqual([{ label: "your hand" }]);
    expect(anchorsOf(MOVE_ALONG, ctx(v))).toEqual([{ label: "actions" }]);
    expect(anchorsOf(WATCH_BOT, ctx(v))).toEqual([{ label: "autopass", within: "actions" }]);
    expect(anchorsOf(ATTACK, ctx(v))).toEqual([
      { label: "creatures", within: "your board" },
      { seatID: BOT },
    ]);
  });

  it("names the player's own next key, and copes with none", () => {
    expect(copyText(MOVE_ALONG.body, keys)).toBe(
      "The dock's next button moves the game on a step, and so does Space. Pass turn skips to the end of your turn.",
    );
    expect(copyText(MOVE_ALONG.body, noKeys)).not.toContain("so does");
    const leave = WATCH_BOT.first!(ctx(board()))!;
    expect(copyText(leave.body, keys)).toBe(
      "Autopass switches itself off in your own main phase. Press next (Space) once, then turn it on.",
    );
    expect(copyText(leave.body, noKeys)).toContain("Press next once");
  });

  it("never sends the player to the ⋯ menu for Undo: it is on the dock's row now", () => {
    for (const s of TUTORIAL_STEPS) {
      const text = [s.title, s.body, s.hint, s.recover?.title, s.recover?.body, s.recover?.hint]
        .map((c) => copyText(c, keys))
        .join(" ");
      expect(text, s.id).not.toContain("⋯");
      // Every sentence that names Undo says where it is: the dock.
      for (const m of text.match(/[^.]*\bUndo\b[^.]*/g) ?? []) {
        expect(m, s.id).toMatch(/in the dock beside autopass/);
      }
    }
    expect(copyText(TAP_LAND.recover?.hint, keys)).toMatch(/Undo, in the dock beside autopass/);
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

describe("step 2: read your hand", () => {
  it("completes on a 600ms rest, and a touch on the hand on a phone", () => {
    expect(READ_HAND.hover).toEqual({ ms: 600, event: "hand-hovered" });
    expect(READ_HAND.done).toBeUndefined();
  });
});

describe("step 3: play a land", () => {
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
      "A land goes down in your main phase. Press next (Space) in the dock until it reads Main.",
    );
    expect(PLAY_LAND.first!(ctx(board({ step: "precombat_main" })))).toBeNull();
    expect(PLAY_LAND.first!(ctx(board({ step: "postcombat_main" })))).toBeNull();
    // Your own spell on the stack is not sorcery speed either.
    expect(PLAY_LAND.first!(ctx(board({ stack: [walker()] })))?.id).toBe("to-main");
    expect(PLAY_LAND.first!(ctx(board({ active: 1 })))?.id).toBe("await-turn");
  });

  it("cannot happen with no land in hand", () => {
    expect(PLAY_LAND.cannot!(ctx(board({ hand: [elves()] })))).toBe("no land in hand");
    expect(PLAY_LAND.cannot!(ctx(board()))).toBeNull();
    expect(PLAY_LAND.cannot!(ctx(null))).toBeNull();
  });
});

describe("step 4: lands stack into piles", () => {
  it("completes on a rest over your lands row, pile or not", () => {
    expect(LAND_PILES.hover).toEqual({ ms: 600, event: "pile-hovered" });
    expect(copyText(LAND_PILES.hint, keys)).toMatch(/Next turn's Forest joins it/);
  });
});

describe("step 5: tap a land for mana", () => {
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

describe("step 6: cast a creature", () => {
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

  it("says to press next while your creature waits on the stack", () => {
    const v = board({ stack: [elves()] });
    const d = CAST_CREATURE.first!(ctx(v))!;
    expect(d.id).toBe("resolve");
    expect(d.anchor).toEqual({ label: "actions" });
    expect(copyText(d.body, keys)).toMatch(/Press next \(Space\) in the dock/);
    expect(CAST_CREATURE.first!(ctx(board({ step: "draw" })))?.id).toBe("to-main");
    expect(CAST_CREATURE.first!(ctx(board()))).toBeNull();
  });

  it("cannot happen with no creature in hand or on the stack", () => {
    expect(CAST_CREATURE.cannot!(ctx(board({ hand: [forest()] })))).toBe("no creature in hand");
    expect(CAST_CREATURE.cannot!(ctx(board({ hand: [forest()], stack: [elves()] })))).toBeNull();
    expect(CAST_CREATURE.cannot!(ctx(board()))).toBeNull();
  });

  it("cannot happen when the server lists no cast for any creature in hand", () => {
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
    expect(CAST_CREATURE.cannot!(ctx(withMoves({ hand: [dreadmaw] }, [])))).toBe(
      "no creature in hand is castable yet",
    );
    expect(CAST_CREATURE.cannot!(ctx(withMoves({ hand: [dreadmaw, e] }, [e.instance_id])))).toBe(
      null,
    );
    // Outside a main phase nothing is castable at sorcery speed: no verdict.
    expect(CAST_CREATURE.cannot!(ctx(withMoves({ hand: [dreadmaw], step: "upkeep" }, [])))).toBe(
      null,
    );
    // No move list at all (the server owes no decision): no verdict.
    expect(CAST_CREATURE.cannot!(ctx(board({ hand: [dreadmaw] })))).toBeNull();
  });
});

describe("step 7: abilities live on right-click", () => {
  it("points at your newest creature with a menu, else an untapped land, else any", () => {
    const f1 = forest({ tapped: true });
    const f2 = forest();
    const e1 = elves();
    const e2 = elves();
    const w = walker();
    expect(abilityCardID(board({ mine: [f1, f2, e1, e2, w] }), ME)).toBe(e2.instance_id);
    expect(abilityCardID(board({ mine: [f1, f2, w] }), ME)).toBe(f2.instance_id);
    expect(abilityCardID(board({ mine: [f1, w] }), ME)).toBe(f1.instance_id);
    // The bot's permanents are never the anchor.
    expect(abilityCardID(board({ mine: [w], theirs: [elves()] }), ME)).toBeNull();
    expect(anchorsOf(RIGHT_CLICK, ctx(board({ mine: [f2] })))).toEqual([
      { cardID: f2.instance_id },
    ]);
    // Nothing with a menu: no anchor, so the step advances itself.
    expect(anchorsOf(RIGHT_CLICK, ctx(board({ mine: [w] })))).toEqual([]);
  });

  it("completes when the ability menu opens, and only then", () => {
    const v = board({ mine: [elves()] });
    expect(RIGHT_CLICK.done!(ctx(v))).toBe(false);
    expect(RIGHT_CLICK.done!({ ...ctx(v), event: "hand-hovered" })).toBe(false);
    expect(RIGHT_CLICK.done!({ ...ctx(v), event: "ability-menu-opened" })).toBe(true);
  });

  // ADR 0117 §6: a left click no longer taps the step's card (usually
  // a summoning-sick mana creature, whose left click does nothing), so
  // the hint stops saying it does.
  it("hints at right-click and the pip, not at a left click that taps", () => {
    expect(RIGHT_CLICK.hint).toBe(
      "Right-click it, or tap its pip. A left click only acts when an ability is ready to use.",
    );
    expect(RIGHT_CLICK.hint).not.toMatch(/taps it/);
  });

  it("recovers from a left click that tapped it", () => {
    const e = elves();
    const start = board({ mine: [e] });
    const tapped = board({ mine: [{ ...e, tapped: true }], pool: ["G"] });
    expect(RIGHT_CLICK.recover!.when(ctx(tapped, start))).toBe(true);
    expect(RIGHT_CLICK.recover!.when(ctx(start, start))).toBe(false);
  });
});

describe("step 8: move the turn along", () => {
  it("completes when the step or the turn changes", () => {
    const main = board();
    expect(MOVE_ALONG.done!(ctx(main, main))).toBe(false);
    expect(MOVE_ALONG.done!(ctx(board({ step: "begin_combat" }), main))).toBe(true);
    expect(MOVE_ALONG.done!(ctx(board({ seq: 2, active: 1 }), main))).toBe(true);
    expect(MOVE_ALONG.done!(ctx(null, main))).toBe(false);
  });
});

describe("step 9: let the bot play (autopass)", () => {
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
    expect(statusText(WATCH_BOT, auto(board()))).toBe("Autopass is on");
    expect(statusText(WATCH_BOT, ctx(board()))).toBeUndefined();
  });
});

describe("step 10: attack", () => {
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
    expect(id({ step: "upkeep" })).toBe("to-combat");
    expect(id({ step: "begin_combat" })).toBe("to-combat");
    expect(id({ step: "declare_attackers" })).toBeNull();
    expect(id({ step: "postcombat_main" })).toBe("next-combat");
    // Only a creature that has just arrived: it attacks next turn.
    expect(
      ATTACK.first!(
        ctx(board({ step: "declare_attackers", mine: [elves({ summoning_sick: true })] })),
      )?.id,
    ).toBe("next-combat");
    const d = ATTACK.first!(ctx(board({ step: "upkeep", mine: [ready()] })))!;
    expect(copyText(d.body, keys)).toBe(
      "Press next (Space) in the dock until it reads Declare Attackers.",
    );
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
    expect(id({ step: "upkeep" })).toBe("to-combat");
    expect(id({ step: "precombat_main" })).toBe("to-combat");
    // autopassPersistThroughTurns: it would pass the whole turn.
    const d = ATTACK.first!(auto(board({ step: "begin_combat", mine: [ready()] })))!;
    expect(d.id).toBe("autopass-off");
    expect(d.anchor).toEqual({ label: "autopass", within: "actions" });
    expect(id({ step: "declare_attackers" })).toBe("autopass-off");
  });

  it("cannot happen with no creature on your board", () => {
    expect(ATTACK.cannot!(ctx(board({ mine: [forest()] })))).toBe("no creature to attack with");
    expect(ATTACK.cannot!(ctx(board({ mine: [ready()] })))).toBeNull();
  });
});

// One player's first two turns, as the practice table plays them: the
// machine walks all eleven steps on nothing but the board and the bus.
describe("a whole tutorial", () => {
  it("walks steps 1 to 11 on the boards a first game produces", () => {
    const log = vi.fn();
    const hand = [forest(), forest(), elves(), walker()];
    let view = board({ step: "upkeep", hand });
    const run = createTutorialRun(TUTORIAL_STEPS, { viewerID: ME, view, log });
    const at = () => run.current().step.id;
    const move = (b: Board) => {
      view = board(b);
      run.observe(view);
    };
    run.start();
    expect(at()).toBe("read-hand");
    run.hovered("read-hand");
    expect(at()).toBe("play-land");
    expect(run.current().detour?.id).toBe("to-main");
    move({ step: "draw", hand });
    move({ step: "precombat_main", hand });
    expect(run.current().detour).toBeNull();
    const f = forest();
    move({ mine: [f], hand: hand.slice(1), landsPlayed: 1 });
    expect(at()).toBe("land-piles");
    run.hovered("land-piles");
    expect(at()).toBe("tap-land");
    move({ mine: [{ ...f, tapped: true }], hand: hand.slice(1), landsPlayed: 1, pool: ["G"] });
    expect(at()).toBe("cast-creature");
    const e = hand[2];
    move({
      mine: [{ ...f, tapped: true }],
      hand: [hand[1], hand[3]],
      landsPlayed: 1,
      stack: [e],
    });
    expect(run.current().detour?.id).toBe("resolve");
    const resolved = {
      mine: [
        { ...f, tapped: true },
        { ...e, summoning_sick: true },
      ],
      hand: [hand[1], hand[3]],
      landsPlayed: 1,
    };
    move(resolved);
    expect(at()).toBe("right-click");
    // The creature just cast: the deck's mana creatures are there for this step.
    expect(anchorsOf(run.current().step, run.context())).toEqual([{ cardID: e.instance_id }]);
    run.observe(view, "ability-menu-opened");
    expect(at()).toBe("move-along");
    move({ ...resolved, step: "begin_combat" });
    expect(at()).toBe("watch-bot");
    expect(run.current().coach).toBe("action");
    // The player turns autopass on: it passes the rest of their turn.
    run.observeClient({ autopass: true });
    expect(at()).toBe("watch-bot");
    move({ ...resolved, step: "end" });
    move({ seq: 2, active: 1, step: "upkeep", mine: [f, e], hand: [hand[1], hand[3]] });
    // The bot's turn is running by itself: step 10 watches it.
    expect(at()).toBe("attack");
    expect(run.current().detour?.id).toBe("watch-bot");
    move({ seq: 3, active: 0, step: "upkeep", mine: [f, e], hand: [hand[1], hand[3]] });
    expect(run.current().detour?.id).toBe("to-combat");
    // The safety belt clears the toggle at the player's main phase.
    move({ seq: 3, active: 0, step: "precombat_main", mine: [f, e], hand: [hand[1], hand[3]] });
    run.observeClient({ autopass: false });
    expect(run.current().detour?.id).toBe("to-combat");
    move({ seq: 3, step: "declare_attackers", mine: [f, e], hand: [hand[1], hand[3]] });
    expect(run.current().detour).toBeNull();
    move({
      seq: 3,
      step: "declare_attackers",
      mine: [f, { ...e, attacking_target: BOT, tapped: true }],
    });
    expect(run.current().step).toBe(HANDOFF);
    expect(log).not.toHaveBeenCalled();
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
    expect(run.current().step).toBe(ATTACK);
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
      "tutorial: step cast-creature cannot happen (no creature in hand); advancing",
    ]);
  });
});
