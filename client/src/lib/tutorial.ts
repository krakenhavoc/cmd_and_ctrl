// tutorial.ts — the tutorial's step machine (ADR 0076 §2.1, §2.3, §2.4;
// #1079, sub-PR 3).
//
// Pure state, no DOM: TutorialCoach.svelte feeds it the game view, the
// tutorialBus events and "the anchor is missing", and draws what it says.
// The step list is data (tutorialSteps.ts): sub-PR 4 (#1081) added the
// middle steps, and ADR 0125 §5.1 three more, without reshaping it.
//
// The rules it holds:
//
//   - Every step can be skipped, silently (§2.3). Skip step moves on;
//     Skip tutorial and Finish hide the coach and leave the player at
//     the practice table.
//   - A step whose anchor cannot be found advances itself and logs
//     (§2.4). It never waits on a predicate that cannot fire.
//   - The hint is the second thing a player gets, never the first: it
//     appears after HINT_AFTER_MS on a step that asks for an action.
//   - A step that waits on the bot rather than the player (step 12) can
//     carry a timeout, and advances on its own when it runs out. An
//     action step may carry one too.
//   - While the opening roll is open nothing is dealt, so no step
//     completes, gives up or advances itself (heldByOpeningRoll, ADR
//     0125 §5.2), except the one that teaches the roll (step 2,
//     `duringOpeningRoll`). A held step's timers wait for the deal: its
//     hint and its timeout start when the hold lifts, not before.
//   - A step that completes (its predicate fires, a hover rests, or a
//     button step's button is pressed) reports itself through
//     `onComplete`, which marks the hints it `teaches` as seen (ADR 0125
//     §5.3). A step that is skipped, gives up or advances itself does
//     not, so its hints are still offered at a real table.
//
// Sub-PR 4 (#1081) added what the middle steps need. All of it is
// optional on a step, so the two button steps are unchanged:
//
//   - `first`, a detour: what the player has to do before the step can
//     happen at all. A land waits for your main phase, and a creature
//     on the stack waits for `next`. While the board needs it, the card
//     says that instead and points at what it names. It never blocks.
//   - `cannot`: the board can no longer produce the step's action, for
//     instance no land left in hand to play. The step advances itself
//     and logs, like a missing anchor (§2.4).
//   - `hover`: steps 3, 5 and 10 complete once the pointer has rested on
//     the anchor for `ms` (§2.1, "hover ≥ 600ms"). The coach measures the
//     rest and reports it through `hovered()`.
//   - An anchor can be read off the board (step 2's dialog or banner,
//     step 9's card, step 13's opponent), and so can a status line.

import type { Readable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type { HintID } from "./hints/hint";
import type { LabelRef } from "./labels";
import type { GameView } from "./protocol";
import type { TutorialEvent } from "./tutorialBus";

/** How many steps the tutorial has (ADR 0125 §5.1), for "N / 14". */
export const TUTORIAL_STEP_COUNT = 14;

/** How often a spotlit anchor is measured (TutorialCoach.svelte). */
export const POLL_MS = 100;

/**
 * How long a step's anchor may be missing before the step advances
 * itself. It gets a moment to render first: the hand mounts with the
 * first snapshot, not with the coach.
 */
export const ANCHOR_GRACE_MS = 1_500;

/** Idle time on an action step before its hint shows (the canvas's ~20s). */
export const HINT_AFTER_MS = 20_000;

/**
 * How long a hover step (3, 5 and 10) stays up on a device with no
 * hover before it moves on by itself. A touch on the hand or a pile
 * completes it sooner; this is for the board that has no pile to touch
 * yet, and for the command zone, which has no touch of its own to wait
 * for (a tap there casts the commander).
 */
export const TOUCH_HOVER_STEP_MS = 12_000;

/**
 * What a spotlight points at. A label anchor names an entry in the
 * contract-label registry (labels.ts, ADR 0125 §2.2): a registered aria
 * name, matched exactly, or a dynamic entry's stem (`L.<key>.any`),
 * matched by prefix, suffix or substring. `LabelRef` is branded, so an
 * anchor cannot be spelled as a raw string, and labels.test.ts fails
 * when a registered label's owner stops rendering it.
 *
 * `within` scopes the label to a labelled container. It is not optional
 * in practice for the battlefield rows: every opponent's panel carries
 * the same lands and creatures lists as the viewer's, so those anchors
 * are `{ label: L.lands, within: L.yourBoard }`.
 *
 * `cardID` is a card's data-instance-id (Card.svelte), and `seatID` a
 * seat's portrait (PlayerIdentity's data-seat-id, which CombatArrows
 * already anchors to).
 */
export type Anchor =
  | { label: LabelRef; within?: LabelRef }
  | { cardID: string }
  | { seatID: string };

/**
 * The step kinds. `opening` and `done` are the two button steps (1 and
 * 14); `action` waits on the player; `watch` waits on the bot.
 */
export type StepKind = "opening" | "action" | "watch" | "done";

/** The coach card's six states (the canvas's "every state it has"). */
export type CoachState = "opening" | "action" | "hint" | "watch" | "recovered" | "done";

/** What a step's predicates see. */
export interface StepContext {
  /** The game as it is now. */
  view: GameView | null;
  /** The game as it was when this step began, for "+1" predicates. */
  start: GameView | null;
  viewerID: string | null;
  /** The bus event that prompted this check, if one did. */
  event: TutorialEvent | null;
  /** This tab's own client state, for the one step that teaches it. */
  client?: ClientState;
}

/**
 * Client-side state a step can read, because it never reaches the wire.
 * Only the dock's autopass toggle (Game.svelte's session state), which
 * step 12 teaches. Like the bus (§2.5), it is not a place for more.
 */
export interface ClientState {
  autopass: boolean;
}

/** A step's anchor: fixed, or read off the board as it is now. */
export type AnchorSpec = Anchor | Anchor[] | ((c: StepContext) => Anchor | Anchor[] | null);

/** Copy that can name the player's own key bindings. */
export interface CopyContext {
  /** The chord that opens the keymap overlay, formatted; "" if unbound. */
  helpKey: string;
  /** The chord that opens settings, formatted; "" if unbound. */
  settingsKey: string;
  /** The chord that presses the dock's `next`, formatted; "" if unbound. */
  nextKey?: string;
}

export type Copy = string | ((c: CopyContext) => string);

/**
 * A detour: what the player must do before a step can happen at all. It
 * rewrites the card and moves the spotlight while the board needs it,
 * and goes away on its own once the board allows the step.
 */
export interface Detour {
  /** Stable id, so an unchanged detour is not republished. */
  id: string;
  title: Copy;
  body: Copy;
  /** What the detour points at. Defaults to the step's own anchor. */
  anchor?: Anchor | Anchor[];
}

export interface TutorialStep {
  /** Stable id, for logs and tests. */
  id: string;
  /** Its number in ADR 0125 §5.1's table (1–14). */
  n: number;
  kind: StepKind;
  title: Copy;
  body: Copy;
  /**
   * One anchor, or several whose rects the hole covers together, or a
   * function of the board that names them. A function that names none
   * is a missing anchor, so the step advances itself.
   */
  anchor?: AnchorSpec;
  /** The second thing they get (action steps). */
  hint?: Copy;
  /** The status line, e.g. "Bot is thinking". Defaults by kind. */
  status?: string | ((c: StepContext) => string | undefined);
  /** True when the step is complete. Action and watch steps. */
  done?: (c: StepContext) => boolean;
  /**
   * The common wrong action, and what to say about it. The step stays
   * live: recovery never blocks, it only rewrites the card.
   */
  recover?: { when: (c: StepContext) => boolean; title: Copy; body: Copy; hint?: Copy };
  /** Advance on its own after this long. Watch steps; an action step may. */
  timeoutMs?: number;
  /** What the player must do first, while the board does not allow the step. */
  first?: (c: StepContext) => Detour | null;
  /**
   * Why the board can no longer produce this step's action, or null
   * while it can. A reason advances the step and is logged.
   */
  cannot?: (c: StepContext) => string | null;
  /**
   * Hover steps: done once the pointer has rested on the anchor for
   * `ms`. `event` is the bus event that stands for the whole gesture on
   * a device with no hover, where a touch is all there is. A step with
   * no event (the command zone, ADR 0125 §5.1, whose touch casts the
   * commander) moves on after TOUCH_HOVER_STEP_MS on such a device: the
   * bus keeps its three events (ADR 0076 §2.5).
   */
  hover?: { ms: number; event?: TutorialEvent };
  /**
   * The first-use hints this step teaches (ADR 0125 §5.3). When the
   * step completes they are marked seen at their current version, so
   * the player is not told again at their first real table.
   */
  teaches?: readonly HintID[];
  /**
   * The step teaches the opening roll itself (step 2), so the hold that
   * keeps every other step still while the roll is open does not apply
   * to it (heldByOpeningRoll).
   */
  duringOpeningRoll?: boolean;
}

export interface TutorialSnapshot {
  /** Index into the step list. */
  index: number;
  step: TutorialStep;
  coach: CoachState;
  /** False once the player skipped the tutorial or finished it. */
  visible: boolean;
  /** The step's detour while the board needs one, else null. */
  detour: Detour | null;
  /**
   * The opening roll is open and this step waits for the deal, so it may
   * not move and its timers have not started (heldFor).
   */
  held: boolean;
}

/**
 * heldByOpeningRoll reports whether the opening roll is open (ADR 0121
 * §1). The practice table opens with it (ADR 0125 §5.2), and nothing is
 * dealt until the winner chooses, so no step can be judged yet: the hand
 * is empty, so "Read your hand" would complete on an empty row and "Play
 * a land" would give up for want of a land. While it holds, no step
 * completes, gives up or advances itself, and a step's hint and timeout
 * wait for the deal; the opening card's Start and the player's own Skip
 * step still work. The step that teaches the roll is exempt (heldFor).
 */
export function heldByOpeningRoll(view: GameView | null): boolean {
  return !!view?.opening_roll;
}

/**
 * heldFor is the hold as it applies to one step: every step but the one
 * that teaches the roll waits for the deal.
 */
export function heldFor(step: TutorialStep, view: GameView | null): boolean {
  return !step.duringOpeningRoll && heldByOpeningRoll(view);
}

export interface TutorialRun extends Readable<TutorialSnapshot> {
  /** The opening card's Start. */
  start(): void;
  /** Skip step: silently on to the next one. */
  skipStep(): void;
  /**
   * Skip tutorial (step 1), and Finish and Replay (step 14): hide the
   * coach. On the last step that completes it.
   */
  close(): void;
  /** The game moved, or the bus spoke: check the current step. */
  observe(view: GameView | null, event?: TutorialEvent | null): void;
  /** This tab's client state changed: check the step. */
  observeClient(client: ClientState): void;
  /** The current step's anchor cannot be found: advance and log. */
  anchorMissing(stepID: string): void;
  /** The pointer rested on a hover step's anchor long enough. */
  hovered(stepID: string): void;
  /** Advance past `stepID`, if it is still the current step, and log why. */
  advance(stepID: string, why: string): void;
  /** What the predicates see right now, for anchors and status lines. */
  context(): StepContext;
  /** Stop every timer. */
  destroy(): void;
  /** The current snapshot, synchronously. */
  current(): TutorialSnapshot;
}

export interface RunOptions {
  viewerID: string | null;
  view?: GameView | null;
  /** Where "advanced itself" is logged. Defaults to console.warn. */
  log?: (msg: string) => void;
  hintAfterMs?: number;
  client?: ClientState;
  /**
   * A step completed: its predicate fired, a hover rested, or its button
   * was pressed. Not called for a step that was skipped, gave up, timed
   * out or had no anchor (ADR 0125 §5.3). The coach marks the step's
   * `teaches` seen here.
   */
  onComplete?: (step: TutorialStep) => void;
}

/** coachStateFor is the card's state for a step, before hint and recovery. */
export function coachStateFor(kind: StepKind): CoachState {
  return kind;
}

/** copyText renders a Copy. */
export function copyText(c: Copy | undefined, ctx: CopyContext): string {
  if (c === undefined) return "";
  return typeof c === "function" ? c(ctx) : c;
}

/**
 * anchorsOf normalises a step's anchor field to a list. An anchor read
 * off the board needs the board; without one it names nothing.
 */
export function anchorsOf(step: TutorialStep, c?: StepContext): Anchor[] {
  const spec = step.anchor;
  if (!spec) return [];
  const a = typeof spec === "function" ? (c ? spec(c) : null) : spec;
  if (!a) return [];
  return Array.isArray(a) ? a : [a];
}

/** spotAnchors is what the hole covers now: the detour's anchor, else the step's. */
export function spotAnchors(step: TutorialStep, detour: Detour | null, c?: StepContext): Anchor[] {
  if (detour?.anchor) return Array.isArray(detour.anchor) ? detour.anchor : [detour.anchor];
  return anchorsOf(step, c);
}

/** statusText renders a step's status line; undefined means the default. */
export function statusText(step: TutorialStep, c: StepContext): string | undefined {
  return typeof step.status === "function" ? step.status(c) : step.status;
}

/**
 * spotlit reports whether the scrim should be drawn for a state. Only a
 * step that asks the player to do something points at anything: the
 * opening card leaves the board lit so they can look around, and a
 * watch step has no ask (the player should be reading the strip).
 */
export function spotlit(state: CoachState): boolean {
  return state === "action" || state === "hint" || state === "recovered";
}

/**
 * createTutorialRun starts the machine on the first step of `steps`.
 * The list must start with an `opening` step and end with a `done` one.
 */
export function createTutorialRun(steps: TutorialStep[], opts: RunOptions): TutorialRun {
  if (steps.length === 0) throw new Error("tutorial: no steps");
  const log = opts.log ?? ((m: string) => console.warn(m));
  const hintAfter = opts.hintAfterMs ?? HINT_AFTER_MS;

  let index = 0;
  let coach: CoachState = coachStateFor(steps[0].kind);
  let visible = true;
  let start: GameView | null = opts.view ?? null;
  let latest: GameView | null = opts.view ?? null;
  let client: ClientState = opts.client ?? { autopass: false };
  let hintTimer: ReturnType<typeof setTimeout> | null = null;
  let stepTimer: ReturnType<typeof setTimeout> | null = null;
  let detour: Detour | null = null;

  /** Is the current step waiting for the deal? */
  const held = (): boolean => heldFor(steps[index], latest);

  const snap = (): TutorialSnapshot => ({
    index,
    step: steps[index],
    coach,
    visible,
    detour,
    held: held(),
  });
  const store = guardedWritable<TutorialSnapshot>(snap(), "tutorial");
  const publish = () => store.set(snap());
  const ctxOf = (event: TutorialEvent | null): StepContext => ({
    view: latest,
    start,
    viewerID: opts.viewerID,
    event,
    client,
  });

  function clearTimers(): void {
    if (hintTimer !== null) clearTimeout(hintTimer);
    if (stepTimer !== null) clearTimeout(stepTimer);
    hintTimer = null;
    stepTimer = null;
  }

  function armHint(): void {
    if (hintTimer !== null) clearTimeout(hintTimer);
    hintTimer = null;
    const step = steps[index];
    if (step.kind !== "action" || !step.hint) return;
    hintTimer = setTimeout(() => {
      hintTimer = null;
      // A recovered card has its own hint already; leave it.
      if (visible && coach === "action") {
        coach = "hint";
        publish();
      }
    }, hintAfter);
  }

  function armStepTimer(): void {
    if (stepTimer !== null) clearTimeout(stepTimer);
    stepTimer = null;
    const step = steps[index];
    if ((step.kind !== "action" && step.kind !== "watch") || step.timeoutMs === undefined) return;
    const id = step.id;
    stepTimer = setTimeout(() => {
      stepTimer = null;
      if (visible && steps[index].id === id && !held()) {
        log(`tutorial: step ${id} timed out waiting; advancing`);
        enter(index + 1);
      }
    }, step.timeoutMs);
  }

  /**
   * armTimers starts the current step's hint and timeout from now. A
   * held step starts neither: both begin when the hold lifts, so a step
   * entered during the opening roll gets its whole timeout on the dealt
   * board, not whatever the roll left of it.
   */
  function armTimers(): void {
    clearTimers();
    if (!visible || held()) return;
    armHint();
    armStepTimer();
  }

  /** The current step completed: report it, then move on. */
  function complete(): void {
    report(steps[index]);
    enter(index + 1);
  }

  function report(step: TutorialStep): void {
    try {
      opts.onComplete?.(step);
    } catch (err) {
      log(`tutorial: step ${step.id} completion failed (${String(err)})`);
    }
  }

  function enter(i: number): void {
    clearTimers();
    index = Math.min(i, steps.length - 1);
    const step = steps[index];
    coach = coachStateFor(step.kind);
    start = latest;
    detour = null;
    armTimers();
    publish();
    // A step can already be satisfied the moment it begins.
    check(null);
  }

  function check(event: TutorialEvent | null): void {
    if (!visible) return;
    const step = steps[index];
    if (step.kind !== "action" && step.kind !== "watch") return;
    // Nothing is dealt yet: no step completes or gives up (ADR 0125 §5.2).
    if (held()) return;
    const ctx = ctxOf(event);
    if (step.done?.(ctx)) {
      complete();
      return;
    }
    const why = step.cannot?.(ctx) ?? null;
    if (why !== null) {
      log(`tutorial: step ${step.id} cannot happen (${why}); advancing`);
      enter(index + 1);
      return;
    }
    if (step.recover?.when(ctx) && coach !== "recovered") {
      coach = "recovered";
      if (hintTimer !== null) clearTimeout(hintTimer);
      hintTimer = null;
      // The wrong action changed the board; measure "+1" from here.
      start = latest;
      publish();
    }
    const d = step.first?.(ctx) ?? null;
    if ((d?.id ?? null) !== (detour?.id ?? null)) {
      detour = d;
      publish();
    }
  }

  function advance(stepID: string, why: string): void {
    if (!visible || steps[index].id !== stepID) return;
    if (index >= steps.length - 1) return;
    if (held()) return;
    log(`tutorial: step ${stepID} ${why}; advancing`);
    enter(index + 1);
  }

  return {
    subscribe: store.subscribe,
    current: snap,
    start() {
      if (!visible || steps[index].kind !== "opening") return;
      complete();
    },
    skipStep() {
      if (!visible) return;
      if (index >= steps.length - 1) return;
      enter(index + 1);
    },
    close() {
      if (!visible) return;
      clearTimers();
      visible = false;
      // Finish or Replay on the last step: the player read it to the end.
      if (steps[index].kind === "done") report(steps[index]);
      publish();
    },
    observeClient(c) {
      if (c.autopass === client.autopass) return;
      client = c;
      check(null);
    },
    observe(view, event = null) {
      const wasHeld = held();
      latest = view;
      if (wasHeld !== held()) {
        // The hold lifted (or began). A "+1" predicate measures from the
        // dealt board; the hint and the timeout start now, for their
        // full length (or stop until the deal); and the coach re-arms
        // its touch timer off `held`.
        if (wasHeld) start = latest;
        armTimers();
        publish();
      }
      check(event);
    },
    anchorMissing(stepID) {
      advance(stepID, "has no anchor on the page");
    },
    hovered(stepID) {
      if (!visible || steps[index].id !== stepID || !steps[index].hover) return;
      if (held()) return;
      complete();
    },
    advance,
    context: () => ctxOf(null),
    destroy() {
      clearTimers();
    },
  };
}
