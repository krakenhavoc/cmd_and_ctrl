import type { ActivatedAbilityView, CardView } from "./protocol";

// crew.ts — the arithmetic behind a Vehicle's crew cost
// (CR 702.122a), kept out of the modal so it can be tested without a
// component harness.
//
// Crew is the only cost in the catalog that is counted rather than
// itemised. Every other cost the client collects is "pick exactly N
// things" (sacrifice, discard) or "pick at most N things" (convoke).
// Crew is "pick any number of things whose POWER adds up to at least
// N", which means the picker needs a running total, a floor, and a
// way to say up front that the board cannot reach it.
//
// Power is read off the CardView, which already carries the
// post-layer effective value the server re-checks against — so the
// number shown here and the number the server computes agree without
// a second round trip.

// crewPower is the total power of the chosen creatures. A chosen ID
// that is not in `options` contributes nothing rather than NaN: the
// snapshot can change under an open prompt, and a creature that has
// left the board should shrink the total, not poison it.
export function crewPower(options: CardView[], chosen: readonly string[]): number {
  let total = 0;
  for (const id of chosen) {
    const c = options.find((o) => o.instance_id === id);
    total += c?.power ?? 0;
  }
  return total;
}

// crewSatisfied reports whether the chosen creatures can pay the
// cost. Overshooting is legal — the printed number is a floor, and a
// single 5-power creature crews a Vehicle that says crew 3 — but
// picking nothing is not, even for a hypothetical crew 0, because
// the cost still says to tap creatures.
export function crewSatisfied(
  options: CardView[],
  chosen: readonly string[],
  need: number,
): boolean {
  if (chosen.length === 0) return false;
  return crewPower(options, chosen) >= need;
}

// crewAvailablePower is the best the whole offered roster could do.
// Below the crew number, no selection can ever pay the cost, and the
// UI should say so rather than let a player hunt for the creature
// that would finish the total.
export function crewAvailablePower(options: CardView[]): number {
  return options.reduce((sum, c) => sum + (c.power ?? 0), 0);
}

// crewOptionsFor narrows a battlefield roster to the creatures the
// server offered for this ability's crew cost. One place so the
// modal and any future caller can't disagree about what the prompt
// is looking at.
//
// Note what is NOT filtered here: summoning sickness. Tapping a
// creature to crew is not paying a {T} cost, so a creature cast this
// turn may crew (CR 702.122b), and the server's crew_options list
// includes them for that reason.
export function crewOptionsFor(a: ActivatedAbilityView, board: CardView[]): CardView[] {
  const ids = new Set(a.crew_options?.cards ?? []);
  return board.filter((c) => ids.has(c.instance_id));
}
