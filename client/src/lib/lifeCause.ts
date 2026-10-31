// lifeCause — what made a life total move, for the life popup (#3379).
//
// The popup (#703, lifePopup.ts) floats each new delta over the seat's
// avatar, but a bare "-1" does not say whether the player paid for a
// fetch land, took a trigger or was hit by a creature. The public log
// does, in two kinds of entry: a `life` entry (`seat` is the player,
// `card_id` the source the engine credits: drains, gains, lifelink,
// life paid as a cost) and a `damage` entry whose `target_seat` is the
// player (damage moves the total with no `life` entry of its own).
//
// THE PAIRING. `life_history` and the log are two records of the same
// changes with different counters, so a popup's deltas are paired with
// the seat's newest logged changes in order, and only when every amount
// matches. Anything else (the log window has moved on, a sandbox edit
// with no source, an entry the frame does not carry) pairs nothing and
// the popup stays a bare number, as before.
//
// THE NAME comes from a card this viewer can see in the frame, never
// from the log line. The server keeps `card_id` on the wire and drops
// only the NAME for a viewer who does not know the card; looking the
// ID up in the view goes through the same knower filter every CardView
// already went through, so a hidden or face-down source stays unnamed.
//
// Plain functions over plain data, unit-tested without a component.
// Board.svelte provides the lookup as a context; PlayerIdentity.svelte
// asks it when a popup goes up.

import { getContext, setContext } from "svelte";
import { popupCombatDamage } from "./faceHit";
import { popupIsExtort } from "./extort";
import type { CardView, GameView, LogEvent } from "./protocol";

function namedIn(cards: readonly CardView[] | undefined, id: string): string | null {
  const c = cards?.find((x) => x.instance_id === id);
  if (!c || c.face_down || !c.name) return null;
  return c.name;
}

/**
 * The name of a card this viewer can see in the frame, wherever it is,
 * or null: not in a visible zone, face down, or nameless for this
 * viewer.
 */
export function visibleCardName(view: GameView, id: string): string | null {
  const zones: (readonly CardView[] | undefined)[] = [
    view.battlefield?.cards,
    view.stack?.cards,
    view.exile?.cards,
    view.phased_out?.cards,
  ];
  for (const s of view.seats ?? []) {
    zones.push(s.graveyard?.cards, s.command?.cards, s.hand?.cards);
  }
  for (const cards of zones) {
    const name = namedIn(cards, id);
    if (name) return name;
  }
  return null;
}

/** One change to a seat's life total as the log tells it: the signed amount and its source. */
interface LoggedChange {
  delta: number;
  source: string | null;
}

/**
 * Every change to one seat's life total the log carries, oldest first.
 * Two kinds: a `life` entry (a drain, a gain, lifelink, life paid as a
 * cost), and a `damage` entry that hit the player, because damage moves
 * the total without a `life` entry of its own (damage_tail.go changes
 * it directly).
 */
function loggedChanges(log: readonly LogEvent[], seatIndex: number): LoggedChange[] {
  const out: LoggedChange[] = [];
  for (const e of log) {
    const amount = e.amount ?? 0;
    if (e.kind === "life" && e.seat === seatIndex && amount !== 0) {
      out.push({ delta: amount, source: e.card_id ?? null });
    } else if (e.kind === "damage" && e.target_seat === seatIndex && amount > 0) {
      out.push({ delta: -amount, source: e.card_id ?? null });
    }
  }
  return out;
}

/**
 * lifeCauses names the source of each of `deltas` (a popup's deltas,
 * oldest first) for the player in seat `seatIndex`, or null for one it
 * cannot name. All null unless the seat's newest logged changes match
 * the deltas one for one.
 */
export function lifeCauses(
  view: GameView,
  seatIndex: number,
  deltas: readonly number[],
): (string | null)[] {
  const none = deltas.map(() => null);
  if (seatIndex < 0 || deltas.length === 0) return none;
  const mine = loggedChanges(view.log ?? [], seatIndex);
  if (mine.length < deltas.length) return none;
  const newest = mine.slice(mine.length - deltas.length);
  if (newest.some((c, k) => c.delta !== deltas[k])) return none;
  return newest.map((c) => (c.source ? visibleCardName(view, c.source) : null));
}

// ---- the context ---------------------------------------------------------

/** Names the sources of a seat's popup deltas. */
export type LifeCauseLookup = (seatID: string, deltas: readonly number[]) => (string | null)[];

const KEY = Symbol("lifeCauses");

/** provideLifeCauses is the board's: call it during component init with a getter for the live view. */
export function provideLifeCauses(getView: () => GameView): void {
  const lookup: LifeCauseLookup = (seatID, deltas) => {
    const view = getView();
    return lifeCauses(
      view,
      (view.seats ?? []).findIndex((s) => s.id === seatID),
      deltas,
    );
  };
  setContext(KEY, lookup);
  // #4432, #4543: the combat damage to the seat that a popup holds.
  const combat: CombatHitLookup = (seatID, deltas) => {
    const view = getView();
    return popupCombatDamage(
      view,
      (view.seats ?? []).findIndex((s) => s.id === seatID),
      deltas,
    );
  };
  setContext(HIT_KEY, combat);
  // #4544: whether a popup shows an extort drain, which the `extort` cue sounds.
  const extort: ExtortDrainLookup = (seatID, deltas) => {
    const view = getView();
    return popupIsExtort(
      view,
      (view.seats ?? []).findIndex((s) => s.id === seatID),
      deltas,
    );
  };
  setContext(EXTORT_KEY, extort);
}

/** The combat damage to a seat that its popup deltas hold, 0 for none. */
export type CombatHitLookup = (seatID: string, deltas: readonly number[]) => number;

const HIT_KEY = Symbol("combatHits");
const EXTORT_KEY = Symbol("extortDrains");

/** Whether a seat's popup deltas are an extort drain (#4544). */
export type ExtortDrainLookup = (seatID: string, deltas: readonly number[]) => boolean;

/** extortDrainLookup is the avatar's: whether a seat's popup deltas are an extort drain. Undefined with no board above. */
export function extortDrainLookup(): ExtortDrainLookup | undefined {
  return getContext<ExtortDrainLookup | undefined>(EXTORT_KEY);
}

/** combatHitLookup is the avatar's: call it during component init. Undefined with no board above. */
export function combatHitLookup(): CombatHitLookup | undefined {
  return getContext<CombatHitLookup | undefined>(HIT_KEY);
}

/** lifeCauseLookup is the avatar's: call it during component init. Undefined with no board above. */
export function lifeCauseLookup(): LifeCauseLookup | undefined {
  return getContext<LifeCauseLookup | undefined>(KEY);
}
