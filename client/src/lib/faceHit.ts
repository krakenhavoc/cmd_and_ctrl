// faceHit — the sound of combat damage landing on a player (#4432,
// ADR 0209).
//
// A creature that hits a player's face should sound like something
// smashed into them, not like a drain. Two surfaces could play a sound
// for the same hit, so the choice is made here, once, and both read it:
//
//   - the strike layer (CombatStrikes.svelte) plays `face_hit` at the
//     beat's first contact (`tl.contactMs`), once per player the beat
//     hit, so the sound lands with the lunge;
//   - the life popup (PlayerIdentity.svelte) plays `damage` for any
//     loss, so it stands down for a combat hit and lets `face_hit` own
//     it. When no strike is scheduled (motion off, reduced motion, a
//     hidden tab) nothing lunges, so the popup plays `face_hit` itself
//     the moment the loss shows, the same split combat_resolve has
//     between the step change and the strike layer.
//
// Non-combat life loss (a drain, life paid, noncombat damage) is not a
// combat hit and keeps `damage` from the popup.
//
// Plain functions over plain data, unit-tested without a component.

import type { GameView, LogEvent } from "./protocol";

// FACE_HIT_STAGGER_MS separates the blows when one beat hits several
// players. sounds.ts drops a repeat of a name inside its 40 ms rate limit,
// so simultaneous plays would collapse into one.
export const FACE_HIT_STAGGER_MS = 90;

/**
 * The seats a set of log entries hit with combat damage, each once, in
 * log order. A `damage` entry names a player in `target_seat` (a card
 * it hit is `target`). Zero damage (fully prevented) is not a hit.
 */
export function faceHitSeats(entries: readonly LogEvent[]): number[] {
  const seats: number[] = [];
  for (const e of entries) {
    if (e.kind !== "damage" || !e.combat) continue;
    if (e.target_seat === undefined || e.target_seat === null || e.target_seat < 0) continue;
    if ((e.amount ?? 0) <= 0) continue;
    if (!seats.includes(e.target_seat)) seats.push(e.target_seat);
  }
  return seats;
}

/**
 * Whether a popup's life changes include combat damage to `seatIndex`.
 * `deltas` is the popup's, oldest first; they are matched against the
 * seat's newest logged changes one for one, as lifeCause.ts does, and
 * false when they do not line up (then the popup keeps `damage`).
 */
export function popupHasCombatHit(
  view: GameView,
  seatIndex: number,
  deltas: readonly number[],
): boolean {
  if (seatIndex < 0 || deltas.length === 0) return false;
  const mine: { delta: number; combat: boolean }[] = [];
  for (const e of view.log ?? []) {
    const amount = e.amount ?? 0;
    if (e.kind === "life" && e.seat === seatIndex && amount !== 0) {
      mine.push({ delta: amount, combat: false });
    } else if (e.kind === "damage" && e.target_seat === seatIndex && amount > 0) {
      mine.push({ delta: -amount, combat: e.combat === true });
    }
  }
  if (mine.length < deltas.length) return false;
  const newest = mine.slice(mine.length - deltas.length);
  if (newest.some((c, k) => c.delta !== deltas[k])) return false;
  return newest.some((c) => c.combat);
}

/**
 * The sound a life popup plays, or null for none. `base` is lifePopup's
 * own choice ("damage" or "heal"), `combatHit` whether the loss includes
 * combat damage to the seat, and `strikesPlay` whether the strike layer
 * is scheduled to play `face_hit` at contact (strikesScheduled). A combat
 * hit is either left to the strike layer (silent here) or, when no strike
 * plays, heard now as `face_hit`; every other loss keeps `damage`.
 */
export function popupSound(
  base: "damage" | "heal",
  combatHit: boolean,
  strikesPlay: boolean,
): "damage" | "heal" | "face_hit" | null {
  if (base === "heal" || !combatHit) return base;
  return strikesPlay ? null : "face_hit";
}
