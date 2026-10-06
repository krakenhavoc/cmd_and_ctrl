// tutorialSteps.ts — the tutorial's script (ADR 0076 §2.1, ADR 0125 §5.1).
//
// Fourteen steps, ordered by the player's first two turns. Steps 1 and
// 14 are buttons (sub-PR 3, #1079); steps 2–13 are the middle steps, each
// a TutorialStep (tutorial.ts) with its anchor, its predicate and its
// copy. Sub-PR 4 (#1081) wrote nine of them; ADR 0125 §5.1 added the
// opening roll (2), the stack (8) and the commander (10). The copy
// assumes the player knows Magic: it explains the client, never a rule.
// "Tap this Forest for mana", never "lands make mana".
//
// A step that completes marks the first-use hints it `teaches` as seen
// (ADR 0125 §5.3), so the player's first real table does not tell them
// again: the roll (2), the stack (8), right-click (9), the commander
// (10), the dock (11), and the ⋯ menu and the keymap (14).
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
//   step 10, the commander         { label: L.commandZone.any, within: L.yourBoard }
//                                  (the command zone beside the hand, #2349)
//   step 12, the autopass toggle   { label: L.autopass, within: L.actions }
//   step 13's detour, the bot      { label: L.seatBoard(name) }
//   step 9, one card               { cardID }  (Card's data-instance-id)
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
//     table forces autoPassPriority off, so nothing moves until the
//     player presses `next`. A land and a creature wait for a main phase,
//     so steps 4 and 7 carry a detour that says so and points at the
//     dock. Steps 3, 4 and 7 point at the opening hand first while it
//     is still to be kept: it is a stage over the whole table (#2346).
//   - With autoPassPriority off, a spell the player casts waits on the
//     stack until they press `next`: step 8 teaches the stack while it
//     waits. The bot's turn waits on them at every step, so step 12
//     teaches the dock's autopass toggle for that (the owner's choice),
//     and step 13 watches the bot's turn while autopass runs it.
//   - On turn one the player has one land, and a single land is not a
//     pile. Step 5 completes on resting the pointer on the lands row
//     whether or not a pile has formed yet, and its copy says the next
//     Forest joins this one.
//   - The table forces strictMana on (ADR 0118 §1), so the move list
//     and the cast agree: the server's move list leaves out what the
//     player cannot pay for and the hand dims it, and a click on a lit
//     card taps the lands for it, spending mana already in the pool
//     first. Step 7 points at the lit cards, and gives up when no
//     creature in hand is castable.
//   - The commander costs more than two turns of lands can pay, so step
//     10 teaches where it is and how it is cast, not the cast.

import type { Anchor, CopyContext, Detour, StepContext, TutorialStep } from "./tutorial";
import { L } from "./labels";
import { openingRollModel, openingRollStatus } from "./openingRoll";
import type { CardView, GameView, PlayerView } from "./protocol";

// ---- What the predicates read off the board ----

const HAND: Anchor = { label: L.yourHand };
const LANDS: Anchor = { label: L.lands, within: L.yourBoard };
const CREATURES: Anchor = { label: L.creatures, within: L.yourBoard };
const DOCK: Anchor = { label: L.actions };
const AUTOPASS: Anchor = { label: L.autopass, within: L.actions };
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

/**
 * abilityCardID picks step 9's card: the viewer's newest creature with
 * something behind right-click (the deck's mana creatures are there for
 * this step), else an untapped permanent with a menu, else anything
 * with one: a tapped Forest still opens its menu. Null when nothing has
 * a menu, which makes the step's anchor missing and the step advance
 * itself.
 */
export function abilityCardID(v: GameView | null, viewerID: string | null): string | null {
  const withMenu = permanentsOf(v, viewerID).filter(
    (c) => (c.mana_abilities?.length ?? 0) > 0 || (c.activated_abilities?.length ?? 0) > 0,
  );
  const pick =
    [...withMenu].reverse().find((c) => isA(c, "Creature")) ??
    withMenu.find((c) => !c.tapped) ??
    withMenu[0];
  return pick?.instance_id ?? null;
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

// ---- Copy that names the player's own `next` chord ----

/** "press next (Space)", or plain "press next" when it is unbound. */
const pressNext = (k: CopyContext): string =>
  k.nextKey ? `press next (${k.nextKey})` : "press next";
const PressNext = (k: CopyContext): string => {
  const s = pressNext(k);
  return s[0].toUpperCase() + s.slice(1);
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
  body: "Your hand shows only the top of each card. Rest the pointer on one and it lifts, so you can read all of it.",
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
  done: (c) => (seatOf(c.view, c.viewerID)?.mana_pool?.length ?? 0) > 0,
  recover: {
    // The issue's common wrong action: playing a land when a tap was
    // asked for. Only possible when the land drop is still open.
    when: (c) => count(c.view, c.viewerID, "Land") > count(c.start, c.viewerID, "Land"),
    title: "Close — that played a land",
    body: "That put a land onto the table. To tap one for mana, click a Forest that is already there.",
    hint: "Undo, in the dock beside autopass, takes the land back if you want it.",
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
  first: (c) => toMainPhase(c, "A creature"),
  cannot: (c) => {
    if (!c.view) return null;
    const onStack = (c.view.stack?.cards ?? []).some((s) => s.controller === c.viewerID);
    if (onStack) return null;
    if (!handOf(c.view, c.viewerID).some((h) => isA(h, "Creature"))) return "no creature in hand";
    // The table enforces costs through the server's move list, so a
    // hand of creatures that all cost more than the mana on the table
    // cannot cast one this turn. Only judged in a main phase, where a
    // cast would be listed, and only when there is a list to read.
    if (sorcerySpeed(c) && castableCreature(c.view, c.viewerID) === false) {
      return "no creature in hand is castable yet";
    }
    return null;
  },
};

/**
 * Step 8 teaches the stack pile (ADR 0119, ADR 0125 §5.1). It replaces
 * step 7's old "Now let it resolve" detour. It begins with the creature
 * on the stack and completes when the stack has emptied with the
 * creature on the battlefield. It gives up if the stack was already
 * empty when it began (the creature resolved before step 7 saw it on
 * the stack), and if the spell left the stack without arriving.
 */
export const ON_THE_STACK: TutorialStep = {
  id: "on-the-stack",
  n: 8,
  kind: "action",
  title: "Your spell is on the stack",
  body: (k) =>
    `Your creature waits on the stack, on the left, until both players pass. ${PressNext(k)} in the dock and it resolves.`,
  hint: "The bot passes too, and then your creature enters the battlefield.",
  anchor: STACK_PILE,
  done: (c) =>
    !!c.start &&
    !stackEmpty(c.start) &&
    stackEmpty(c.view) &&
    count(c.view, c.viewerID, "Creature") > count(c.start, c.viewerID, "Creature"),
  cannot: (c) => {
    if (!c.view || !c.start) return null;
    if (stackEmpty(c.start)) return "the stack was already empty";
    if (stackEmpty(c.view)) return "the spell left the stack without resolving";
    return null;
  },
  teaches: ["table.stack"],
};

export const RIGHT_CLICK: TutorialStep = {
  id: "right-click",
  n: 9,
  kind: "action",
  title: "Abilities live on right-click",
  body: "Right-click a permanent to open its abilities; that is where you activate them. Try this one.",
  // ADR 0117 §6: a left click no longer taps this card. The usual
  // anchor is the mana creature cast in step 7, which is summoning
  // sick, so its left click does nothing at all.
  hint: "Right-click it, or tap its pip. A left click only acts when an ability is ready to use.",
  anchor: (c) => {
    const id = abilityCardID(c.view, c.viewerID);
    return id ? { cardID: id } : null;
  },
  done: (c) => c.event === "ability-menu-opened",
  recover: {
    when: (c) =>
      permanentsOf(c.view, c.viewerID).filter((p) => p.tapped).length >
      permanentsOf(c.start, c.viewerID).filter((p) => p.tapped).length,
    title: "That tapped it",
    body: "A left click taps a permanent for mana. A right-click opens the menu with every ability it has.",
    hint: "Undo, in the dock beside autopass, untaps it.",
  },
  teaches: ["table.right-click"],
};

/**
 * Step 10 shows the commander (ADR 0125 §5.1): the viewer's command
 * zone, which sits beside the hand (#2349), the element the
 * table.commander hint points at too. A hover step with no bus event:
 * on a device with no hover it moves on after a read, because a tap on
 * the card would cast the commander. The practice decks' commanders
 * cost three, which two turns of lands cannot pay, so the step teaches
 * where it is and how it is cast, not the cast.
 */
export const COMMANDER: TutorialStep = {
  id: "commander",
  n: 10,
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

export const MOVE_ALONG: TutorialStep = {
  id: "move-along",
  n: 11,
  kind: "action",
  title: "Move the turn along",
  body: (k) =>
    `The dock's next button moves the game on a step${k.nextKey ? `, and so does ${k.nextKey}` : ""}. Pass turn skips to the end of your turn.`,
  hint: "The icons across the dock's top are the steps still to come this turn.",
  anchor: DOCK,
  done: (c) =>
    !!c.view &&
    !!c.start &&
    (c.view.turn.step !== c.start.turn.step || c.view.turn.seq !== c.start.turn.seq),
  teaches: ["table.dock"],
};

/** How long step 12 waits for autopass and the bot's turn before moving on regardless (§3). */
export const WATCH_TIMEOUT_MS = 90_000;

/** Is the dock's autopass toggle on? */
const autopassOn = (c: StepContext): boolean => c.client?.autopass === true;

/**
 * Step 12 teaches the autopass toggle (the owner's choice, 2026-10-02):
 * the practice table forces autoPassPriority off, so without it the
 * bot's turn waits on the player at every step. The toggle is session
 * state, separate from that setting, and its safety belt switches it off
 * when the player's own main phase comes round (autopassDecision.ts
 * rule 2), so step 13 finds it off and the forced setting untouched.
 */
export const WATCH_BOT: TutorialStep = {
  id: "watch-bot",
  n: 12,
  kind: "action",
  title: "Let the bot play",
  body: "Turn on autopass in the dock. It passes for you, so the bot plays its turn while you watch.",
  hint: "Autopass switches itself off when your next main phase comes round, so it never skips your turn.",
  anchor: AUTOPASS,
  status: (c) => (autopassOn(c) ? "Autopass is on" : undefined),
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
          `Autopass switches itself off in your own main phase. ${PressNext(k)} once, then turn it on.`,
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

export const ATTACK: TutorialStep = {
  id: "attack",
  n: 13,
  kind: "action",
  title: "Attack",
  body: "Click one of your creatures, then the bot's portrait. That declares the attack.",
  hint: "The dock has an attack-with-all button too, which sends every creature that can attack.",
  anchor: (c) => {
    const opp = opponentOf(c.view, c.viewerID);
    return opp ? [CREATURES, { seatID: opp.id }] : CREATURES;
  },
  done: (c) => permanentsOf(c.view, c.viewerID).some((p) => !!p.attacking_target),
  first: (c) => {
    if (!c.view) return null;
    if (!myTurn(c) && autopassOn(c)) {
      // Step 12 hands over here as soon as the bot's turn is running.
      const opp = opponentOf(c.view, c.viewerID);
      return {
        id: "watch-bot",
        title: "Watch the bot play",
        body: "Autopass is passing for you while the bot takes its turn. It hands back at your main phase.",
        anchor: opp ? { label: L.seatBoard(opp.name) } : DOCK,
      };
    }
    if (!myTurn(c)) {
      return {
        id: "await-turn",
        title: "Attacks wait for your turn",
        body: (k) => `${PressNext(k)} in the dock until your turn comes round.`,
        anchor: DOCK,
      };
    }
    const step = c.view.turn.step;
    // Autopass outliving the main phase (the player's own
    // autopassPersistThroughTurns) would pass the whole turn, combat too.
    if (autopassOn(c) && !EARLY_STEPS.has(step)) {
      return {
        id: "autopass-off",
        title: "First, autopass off",
        body: "Autopass would pass your whole turn, combat included. Click it off in the dock.",
        anchor: AUTOPASS,
      };
    }
    if (BEFORE_ATTACKS.has(step)) {
      return {
        id: "to-combat",
        title: "First, go to combat",
        body: (k) => `${PressNext(k)} in the dock until it reads Declare Attackers.`,
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
  cannot: (c) =>
    c.view && count(c.view, c.viewerID, "Creature") === 0 ? "no creature to attack with" : null,
};

export const HANDOFF: TutorialStep = {
  id: "handoff",
  n: 14,
  kind: "done",
  title: "That is the whole interface",
  body: ({ helpKey, settingsKey }) => {
    const help = helpKey ? `Press ${helpKey} for the keymap` : "The keymap is in Settings";
    const set = settingsKey ? ` and ${settingsKey} for settings` : "";
    return `${help}${set}. Your zone piles are on the rail: library draws, graveyard and exile open a browser. Help, in the ⋯ menu here and in the header elsewhere, replays any of this.`;
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
  RIGHT_CLICK,
  COMMANDER,
  MOVE_ALONG,
  WATCH_BOT,
  ATTACK,
  HANDOFF,
];
