// tutorial.ts — the tutorial's step machine (ADR 0076 §2.1, §2.3, §2.4;
// #1079, sub-PR 3).
//
// Pure state, no DOM: TutorialCoach.svelte feeds it the game view, the
// tutorialBus events and "the anchor is missing", and draws what it says.
// The step list is data (tutorialSteps.ts), so sub-PR 4 (#1081) adds the
// nine middle steps without touching this file.
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
//   - A step that waits on the bot rather than the player (step 9) can
//     carry a timeout, and advances on its own when it runs out.

import type { Readable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type { GameView } from "./protocol";
import type { TutorialEvent } from "./tutorialBus";

/** How many steps the tutorial has (ADR 0076 §2.1), for "N / 11". */
export const TUTORIAL_STEP_COUNT = 11;

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
 * What a spotlight points at. Anchors are the aria-labels the e2e suite
 * asserts on (ADR 0076 §2.4), so renaming one breaks a test before it
 * breaks the tutorial.
 *
 * `within` scopes the label to a labelled container. It is not optional
 * in practice for the battlefield rows: every opponent's panel carries
 * the same "lands" and "creatures" lists as the viewer's, so those
 * anchors are `{ label: "lands", within: "your board" }`.
 */
export type Anchor = { label: string; within?: string } | { cardID: string };

/**
 * The step kinds. `opening` and `done` are the two button steps (1 and
 * 11); `action` waits on the player; `watch` waits on the bot.
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
}

/** Copy that can name the player's own key bindings. */
export interface CopyContext {
  /** The chord that opens the keymap overlay, formatted; "" if unbound. */
  helpKey: string;
  /** The chord that opens settings, formatted; "" if unbound. */
  settingsKey: string;
}

export type Copy = string | ((c: CopyContext) => string);

export interface TutorialStep {
  /** Stable id, for logs and tests. */
  id: string;
  /** Its number in ADR 0076 §2.1's table (1–11). */
  n: number;
  kind: StepKind;
  title: Copy;
  body: Copy;
  /** One anchor, or several whose rects the hole covers together. */
  anchor?: Anchor | Anchor[];
  /** The second thing they get (action steps). */
  hint?: Copy;
  /** The status line, e.g. "Bot is thinking". Defaults by kind. */
  status?: string;
  /** True when the step is complete. Action and watch steps. */
  done?: (c: StepContext) => boolean;
  /**
   * The common wrong action, and what to say about it. The step stays
   * live: recovery never blocks, it only rewrites the card.
   */
  recover?: { when: (c: StepContext) => boolean; title: Copy; body: Copy; hint?: Copy };
  /** Watch steps: advance on their own after this long. */
  timeoutMs?: number;
}

export interface TutorialSnapshot {
  /** Index into the step list. */
  index: number;
  step: TutorialStep;
  coach: CoachState;
  /** False once the player skipped the tutorial or finished it. */
  visible: boolean;
}

export interface TutorialRun extends Readable<TutorialSnapshot> {
  /** The opening card's Start. */
  start(): void;
  /** Skip step: silently on to the next one. */
  skipStep(): void;
  /** Skip tutorial (step 1) and Finish (step 11): hide the coach. */
  close(): void;
  /** The game moved, or the bus spoke: check the current step. */
  observe(view: GameView | null, event?: TutorialEvent | null): void;
  /** The current step's anchor cannot be found: advance and log. */
  anchorMissing(stepID: string): void;
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

/** anchorsOf normalises a step's anchor field to a list. */
export function anchorsOf(step: TutorialStep): Anchor[] {
  if (!step.anchor) return [];
  return Array.isArray(step.anchor) ? step.anchor : [step.anchor];
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
  let hintTimer: ReturnType<typeof setTimeout> | null = null;
  let stepTimer: ReturnType<typeof setTimeout> | null = null;

  const snap = (): TutorialSnapshot => ({ index, step: steps[index], coach, visible });
  const store = guardedWritable<TutorialSnapshot>(snap(), "tutorial");
  const publish = () => store.set(snap());

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

  function enter(i: number): void {
    clearTimers();
    index = Math.min(i, steps.length - 1);
    const step = steps[index];
    coach = coachStateFor(step.kind);
    start = latest;
    armHint();
    if (step.kind === "watch" && step.timeoutMs !== undefined) {
      const id = step.id;
      stepTimer = setTimeout(() => {
        stepTimer = null;
        if (visible && steps[index].id === id) {
          log(`tutorial: step ${id} timed out waiting; advancing`);
          enter(index + 1);
        }
      }, step.timeoutMs);
    }
    publish();
    // A step can already be satisfied the moment it begins.
    check(null);
  }

  function check(event: TutorialEvent | null): void {
    if (!visible) return;
    const step = steps[index];
    if (step.kind !== "action" && step.kind !== "watch") return;
    const ctx: StepContext = { view: latest, start, viewerID: opts.viewerID, event };
    if (step.done?.(ctx)) {
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
  }

  return {
    subscribe: store.subscribe,
    current: snap,
    start() {
      if (!visible || steps[index].kind !== "opening") return;
      enter(index + 1);
    },
    skipStep() {
      if (!visible) return;
      if (index >= steps.length - 1) return;
      enter(index + 1);
    },
    close() {
      clearTimers();
      visible = false;
      publish();
    },
    observe(view, event = null) {
      latest = view;
      check(event);
    },
    anchorMissing(stepID) {
      if (!visible || steps[index].id !== stepID) return;
      log(`tutorial: step ${stepID} has no anchor on the page; advancing`);
      enter(index + 1);
    },
    destroy() {
      clearTimers();
    },
  };
}
