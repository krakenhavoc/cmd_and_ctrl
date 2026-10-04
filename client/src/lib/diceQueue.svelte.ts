// diceQueue — the one dice schedule a game screen shares (ADR 0121 §7).
//
// Two components read it. DiceLayer (in Board) draws each die at its
// roller's seat; RevealBanner (in the attention strip) holds a roll's
// text cue until that die settles. They are in different subtrees, so
// the schedule lives here, provided once by Game.svelte through a
// context and found by both. A component mounted without it (a render
// test, a page that is not a game) gets no queue and falls back to what
// it did before: the banner cues at once.
//
// The rules are in dice.ts, as pure functions. This class only holds
// the state reactively and wakes its readers when something in the
// schedule changes (a die starts, settles, fades, or a cue is
// released), so neither reader needs its own clock.

import { getContext, setContext } from "svelte";
import {
  emptyDiceState,
  isReleased,
  nextDiceChange,
  primeDice,
  trackDice,
  type DiceOptions,
  type DicePlay,
  type DiceRoll,
  type DiceState,
} from "./dice";

const KEY = Symbol("dice-queue");

export class DiceQueue {
  #state: DiceState = emptyDiceState();
  #timer: ReturnType<typeof setTimeout> | null = null;
  /** The schedule. Replaced, never mutated, on every change. */
  plays = $state.raw<DicePlay[]>([]);
  /**
   * Bumped whenever the schedule reaches a new instant. Reading it is
   * how a reader subscribes to "something changed, look again".
   */
  tick = $state(0);
  /** Whether the first frame has been primed. */
  primed = false;

  /** Mark a window as already seen: nothing in it animates or announces. */
  prime(rolls: readonly DiceRoll[]): void {
    this.primed = true;
    this.#state = primeDice(rolls);
    this.#publish(Date.now());
  }

  /** Fold one frame's rolls (the whole window, log order) into the schedule. */
  ingest(rolls: readonly DiceRoll[], now: number, opts: DiceOptions): void {
    if (!this.primed) {
      this.prime(rolls);
      return;
    }
    this.#state = trackDice(this.#state, rolls, now, opts);
    this.#publish(now);
  }

  /** Whether the strip may show `seq`'s cue yet. Subscribes the caller. */
  released(seq: number, now = Date.now()): boolean {
    void this.tick;
    return isReleased(this.#state, seq, now);
  }

  /** Whether `seq` was in the window when it was primed: never announced. */
  wasPrimed(seq: number): boolean {
    return this.#state.primed.has(seq);
  }

  dispose(): void {
    if (this.#timer !== null) clearTimeout(this.#timer);
    this.#timer = null;
  }

  #publish(now: number): void {
    this.plays = this.#state.plays;
    this.tick++;
    this.#arm(now);
  }

  // One timer, for the schedule's next instant. Each wake bumps `tick`
  // and arms the next, so a reader re-evaluates exactly when a die
  // settles or a cue is released, and nothing runs once it is idle.
  #arm(now: number): void {
    if (this.#timer !== null) clearTimeout(this.#timer);
    this.#timer = null;
    const next = nextDiceChange(this.#state, now);
    if (next === null) return;
    this.#timer = setTimeout(
      () => {
        this.#timer = null;
        const t = Date.now();
        this.#state = { ...this.#state, plays: this.#state.plays.filter((p) => p.endAt > t) };
        this.#publish(t);
      },
      Math.max(0, next - now),
    );
  }
}

/** Game.svelte: one queue for the whole game screen. */
export function provideDiceQueue(): DiceQueue {
  const q = new DiceQueue();
  setContext(KEY, q);
  return q;
}

/** The game screen's queue, or null outside one. */
export function useDiceQueue(): DiceQueue | null {
  return getContext<DiceQueue | undefined>(KEY) ?? null;
}
