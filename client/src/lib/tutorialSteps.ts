// tutorialSteps.ts — the tutorial's script (ADR 0076 §2.1).
//
// Eleven steps, ordered by the player's first two turns. Steps 1 and 11
// are buttons (sub-PR 3, #1079); steps 2–10 are the nine middle steps
// (sub-PR 4, #1081), each a TutorialStep (tutorial.ts) with its anchor,
// its predicate and its copy. The copy assumes the player knows Magic:
// it explains the client, never a rule. "Tap this Forest for mana",
// never "lands make mana".
//
// Anchors, checked against the DOM after ADR 0111:
//   your hand                      { label: "your hand" }
//   lands / creatures              { label: "lands", within: "your board" }
//                                  (every opponent's panel has the same lists)
//   the action dock                { label: "actions" }
//   the attention strip            { label: "attention" }
//   step 7, one card               { cardID }  (Card's data-instance-id)
//   step 10, the bot's portrait    { seatID }  (PlayerIdentity's data-seat-id)
//
// What the practice table does that the ADR's table did not foresee, and
// how the steps cope (ADR 0076 §7, sub-PR 4):
//
//   - The game opens in the player's upkeep, and the table forces
//     autoPassPriority off, so nothing moves until the player presses
//     `next`. A land and a creature wait for a main phase, so steps 3
//     and 6 carry a detour that says so and points at the dock.
//   - With autoPassPriority off, a spell the player casts waits on the
//     stack until they press `next`, and the bot's turn waits on them at
//     every step. Step 6's detour and step 9's copy say so.
//   - On turn one the player has one land, and a single land is not a
//     pile. Step 4 completes on resting the pointer on the lands row
//     whether or not a pile has formed yet, and its copy says the next
//     Forest joins this one.
//   - strictMana off does not make every spell castable: the server's
//     move list leaves out what the player cannot pay for, and the hand
//     dims it. Step 6 points at the lit cards, and gives up when no
//     creature in hand is castable.
//   - The coach's cell squeezes the creature row at 1280×800, so step 7
//     points at a land before a creature (abilityCardID).

import type { Anchor, CopyContext, Detour, StepContext, TutorialStep } from "./tutorial";
import type { CardView, GameView, PlayerView } from "./protocol";

// ---- What the predicates read off the board ----

const HAND: Anchor = { label: "your hand" };
const LANDS: Anchor = { label: "lands", within: "your board" };
const CREATURES: Anchor = { label: "creatures", within: "your board" };
const DOCK: Anchor = { label: "actions" };
const STRIP: Anchor = { label: "attention" };

const MAIN_STEPS = new Set(["precombat_main", "postcombat_main"]);
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

/** sorcerySpeed: the viewer's own main phase with nothing on the stack. */
function sorcerySpeed(c: StepContext): boolean {
  return myTurn(c) && MAIN_STEPS.has(c.view?.turn?.step ?? "") && stackEmpty(c.view);
}

/**
 * abilityCardID picks step 7's card: one of the viewer's lands with a
 * menu, untapped first, else their newest creature with one, else
 * anything with one. Null when nothing has a menu, which makes the
 * step's anchor missing and the step advance itself.
 *
 * A land before a mana creature, though the deck's mana creatures were
 * put there for this step: at 1280×800 the self panel's creature row is
 * about 45px tall while the coach card shows, so a creature is a sliver
 * under the lands row, and the lands row shows its whole card. The menu
 * a Forest opens teaches the same gesture.
 */
export function abilityCardID(v: GameView | null, viewerID: string | null): string | null {
  const withMenu = permanentsOf(v, viewerID).filter(
    (c) => (c.mana_abilities?.length ?? 0) > 0 || (c.activated_abilities?.length ?? 0) > 0,
  );
  const lands = withMenu.filter((c) => isA(c, "Land"));
  const pick =
    lands.find((c) => !c.tapped) ??
    lands[0] ??
    [...withMenu].reverse().find((c) => isA(c, "Creature")) ??
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

/** The detour steps 3 and 6 share: a land or a creature waits for a main phase. */
function toMainPhase(c: StepContext, what: string): Detour | null {
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
  body: "You are seated against a practice bot. Mana is not enforced and you can undo, so nothing here can go wrong.",
};

export const READ_HAND: TutorialStep = {
  id: "read-hand",
  n: 2,
  kind: "action",
  title: "Read your hand",
  body: "Your hand shows only the top of each card. Rest the pointer on one and it lifts, so you can read all of it.",
  hint: "Hold the pointer still over any card in your hand for a moment.",
  anchor: HAND,
  hover: { ms: 600, event: "hand-hovered" },
};

export const PLAY_LAND: TutorialStep = {
  id: "play-land",
  n: 3,
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
  n: 4,
  kind: "action",
  title: "Lands stack into piles",
  body: "Lands with the same name share one pile, with a count. Rest the pointer on your lands to fan the pile out and pick one.",
  hint: "One Forest is a pile of one. Next turn's Forest joins it rather than taking a new spot.",
  anchor: LANDS,
  hover: { ms: 600, event: "pile-hovered" },
};

export const TAP_LAND: TutorialStep = {
  id: "tap-land",
  n: 5,
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
  n: 6,
  kind: "action",
  title: "Cast a creature",
  body: "Click a creature in your hand to cast it. The ones you can cast right now have a lit edge.",
  hint: "A dimmed card costs more than the mana you have. A one-drop is castable off one Forest.",
  anchor: HAND,
  done: (c) => {
    if (count(c.view, c.viewerID, "Creature") > count(c.start, c.viewerID, "Creature")) return true;
    // Already cast one this turn: a creature that has just arrived.
    return permanentsOf(c.view, c.viewerID).some((p) => isA(p, "Creature") && p.summoning_sick);
  },
  first: (c) => {
    const onStack = (c.view?.stack?.cards ?? []).some(
      (s) => s.controller === c.viewerID && isA(s, "Creature"),
    );
    if (onStack) {
      return {
        id: "resolve",
        title: "Now let it resolve",
        body: (k) =>
          `Your creature is on the stack, at the top of the screen. ${PressNext(k)} in the dock and it resolves.`,
        anchor: DOCK,
      };
    }
    return toMainPhase(c, "A creature");
  },
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

export const RIGHT_CLICK: TutorialStep = {
  id: "right-click",
  n: 7,
  kind: "action",
  title: "Abilities live on right-click",
  body: "Right-click a permanent to open its abilities; that is where you activate them. Try this one.",
  hint: "Right-click, not left: a left click taps it for mana.",
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
};

export const MOVE_ALONG: TutorialStep = {
  id: "move-along",
  n: 8,
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
};

/** How long step 9 waits on the bot's turn before moving on regardless (§3). */
export const WATCH_TIMEOUT_MS = 90_000;

export const WATCH_BOT: TutorialStep = {
  id: "watch-bot",
  n: 9,
  kind: "watch",
  title: "The bot takes its turn",
  body: (k) =>
    `Finish your turn and the bot plays its own; what it casts shows at the top. Whenever next lights up, press it${k.nextKey ? ` (${k.nextKey})` : ""} to let the bot carry on.`,
  anchor: STRIP,
  status: (c) => {
    const me = seatOf(c.view, c.viewerID);
    if (!me || !c.view) return undefined;
    if (c.view.turn.priority_holder === me.seat) return "Waiting for you";
    return undefined;
  },
  done: (c) => {
    const me = seatOf(c.view, c.viewerID);
    if (!me || !c.view || !c.start) return false;
    // Back to you: your turn, and not the turn the step began in.
    return c.view.turn.active_seat === me.seat && c.view.turn.seq !== c.start.turn.seq;
  },
  timeoutMs: WATCH_TIMEOUT_MS,
};

export const ATTACK: TutorialStep = {
  id: "attack",
  n: 10,
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
    if (!myTurn(c)) {
      return {
        id: "await-turn",
        title: "Attacks wait for your turn",
        body: (k) => `${PressNext(k)} in the dock until your turn comes round.`,
        anchor: DOCK,
      };
    }
    const step = c.view.turn.step;
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
  n: 11,
  kind: "done",
  title: "That is the whole interface",
  body: ({ helpKey, settingsKey }) => {
    const help = helpKey ? `Press ${helpKey} for the keymap` : "The keymap is in Settings";
    const set = settingsKey ? ` and ${settingsKey} for settings` : "";
    return `${help}${set}. Your zone piles are on the rail: library draws, graveyard and exile open a browser.`;
  },
};

/** The script, in order. */
export const TUTORIAL_STEPS: TutorialStep[] = [
  WELCOME,
  READ_HAND,
  PLAY_LAND,
  LAND_PILES,
  TAP_LAND,
  CAST_CREATURE,
  RIGHT_CLICK,
  MOVE_ALONG,
  WATCH_BOT,
  ATTACK,
  HANDOFF,
];
