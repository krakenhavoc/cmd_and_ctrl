// tutorialSteps.ts — the tutorial's script (ADR 0076 §2.1, ADR 0125 §5.1).
//
// Fourteen steps, ordered by the player's first two turns. Steps 1 and
// 14 are buttons (sub-PR 3, #1079); steps 2–13 are the middle steps, each
// a TutorialStep (tutorial.ts) with its anchor, its predicate and its
// copy. Sub-PR 4 (#1081) wrote nine of them; ADR 0125 §5.1 added the
// opening roll (2), the stack (8) and the commander (9). The copy
// assumes the player knows Magic: it explains the client, never a rule.
// "Tap this Forest for mana", never "lands make mana".
//
// Turn one runs steps 2 to 10, the bot's turn step 11, and the player's
// turn two steps 12 and 13 (ADR 0076 §2.1 as amended 2026-10-11,
// #3012). Right-click (12) waits for turn two because the mana creature
// it points at was cast on turn one and could not tap for mana then: on
// turn two the player taps it through its menu, and the step completes
// on that activation. Memnite, free in the scripted hand, is cast on
// turn one (step 10 asks for it) so that it can attack on turn two (13)
// while the mana creature is tapped.
//
// A step that completes marks the first-use hints it `teaches` as seen
// (ADR 0125 §5.3), so the player's first real table does not tell them
// again: the roll (2), the stack (8), the commander (9), the dock (10),
// right-click (12), and the ⋯ menu and the keymap (14).
//
// Anchors come from the contract-label registry (labels.ts, ADR 0125
// §2.2), so a renamed label moves them too:
//   your hand                      { label: L.yourHand }
//   lands / creatures              { label: L.lands, within: L.yourBoard }
//                                  (every opponent's panel has the same lists)
//   the action dock                { label: L.actions }
//   step 2, the roll               { label: L.rollForFirstTurn } while the
//                                  viewer owes a die, else the strip's
//                                  { label: L.openingRoll } banner; the
//                                  chooser's detour, L.chooseFirstTurn
//                                  and its confirm, L.giveFirstTurn.any
//   step 8, the stack pile         { label: L.stackPile.any }
//   step 9, the commander          { label: L.commandZone.any, within: L.yourBoard }
//                                  (the command zone beside the hand, #2349)
//   step 11, Skip to my turn       { label: L.skipToMyTurn, within: L.actions }
//   steps 12/13's detour, the bot  { label: L.seatBoard(name) }
//   step 12, one card              { cardID }  (Card's data-instance-id)
//   step 13, the bot's portrait    { seatID }  (PlayerIdentity's data-seat-id)
//
// What the practice table does that ADR 0076's table did not foresee,
// and how the steps cope (ADR 0076 §7, sub-PR 4; ADR 0125 §5.2):
//
//   - The table opens the opening roll (ADR 0125 §5.2), and nothing is
//     dealt until its winner chooses. Step 2 teaches it; every other step
//     is held still until the deal (heldByOpeningRoll). A practice bot
//     that wins hands the first turn to the player.
//   - The game opens in the player's upkeep once hands are kept, and the
//     table forces Manual auto-pass, so nothing moves until the
//     player presses `next`. A land and a creature wait for a main phase,
//     so steps 4 and 7 carry a detour that says so and points at the
//     dock. Steps 3, 4 and 7 point at the opening hand first while it
//     is still to be kept: it is a stage over the whole table (#2346).
//   - In Manual auto-pass, a spell the player casts waits on the
//     stack until they press `next`: step 8 teaches the stack while it
//     waits. The bot's turn waits on them at every step, so step 11
//     teaches the dock's Skip to my turn toggle (ADR 0143 §4.2; it was
//     autopass) for that (the owner's choice),
//     and step 12 watches the bot's turn while autopass runs it.
//   - On turn one the player has one land, and a single land is not a
//     pile. Step 5 completes on resting the pointer on the lands row
//     whether or not a pile has formed yet, and its copy says the next
//     Forest joins this one.
//   - The table forces strictMana on (ADR 0118 §1), so the move list
//     and the cast agree: the server's move list leaves out what the
//     player cannot pay for and the hand dims it, and a click on a lit
//     card taps the lands for it, spending mana already in the pool
//     first. Step 7 points at the lit cards.
//   - The table deals the player a scripted opening seven (ADR 0076
//     §2.2 as amended 2026-10-11, #2992, #3012), so steps 7, 8, 12 and
//     13 have a one-drop to cast and tap for mana and Memnite to attack
//     with on the first try. A mulligan
//     deals a random seven, so they still wait rather than give up:
//     with nothing castable, step 7 says so and waits for the next
//     main phase; a spell that left the stack before step 8 could show
//     it waits for the next creature; step 12 waits for a turn in which
//     a mana ability is ready; and step 13 waits for a creature that can
//     attack. Each gives up only when the lesson cannot happen this game
//     at all (lessonGone): the game is over, or no creature is left in
//     hand, in the command zone or in the library.
//   - The commander costs more than one turn of lands can pay, so step
//     9 teaches where it is and how it is cast, not the cast.

import type { Anchor, CopyContext, Detour, StepContext, TutorialStep } from "./tutorial";
import { L } from "./labels";
import { openingRollModel, openingRollStatus } from "./openingRoll";
import type { CardView, GameView, PlayerView } from "./protocol";

// ---- What the predicates read off the board ----

const HAND: Anchor = { label: L.yourHand };
const LANDS: Anchor = { label: L.lands, within: L.yourBoard };
const CREATURES: Anchor = { label: L.creatures, within: L.yourBoard };
const DOCK: Anchor = { label: L.actions };
const PASS: Anchor = { label: L.pass, within: L.actions };
const ATTACK_ALL: Anchor = { label: L.attackAll.any, within: L.actions };
const AUTOPASS: Anchor = { label: L.skipToMyTurn, within: L.actions };
const ROLL_REQUEST: Anchor = { label: L.rollForFirstTurn };
const ROLL_BANNER: Anchor = { label: L.openingRoll };
const STACK_PILE: Anchor = { label: L.stackPile.any };
// The viewer's command zone, beside the hand (#2349); the table.commander
// hint anchors to the same element.
const COMMAND_ZONE: Anchor = { label: L.commandZone.any, within: L.yourBoard };

const MAIN_STEPS = new Set(["precombat_main", "postcombat_main"]);
const EARLY_STEPS = new Set(["untap", "upkeep", "draw", "precombat_main"]);
const BEFORE_ATTACKS = new Set(["untap", "upkeep", "draw", "precombat_main", "begin_combat"]);

const isA = (c: CardView, type: string): boolean => (c.type_line ?? "").includes(type);

/** seatOf is the viewer's own seat. */
export function seatOf(v: GameView | null, viewerID: string | null): PlayerView | undefined {
  if (!v || !viewerID) return undefined;
  return v.seats.find((s) => s.id === viewerID);
}

/** opponentOf is the first seat that is not the viewer: the practice bot. */
export function opponentOf(v: GameView | null, viewerID: string | null): PlayerView | undefined {
  if (!v || !viewerID) return undefined;
  return v.seats.find((s) => s.id !== viewerID);
}

/** permanentsOf is what the viewer controls on the battlefield. */
export function permanentsOf(v: GameView | null, viewerID: string | null): CardView[] {
  if (!v || !viewerID) return [];
  return (v.battlefield?.cards ?? []).filter((c) => c.controller === viewerID);
}

function handOf(v: GameView | null, viewerID: string | null): CardView[] {
  return seatOf(v, viewerID)?.hand?.cards ?? [];
}

/** count is how many of the viewer's permanents have `type` on their type line. */
export function count(v: GameView | null, viewerID: string | null, type: string): number {
  return permanentsOf(v, viewerID).filter((c) => isA(c, type)).length;
}

function myTurn(c: StepContext): boolean {
  const me = seatOf(c.view, c.viewerID);
  return !!me && c.view?.turn?.active_seat === me.seat;
}

function stackEmpty(v: GameView | null): boolean {
  return (v?.stack?.count ?? 0) === 0 && (v?.stack_items?.length ?? 0) === 0;
}

/** creatureOnStack: a creature spell of the viewer's is on the stack. */
function creatureOnStack(c: StepContext): boolean {
  return (c.view?.stack?.cards ?? []).some(
    (s) => s.controller === c.viewerID && isA(s, "Creature"),
  );
}

/** sorcerySpeed: the viewer's own main phase with nothing on the stack. */
function sorcerySpeed(c: StepContext): boolean {
  return myTurn(c) && MAIN_STEPS.has(c.view?.turn?.step ?? "") && stackEmpty(c.view);
}

/** The opening roll, as it stands for the viewer. */
function rollFor(c: StepContext): { owes: boolean; chooses: boolean } | null {
  const me = seatOf(c.view, c.viewerID);
  const model = c.view ? openingRollModel(c.view) : null;
  if (!me || !model) return null;
  return {
    owes: model.chooser === null && model.waiting.includes(me.seat),
    chooses: model.chooser === me.seat,
  };
}

const hasMana = (c: CardView): boolean => (c.mana_abilities?.length ?? 0) > 0;
const hasMenu = (c: CardView): boolean => hasMana(c) || (c.activated_abilities?.length ?? 0) > 0;

/**
 * manaReady: the permanent's mana ability can be used now. It is
 * untapped and, if it is a creature, no longer summoning sick: a
 * creature cast this turn cannot pay {T} (CR 302.6), which is why step
 * 12 waits for turn two.
 */
export function manaReady(c: CardView): boolean {
  return hasMana(c) && !c.tapped && !(isA(c, "Creature") && !!c.summoning_sick);
}

/**
 * manaLessonCards are the viewer's permanents step 12 can be done with:
 * those with a mana ability that are not lands (the deck's mana
 * creatures are there for this step), or the lands when there are none,
 * since a Forest's menu has the same row.
 */
function manaLessonCards(v: GameView | null, viewerID: string | null): CardView[] {
  const mana = permanentsOf(v, viewerID).filter(hasMana);
  const others = mana.filter((c) => !isA(c, "Land"));
  return others.length > 0 ? others : mana;
}

/**
 * abilityCardID picks step 12's card: the viewer's newest permanent
 * whose mana ability is ready (a mana creature, else a land, by
 * manaLessonCards), else the newest that has one but is not ready yet,
 * else anything with a menu. Null when nothing has a menu, which makes
 * the step's anchor missing and the step advance itself.
 */
export function abilityCardID(v: GameView | null, viewerID: string | null): string | null {
  const lesson = manaLessonCards(v, viewerID);
  const newest = (cs: CardView[]): CardView | undefined => cs[cs.length - 1];
  const pick =
    newest(lesson.filter(manaReady)) ??
    newest(lesson) ??
    newest(permanentsOf(v, viewerID).filter(hasMenu));
  return pick?.instance_id ?? null;
}

function abilityCard(v: GameView | null, viewerID: string | null): CardView | undefined {
  const id = abilityCardID(v, viewerID);
  return id ? permanentsOf(v, viewerID).find((c) => c.instance_id === id) : undefined;
}

/** abilityCardName is the name of step 12's card, for its copy; undefined when no card has a menu. */
export function abilityCardName(v: GameView | null, viewerID: string | null): string | undefined {
  return abilityCard(v, viewerID)?.name;
}

/**
 * abilityCardLabel is the menu row step 12 asks for: the label of that
 * card's first mana ability ("Add {G}"); undefined when it has none.
 */
export function abilityCardLabel(v: GameView | null, viewerID: string | null): string | undefined {
  return abilityCard(v, viewerID)?.mana_abilities?.[0]?.label || undefined;
}

/** castableCreatureInHand: the first creature in hand the server lists a cast for. */
function castableCreatureInHand(v: GameView | null, viewerID: string | null): CardView | undefined {
  const casts = new Set(
    (v?.legal_moves ?? []).filter((m) => m.kind === "cast").map((m) => m.source ?? ""),
  );
  return handOf(v, viewerID).find((h) => isA(h, "Creature") && casts.has(h.instance_id));
}

/** castableCreature: a creature in hand the server lists a cast for; undefined with no list. */
function castableCreature(v: GameView | null, viewerID: string | null): boolean | undefined {
  const moves = v?.legal_moves;
  if (!moves) return undefined;
  const ids = new Set(
    handOf(v, viewerID)
      .filter((h) => isA(h, "Creature"))
      .map((h) => h.instance_id),
  );
  return moves.some((m) => m.kind === "cast" && ids.has(m.source ?? ""));
}

/** gameOver: the game has ended, or the viewer is out of it. */
function gameOver(c: StepContext): boolean {
  return c.view?.state === "ended" || !!seatOf(c.view, c.viewerID)?.eliminated;
}

/**
 * creatureToCome: a creature the viewer could still cast this game: one
 * in hand, a commander in the command zone, or a library left to draw
 * one from. True when there is no seat to judge.
 */
function creatureToCome(c: StepContext): boolean {
  const me = seatOf(c.view, c.viewerID);
  if (!me) return true;
  return (
    handOf(c.view, c.viewerID).some((h) => isA(h, "Creature")) ||
    (me.command?.cards ?? []).some((h) => isA(h, "Creature")) ||
    (me.library?.count ?? 0) > 0
  );
}

/**
 * lessonGone is why a lesson that needs a creature of the viewer's
 * (steps 7, 8 and 13) cannot happen this game, or null while it can.
 * The reason is shown to the player on the next card.
 */
function lessonGone(c: StepContext, what: string): string | null {
  if (gameOver(c)) return "the game is over";
  if (!creatureToCome(c)) return `no creature left to ${what} this game`;
  return null;
}

// ---- Copy that names the player's own `next` chord ----

/**
 * "press Next (Space)", or plain "press Next" when it is unbound. The
 * same key presses the dock's primary whatever it reads (#3267): Pass
 * while the stack holds something, Next on an empty one.
 */
const pressNext = (k: CopyContext, button: string = L.next): string =>
  k.nextKey ? `press ${button} (${k.nextKey})` : `press ${button}`;
const PressNext = (k: CopyContext): string => {
  const s = pressNext(k);
  return s[0].toUpperCase() + s.slice(1);
};
const PressPass = (k: CopyContext): string => {
  const s = pressNext(k, L.pass);
  return s[0].toUpperCase() + s.slice(1);
};

/** Punctuation keys, named in words: a bare "," reads as a stray comma in a sentence. */
const PUNCTUATION_NAMES: Record<string, string> = {
  ",": "comma",
  ".": "period",
  "?": "question mark",
  ";": "semicolon",
  "/": "slash",
  "\\": "backslash",
  "-": "minus",
  "=": "equals",
  "'": "apostrophe",
  "[": "left bracket",
  "]": "right bracket",
  "`": "backtick",
};

/** keyName says a bound key in a sentence: "the comma key" for ",", the chord itself otherwise. */
const keyName = (key: string): string => {
  const name = PUNCTUATION_NAMES[key];
  return name ? `the ${name} key` : key;
};

/**
 * keepHandFirst is the detour steps 3, 4 and 7 share while the viewer's
 * opening hand waits on them. The opening hand is a stage over the whole
 * table (#2346), so the hand on the board cannot be rested on, and the
 * dock's `next` does nothing for the player until they keep.
 */
function keepHandFirst(c: StepContext): Detour | null {
  const me = seatOf(c.view, c.viewerID);
  if (!c.view?.mulligans_open || !me || me.hand_kept) return null;
  return {
    id: "keep-hand",
    title: "First, keep your hand",
    body: "Keep hand, under your opening hand, starts the game. A mulligan works too.",
    anchor: { label: L.mulligan },
  };
}

/**
 * The detour steps 4 and 7 share: a land or a creature waits for a main
 * phase, and before that for the opening hand to be kept.
 */
function toMainPhase(c: StepContext, what: string): Detour | null {
  const keep = keepHandFirst(c);
  if (keep) return keep;
  if (sorcerySpeed(c)) return null;
  return {
    id: myTurn(c) ? "to-main" : "await-turn",
    title: myTurn(c) ? "First, your main phase" : "First, your turn",
    body: (k) =>
      myTurn(c)
        ? `${what} goes down in your main phase. ${PressNext(k)} in the dock until it reads Main.`
        : `${what} goes down on your own turn. ${PressNext(k)} in the dock until it comes round.`,
    anchor: DOCK,
  };
}

/**
 * nothingToCast is the detour steps 7 and 8 share in the viewer's own
 * main phase when the server lists no cast for any creature in hand: the
 * step waits for a turn in which one can be cast, rather than giving up
 * (ADR 0076 §2.2 as amended 2026-10-11).
 */
function nothingToCast(c: StepContext): Detour | null {
  if (!sorcerySpeed(c) || castableCreature(c.view, c.viewerID) !== false) return null;
  return {
    id: "cast-next-turn",
    title: "Nothing to cast yet",
    body: (k) =>
      `No creature in your hand can be cast yet. Play a land if you have one, or ${pressNext(k)} in the dock until your next turn. This step waits for it.`,
    anchor: DOCK,
  };
}

// ---- The steps ----

export const WELCOME: TutorialStep = {
  id: "welcome",
  n: 1,
  kind: "opening",
  title: "A five-minute practice game",
  // What the table does with mana (strictMana on, ADR 0118 §1,
  // manaEnforcement.ts): the hand offers only what your mana can pay
  // for, and a click casts it with auto_tap, so the game taps the lands
  // the pool is missing.
  body: "You are seated against a practice bot. Click a card your lands can pay for and the game taps them for you. You can undo, so nothing here can go wrong.",
};

/**
 * Step 2 teaches the opening roll (ADR 0125 §5.1, §5.2). It is the one
 * step the roll does not hold: it completes when the roll is over. The
 * spotlight is on the dock's Roll request while the viewer owes a die,
 * and on the strip's banner, where the dice land, while the table waits
 * on the bot. A player who wins gets a detour onto the chooser's sheet
 * (or its confirm, if they pick the bot). A practice bot that wins hands
 * the first turn to the player by itself.
 */
export const OPENING_ROLL: TutorialStep = {
  id: "opening-roll",
  n: 2,
  kind: "action",
  duringOpeningRoll: true,
  title: "Roll for the first turn",
  body: "Every game starts with a roll for the first turn. Press Roll in the dock.",
  hint: "Each player rolls a d20, and a tie rolls again. The highest roll chooses who goes first.",
  anchor: (c) => (rollFor(c)?.owes ? ROLL_REQUEST : ROLL_BANNER),
  // While the table waits on the bot: "Waiting for Practice Bot to
  // roll", or "Practice Bot is choosing who goes first".
  status: (c) => {
    const r = rollFor(c);
    if (!r || r.owes || r.chooses || !c.view) return undefined;
    return openingRollStatus(c.view) || undefined;
  },
  done: (c) => !!c.view && !c.view.opening_roll,
  first: (c) =>
    rollFor(c)?.chooses
      ? {
          id: "choose-first",
          title: "You won the roll",
          body: "Choose who goes first. For this game, take it yourself: press I go first.",
          anchor: [{ label: L.chooseFirstTurn }, { label: L.giveFirstTurn.any }],
        }
      : null,
  teaches: ["table.opening-roll"],
};

export const READ_HAND: TutorialStep = {
  id: "read-hand",
  n: 3,
  kind: "action",
  title: "Read your hand",
  body: "Your hand shows only the top of each card. Rest the pointer on one and the full card shows in the preview at the top right.",
  hint: "Hold the pointer still over any card in your hand for a moment.",
  anchor: HAND,
  hover: { ms: 600, event: "hand-hovered" },
  first: keepHandFirst,
};

export const PLAY_LAND: TutorialStep = {
  id: "play-land",
  n: 4,
  kind: "action",
  title: "Play a land",
  body: "Click a Forest in your hand to play it, or drag it onto the table.",
  hint: "Cards you can play right now have a lit edge; the rest are dimmed.",
  anchor: HAND,
  done: (c) => {
    if (count(c.view, c.viewerID, "Land") > count(c.start, c.viewerID, "Land")) return true;
    // Already played one this turn: the lesson is learned.
    return myTurn(c) && (seatOf(c.view, c.viewerID)?.lands_played_this_turn ?? 0) > 0;
  },
  first: (c) => toMainPhase(c, "A land"),
  cannot: (c) =>
    c.view && !handOf(c.view, c.viewerID).some((h) => isA(h, "Land")) ? "no land in hand" : null,
};

export const LAND_PILES: TutorialStep = {
  id: "land-piles",
  n: 5,
  kind: "action",
  title: "Lands stack into piles",
  body: "Lands with the same name share one pile, with a count. Rest the pointer on your lands to fan the pile out and pick one.",
  hint: "One Forest is a pile of one. Next turn's Forest joins it rather than taking a new spot.",
  anchor: LANDS,
  hover: { ms: 600, event: "pile-hovered" },
};

export const TAP_LAND: TutorialStep = {
  id: "tap-land",
  n: 6,
  kind: "action",
  title: "Tap a land for mana",
  body: "Click a Forest on the table to tap it. Its mana goes into your pool, by your portrait.",
  hint: "Click the Forest that is already on the table, not one in your hand.",
  anchor: LANDS,
  // The Forest just clicked is still under the pointer, and its preview
  // would cover the portrait and the pool this copy points at.
  previewLeft: true,
  done: (c) => (seatOf(c.view, c.viewerID)?.mana_pool?.length ?? 0) > 0,
  recover: {
    // The issue's common wrong action: playing a land when a tap was
    // asked for. Only possible when the land drop is still open.
    when: (c) => count(c.view, c.viewerID, "Land") > count(c.start, c.viewerID, "Land"),
    title: "Close — that played a land",
    body: "That put a land onto the table. To tap one for mana, click a Forest that is already there.",
    hint: "Undo, in the dock beside Skip to my turn, takes the land back if you want it.",
  },
  cannot: (c) => {
    if (!c.view) return null;
    const untapped = permanentsOf(c.view, c.viewerID).some((p) => isA(p, "Land") && !p.tapped);
    return untapped ? null : "no untapped land";
  },
};

export const CAST_CREATURE: TutorialStep = {
  id: "cast-creature",
  n: 7,
  kind: "action",
  title: "Cast a creature",
  body: "Click a creature in your hand to cast it. The ones you can cast right now have a lit edge.",
  hint: "A dimmed card costs more than the mana you have. A one-drop is castable off one Forest.",
  anchor: HAND,
  // Done the moment the spell is on the stack, so step 8 can show the
  // pile while it waits there (ADR 0125 §5.1).
  done: (c) => {
    if (creatureOnStack(c)) return true;
    if (count(c.view, c.viewerID, "Creature") > count(c.start, c.viewerID, "Creature")) return true;
    // Already cast one this turn: a creature that has just arrived.
    return permanentsOf(c.view, c.viewerID).some((p) => isA(p, "Creature") && p.summoning_sick);
  },
  // A hand with nothing castable this turn waits for the next main
  // phase (nothingToCast), and gives up only when no creature is left
  // to cast this game (#2992).
  first: (c) => toMainPhase(c, "A creature") ?? nothingToCast(c),
  cannot: (c) => {
    if (!c.view) return null;
    if (gameOver(c)) return "the game is over";
    const onStack = (c.view.stack?.cards ?? []).some((s) => s.controller === c.viewerID);
    if (onStack) return null;
    return lessonGone(c, "cast");
  },
};

/**
 * Step 8 teaches the stack pile (ADR 0119, ADR 0125 §5.1). It replaces
 * step 7's old "Now let it resolve" detour. It begins with the creature
 * on the stack and completes when the stack has emptied with a new
 * creature on the battlefield. If the stack is empty (the creature
 * resolved before step 7 saw it on the stack, or the spell left it
 * without arriving), it waits for the next creature the player casts,
 * with a detour that asks for one (#2992), and gives up only when no
 * creature is left to cast this game.
 */
export const ON_THE_STACK: TutorialStep = {
  id: "on-the-stack",
  n: 8,
  kind: "action",
  title: "Your spell is on the stack",
  body: (k) =>
    `Your creature waits on the stack, on the left, until both players pass. ${PressPass(k)} in the dock and it resolves.`,
  hint: "The bot passes too, and then your creature enters the battlefield.",
  anchor: [STACK_PILE, PASS],
  done: (c) =>
    !!c.start &&
    stackEmpty(c.view) &&
    count(c.view, c.viewerID, "Creature") > count(c.start, c.viewerID, "Creature"),
  first: (c) => {
    if (!c.view || creatureOnStack(c)) return null;
    // Nothing of the viewer's on the stack to show: the next creature
    // they cast waits there.
    return (
      toMainPhase(c, "A creature") ??
      nothingToCast(c) ?? {
        id: "cast-another",
        title: "Cast another creature",
        body: "The stack is empty now. Cast another creature and it waits there, where this step can show it: click a lit creature in your hand.",
        anchor: HAND,
      }
    );
  },
  cannot: (c) => {
    if (!c.view) return null;
    if (gameOver(c)) return "the game is over";
    if (creatureOnStack(c)) return null;
    return lessonGone(c, "cast");
  },
  teaches: ["table.stack"],
};

/**
 * Step 9 shows the commander (ADR 0125 §5.1): the viewer's command
 * zone, which sits beside the hand (#2349), the element the
 * table.commander hint points at too. A hover step with no bus event:
 * on a device with no hover it moves on after a read, because a tap on
 * the card would cast the commander. The practice decks' commanders
 * cost three, which two turns of lands cannot pay, so the step teaches
 * where it is and how it is cast, not the cast.
 */
export const COMMANDER: TutorialStep = {
  id: "commander",
  n: 9,
  kind: "action",
  title: "Your commander",
  body: "Your commander waits here, in the command zone beside your hand. Once your lands can pay for it, click it to cast it.",
  hint: "Rest the pointer on your commander, beside your hand, for a moment.",
  anchor: COMMAND_ZONE,
  hover: { ms: 600 },
  cannot: (c) =>
    c.view && (seatOf(c.view, c.viewerID)?.command?.count ?? 0) === 0
      ? "no commander in the command zone"
      : null,
  teaches: ["table.commander"],
};

/**
 * castToo is step 10's detour on turn one (#3012): before the player
 * moves the turn on, a creature they can still cast this turn, which
 * with the scripted hand is Memnite, free beside the one-drop step 7
 * cast. Cast now, it can attack on turn two while the one-drop taps for
 * mana in step 12. The detour does not block: `next` moves on without
 * it, and step 13 waits for an attacker. While the creature waits on
 * the stack, the detour says to let it resolve, as step 8 taught.
 */
function castToo(c: StepContext): Detour | null {
  if (!c.view || !myTurn(c)) return null;
  if (creatureOnStack(c)) {
    return {
      id: "resolve-too",
      title: "Let it resolve",
      body: (k) => `${PressPass(k)} in the dock and your creature resolves, as before.`,
      anchor: [STACK_PILE, PASS],
    };
  }
  if (!sorcerySpeed(c)) return null;
  const more = castableCreatureInHand(c.view, c.viewerID);
  if (!more) return null;
  return {
    // The card's id too: the copy names it, so another card is another detour.
    id: `cast-too:${more.instance_id}`,
    title: "One more creature",
    body: `You can still cast ${more.name} this turn. Cast it too: next turn one creature taps for mana while the other attacks.`,
    anchor: HAND,
  };
}

export const MOVE_ALONG: TutorialStep = {
  id: "move-along",
  n: 10,
  kind: "action",
  title: "Move the turn along",
  body: (k) =>
    `The dock's Next button moves the game on a step${k.nextKey ? `, and so does ${k.nextKey}` : ""}. End turn plays out the rest of your turn for you.`,
  hint: "The icons across the dock's top are the steps still to come this turn.",
  anchor: DOCK,
  done: (c) =>
    !!c.view &&
    !!c.start &&
    (c.view.turn.step !== c.start.turn.step || c.view.turn.seq !== c.start.turn.seq),
  first: castToo,
  teaches: ["table.dock"],
};

/** How long step 11 waits for autopass and the bot's turn before moving on regardless (§3). */
export const WATCH_TIMEOUT_MS = 90_000;

/** Is the dock's autopass toggle on? */
const autopassOn = (c: StepContext): boolean => c.client?.autopass === true;

/**
 * Step 11 teaches the Skip to my turn toggle (formerly autopass; the owner's choice, 2026-10-02):
 * the practice table forces Manual auto-pass, so without it the
 * bot's turn waits on the player at every step. The toggle is session
 * state, separate from that setting, and its safety belt switches it off
 * when the player's own main phase comes round (autopassDecision.ts
 * rule 2), so it is off again in the main phase steps 12 and 13 work
 * in, and the forced setting is untouched.
 */
export const WATCH_BOT: TutorialStep = {
  id: "watch-bot",
  n: 11,
  kind: "action",
  title: "Let the bot play",
  body: "Turn on Skip to my turn in the dock. It passes for you, so the bot plays its turn while you watch.",
  hint: "Skip to my turn switches itself off when your next main phase comes round, so it never skips your turn.",
  anchor: AUTOPASS,
  status: (c) => (autopassOn(c) ? "Skip to my turn is on" : undefined),
  done: (c) => {
    const me = seatOf(c.view, c.viewerID);
    // Autopass on, and the bot's turn running by itself.
    return !!me && !!c.view && autopassOn(c) && c.view.turn.active_seat !== me.seat;
  },
  first: (c) => {
    // The safety belt clears the toggle the moment it is switched on in
    // the viewer's own first main phase, so it cannot be taught there.
    if (myTurn(c) && c.view?.turn.step === "precombat_main" && !autopassOn(c)) {
      return {
        id: "leave-main",
        title: "First, leave your main phase",
        body: (k) =>
          `Skip to my turn switches itself off in your own main phase. ${PressNext(k)} once, then turn it on.`,
        anchor: DOCK,
      };
    }
    return null;
  },
  cannot: (c) => {
    const me = seatOf(c.view, c.viewerID);
    if (!me || !c.view || !c.start) return null;
    // Pressed through the bot's whole turn by hand: the moment has gone.
    return c.view.turn.active_seat === me.seat && c.view.turn.seq !== c.start.turn.seq
      ? "the bot's turn is over"
      : null;
  },
  timeoutMs: WATCH_TIMEOUT_MS,
};

/**
 * botTurn is the detour steps 12 and 13 share while it is not the
 * viewer's turn: watch the bot while Skip to my turn plays it out (step
 * 11 hands over here as soon as the bot's turn is running), else press
 * next until the turn comes round. `waits` is the latter's title.
 */
function botTurn(c: StepContext, waits: string): Detour | null {
  if (!c.view || myTurn(c)) return null;
  if (autopassOn(c)) {
    const opp = opponentOf(c.view, c.viewerID);
    return {
      id: "watch-bot",
      title: "Watch the bot play",
      body: "Skip to my turn is passing for you while the bot takes its turn. It hands back at your main phase.",
      anchor: opp ? { label: L.seatBoard(opp.name) } : DOCK,
    };
  }
  return {
    id: "await-turn",
    title: waits,
    body: (k) => `${PressNext(k)} in the dock until your turn comes round.`,
    anchor: DOCK,
  };
}

/**
 * autopassOff is the detour steps 12 and 13 share in the viewer's own
 * turn: Skip to my turn switched on after the main phase (in combat,
 * say) would pass the rest of the turn, combat too.
 */
function autopassOff(c: StepContext): Detour | null {
  if (!c.view || !myTurn(c) || !autopassOn(c) || EARLY_STEPS.has(c.view.turn.step)) return null;
  return {
    id: "autopass-off",
    title: "First, turn off Skip to my turn",
    body: "Skip to my turn would pass the rest of your turn, combat included. Click it off in the dock.",
    anchor: AUTOPASS,
  };
}

/**
 * Step 12 teaches right-click (ADR 0076 §2.1, as amended 2026-10-11 for
 * #3012) on the player's turn two. Its card is the mana creature step 7
 * cast, no longer summoning sick, so its menu's mana row is live: the
 * player right-clicks it, chooses the row, and the mana goes into the
 * pool. The step completes on that activation, read off the snapshot
 * (the card tapped, with the mana in the pool), not on the menu opening:
 * opening a menu is not using it. A left click on a ready mana ability
 * uses it too (ADR 0117 §6), and counts. Step 13's first detour then asks
 * for the mana to be spent.
 *
 * Until the card is ready the step waits: through the bot's turn, to
 * the player's main phase, or a whole turn when nothing of theirs can
 * tap for mana yet. It gives up only when the game is over; with no
 * permanent that has a menu at all, its anchor is missing and it moves
 * on with a note.
 */
export const RIGHT_CLICK: TutorialStep = {
  id: "right-click",
  n: 12,
  kind: "action",
  title: "Abilities live on right-click",
  body: (k) =>
    `Right-click a permanent to open its abilities; that is where you activate them. ${
      k.cardName
        ? `Try it on ${k.cardName}: choose ${k.abilityLabel ?? "its mana ability"}, and the mana goes into your pool to spend.`
        : "Try it on one of your permanents that has a menu."
    }`,
  // ADR 0117 §6: a left click acts only when an ability is ready, as
  // this one is on turn two.
  hint: "Right-click it, or tap its pip, then choose the row. Now that it is ready, a left click uses it too.",
  anchor: (c) => {
    const id = abilityCardID(c.view, c.viewerID);
    return id ? { cardID: id } : null;
  },
  done: (c) => {
    if (!c.view) return false;
    const pool = (seatOf(c.view, c.viewerID)?.mana_pool?.length ?? 0) > 0;
    const before = new Map(permanentsOf(c.start, c.viewerID).map((p) => [p.instance_id, p]));
    // Tapped since the step began, or tapped with its mana still in the
    // pool: a card that was tapped as the step began (a Forest from
    // turn one) untaps and is used again.
    return manaLessonCards(c.view, c.viewerID).some(
      (p) => p.tapped && !p.attacking_target && (pool || !before.get(p.instance_id)?.tapped),
    );
  },
  first: (c) => {
    if (!c.view) return null;
    const wait = botTurn(c, "Wait for your turn") ?? autopassOff(c);
    if (wait) return wait;
    const lesson = manaLessonCards(c.view, c.viewerID);
    if (lesson.length > 0 && !lesson.some(manaReady)) {
      return {
        id: "ready-next-turn",
        title: "Ready next turn",
        body: (k) =>
          `Nothing of yours can tap for mana right now: a creature can from the turn after it arrives, and a tapped one untaps on your turn. ${PressNext(k)} in the dock until your next turn.`,
        anchor: DOCK,
      };
    }
    if (EARLY_STEPS.has(c.view.turn.step) && c.view.turn.step !== "precombat_main") {
      // Step 11's Skip to my turn is still on and carries the player to
      // Main by itself; a press of next as it gets there would pass the
      // main phase away, so this one asks for none.
      if (autopassOn(c)) {
        return {
          id: "main-next",
          title: "Your main phase is next",
          body: "Skip to my turn passes your upkeep and draw for you, and stops at your main phase.",
          anchor: DOCK,
        };
      }
      return {
        id: "to-main",
        title: "First, your main phase",
        body: (k) =>
          `${PressNext(k)} in the dock until it reads Main, where the mana can be spent.`,
        anchor: DOCK,
      };
    }
    return null;
  },
  cannot: (c) => (c.view && gameOver(c) ? "the game is over" : null),
  teaches: ["table.right-click"],
};

export const ATTACK: TutorialStep = {
  id: "attack",
  n: 13,
  kind: "action",
  title: "Attack",
  body: "Click one of your creatures, then the bot's portrait. That declares the attack.",
  hint: "The dock has an attack-with-all button too, which sends every creature that can attack.",
  anchor: (c) => {
    const opp = opponentOf(c.view, c.viewerID);
    return opp ? [CREATURES, { seatID: opp.id }, ATTACK_ALL] : [CREATURES, ATTACK_ALL];
  },
  done: (c) => permanentsOf(c.view, c.viewerID).some((p) => !!p.attacking_target),
  first: (c) => {
    if (!c.view) return null;
    const wait = botTurn(c, "Attacks wait for your turn") ?? autopassOff(c);
    if (wait) return wait;
    const step = c.view.turn.step;
    // No creature at all: the step waits for one (#2992).
    if (count(c.view, c.viewerID, "Creature") === 0) {
      return {
        id: "no-attacker",
        title: "First, a creature to attack with",
        body: "You have no creature on the battlefield. Cast one when you can; it can attack from your next turn.",
        anchor: HAND,
      };
    }
    // Each phrasing has its own id and fixed copy. The coach republishes
    // a detour only when its id changes (tutorial.ts check), so one id
    // whose copy depended on the phase went stale when the phase moved
    // on (#3006).
    // Mana from step 12 still in the pool, and a creature it can help
    // cast: spend it first (#3012). A click on a lit card spends the
    // pool before it taps anything.
    if (
      step === "precombat_main" &&
      (seatOf(c.view, c.viewerID)?.mana_pool?.length ?? 0) > 0 &&
      castableCreatureInHand(c.view, c.viewerID)
    ) {
      return {
        id: "spend-mana",
        title: "First, spend your mana",
        body: "Your mana waits in your pool, by your portrait, until this step ends. Click a lit creature in your hand to spend it.",
        anchor: HAND,
      };
    }
    if (step === "precombat_main") {
      return {
        id: "to-combat",
        title: "First, go to combat",
        // In the main phase the player may still play a land or cast,
        // so invite it before the detour moves them on (turn 2, #2994).
        body: (k) =>
          `Play a land and cast what you like, then ${pressNext(k)} until it reads Declare Attackers.`,
        anchor: DOCK,
      };
    }
    if (step === "begin_combat") {
      // Combat has begun: Main 1 is behind the player, so only the way on.
      return {
        id: "to-attackers",
        title: "First, go to combat",
        body: (k) => `${PressNext(k)} in the dock until it reads Declare Attackers.`,
        anchor: DOCK,
      };
    }
    if (BEFORE_ATTACKS.has(step)) {
      return {
        id: "to-main",
        title: "First, go to combat",
        body: (k) =>
          `${PressNext(k)} in the dock until it reads Main, then play a land and cast what you like.`,
        anchor: DOCK,
      };
    }
    const ready = permanentsOf(c.view, c.viewerID).some(
      (p) => isA(p, "Creature") && !p.tapped && !p.summoning_sick,
    );
    if (step !== "declare_attackers" || !ready) {
      return {
        id: "next-combat",
        title: "Attack next turn",
        body: (k) =>
          `Nothing of yours can attack in this combat. ${PressNext(k)} until your next turn's Declare Attackers.`,
        anchor: DOCK,
      };
    }
    return null;
  },
  // Waits, through the detours above, while a creature can still
  // attack this game; gives up only when none can (#2992).
  cannot: (c) => {
    if (!c.view) return null;
    if (gameOver(c)) return "the game is over";
    if (count(c.view, c.viewerID, "Creature") > 0) return null;
    return lessonGone(c, "attack with");
  },
};

export const HANDOFF: TutorialStep = {
  id: "handoff",
  n: 14,
  kind: "done",
  title: "That is the whole interface",
  body: ({ helpKey, settingsKey }) => {
    const help = helpKey ? `Press ${keyName(helpKey)} for the keymap` : "The keymap is in Settings";
    const set = settingsKey ? ` and ${keyName(settingsKey)} for settings` : "";
    return `${help}${set}. Your zone piles are on the rail: library opens a draw menu, graveyard and exile open a browser. Help, in the ⋯ menu here and in the header elsewhere, replays any of this.`;
  },
  teaches: ["table.more", "table.shortcuts"],
};

/** The script, in order. */
export const TUTORIAL_STEPS: TutorialStep[] = [
  WELCOME,
  OPENING_ROLL,
  READ_HAND,
  PLAY_LAND,
  LAND_PILES,
  TAP_LAND,
  CAST_CREATURE,
  ON_THE_STACK,
  COMMANDER,
  MOVE_ALONG,
  WATCH_BOT,
  RIGHT_CLICK,
  ATTACK,
  HANDOFF,
];
