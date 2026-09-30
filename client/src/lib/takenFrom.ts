// takenFrom.ts — ADR 0104 (owner decision 6): a permanent controlled
// by a player who does not own it says whose it is.
//
// General on purpose. The permanent a stolen spell becomes is the case
// that asked for it, but an Act of Treason, a Mind Control and a
// Switcheroo leave a permanent in exactly the same state — drawn in
// its controller's row with nothing saying it belongs to someone else —
// so the chip is derived from the two fields every CardView already
// carries rather than from how the permanent got there.

import type { CardView, PlayerView } from "./protocol";

/** Owner name keyed by instance ID, for every permanent whose controller is not its owner. */
export function takenFromByCard(
  cards: readonly CardView[] | undefined,
  seats: readonly PlayerView[] | undefined,
): Record<string, string> {
  const names = new Map((seats ?? []).map((s) => [s.id, s.name]));
  const out: Record<string, string> = {};
  for (const c of cards ?? []) {
    if (!c.owner || !c.controller || c.owner === c.controller) continue;
    out[c.instance_id] = names.get(c.owner) ?? "another player";
  }
  return out;
}
