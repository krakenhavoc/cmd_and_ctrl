// combatCues — the one combat damage clock a board shares (ADR 0134 §1).
//
// ADR 0053's BeatDirector used to live inside CombatArrows.svelte. Two
// layers now read the same beats: CombatArrows (the pulses, the ghosts
// and the "First strike" text cue) and CombatStrikes (attackers lunge,
// hits land). One tracker and one clock mean the label and the lunge it
// describes can never disagree, so Board.svelte creates one of these and
// both layers subscribe to it, as DiceQueue is shared (ADR 0121).
//
// It also holds the set of creatures whose art-only copy is in flight:
// Card.svelte reads it, through a context, to hide the live tile for the
// copy's flight only (ADR 0134 §2).

import { getContext, setContext } from "svelte";
import { SvelteMap } from "svelte/reactivity";
import { BeatDirector, beatMode, type FramePlan, type ScheduledCue } from "./combatBeats";
import { combatMotion, strikesScheduled } from "./combatStrikes";
import type { LogEvent } from "./protocol";

const KEY = Symbol("combat-cues");

/** What a layer hears from the shared clock. Every handler is optional. */
export interface CueListener {
  /** A beat is due now. */
  onCue?(cue: ScheduledCue): void;
  /** The text cue for this step comes down. */
  onHide?(stepSeq: number): void;
  /** A priming frame cancelled every pending cue: clear the screen. */
  onReset?(): void;
  /** A frame has been folded in (after its cues were scheduled). */
  onFrame?(plan: FramePlan): void;
}

/** The settings a frame is planned with, read from the settings store. */
export interface CueSettings {
  animations: { enabled: boolean; speed: number; damagePopups: boolean; combat: boolean };
  accessibility: { reduceMotion: boolean };
}

export class CombatCues {
  readonly #listeners = new Set<CueListener>();
  readonly #director: BeatDirector;
  #reprimeNext = false;
  // Instance ID → how many copies of it are in flight. Per-key reactive,
  // so a Card re-renders only when its own ID changes.
  readonly #striking = new SvelteMap<string, number>();

  constructor() {
    this.#director = new BeatDirector({
      onCue: (cue) => {
        for (const l of [...this.#listeners]) l.onCue?.(cue);
      },
      onHide: (stepSeq) => {
        for (const l of [...this.#listeners]) l.onHide?.(stepSeq);
      },
      onReset: () => {
        this.#striking.clear();
        for (const l of [...this.#listeners]) l.onReset?.();
      },
    });
  }

  /** Listen to the clock. Returns the unsubscribe. */
  subscribe(listener: CueListener): () => void {
    this.#listeners.add(listener);
    return () => this.#listeners.delete(listener);
  }

  /**
   * The next frame primes, as a first frame does: a reconnect, a
   * replay toggle, or a replay move that is not a one-frame step
   * forward (Game.svelte's beatsPrimeKey, ADR 0134 §6).
   */
  requestReprime(): void {
    this.#reprimeNext = true;
  }

  /**
   * Fold one frame's log in and schedule its beats. Settings are read
   * at plan time. A normal frame never cancels a cue; a priming frame
   * cancels them all.
   */
  frame(
    log: readonly LogEvent[] | undefined,
    s: CueSettings,
    visibility: string | undefined = typeof document === "undefined"
      ? undefined
      : document.visibilityState,
  ): FramePlan {
    const motion = combatMotion({
      enabled: s.animations.enabled,
      combat: s.animations.combat,
      reduceMotion: s.accessibility.reduceMotion,
    });
    const plan = this.#director.frame(log, {
      mode: beatMode(s.animations.enabled, s.animations.damagePopups, s.accessibility.reduceMotion),
      speed: s.animations.speed,
      reprime: this.#reprimeNext,
      combatMotion: strikesScheduled(motion, visibility),
    });
    this.#reprimeNext = false;
    for (const l of [...this.#listeners]) l.onFrame?.(plan);
    return plan;
  }

  /** Cues scheduled or playing right now. */
  get pending(): number {
    return this.#director.pending;
  }

  /** Whether a copy of this creature is in flight. Subscribes the caller. */
  isStriking(instanceID: string): boolean {
    return (this.#striking.get(instanceID) ?? 0) > 0;
  }

  /** A copy of this creature took off. */
  strikeStart(instanceID: string): void {
    this.#striking.set(instanceID, (this.#striking.get(instanceID) ?? 0) + 1);
  }

  /** A copy of this creature is gone. */
  strikeEnd(instanceID: string): void {
    const n = (this.#striking.get(instanceID) ?? 0) - 1;
    if (n > 0) this.#striking.set(instanceID, n);
    else this.#striking.delete(instanceID);
  }

  /** Every copy is gone at once. */
  clearStriking(): void {
    this.#striking.clear();
  }

  dispose(): void {
    this.#director.dispose();
    this.#listeners.clear();
    this.#striking.clear();
  }
}

/**
 * Board.svelte: one clock for the whole board. A render test may hand in
 * its own, to drive it from outside.
 */
export function provideCombatCues(cues: CombatCues = new CombatCues()): CombatCues {
  setContext(KEY, cues);
  return cues;
}

/** The board's clock, or null outside a board. */
export function useCombatCues(): CombatCues | null {
  return getContext<CombatCues | undefined>(KEY) ?? null;
}
