import type { PlayerView } from "./protocol";

// Mulligan decisions go in turn order (CR 103.5, #2237): the server
// marks the one seat whose turn it is with `mulligan_turn`, and refuses a
// keep or mulligan from any other seat.

type MulliganSeat = Pick<PlayerView, "name" | "eliminated" | "hand_kept" | "mulligan_turn">;

/** The seat deciding now, or null when none is (no window, or all done). */
export function mulliganDecider<T extends MulliganSeat>(seats: T[]): T | null {
  return seats.find((s) => s.mulligan_turn === true && !s.eliminated && !s.hand_kept) ?? null;
}

/** What a seat that is not deciding is told: "Waiting for X to decide". */
export function mulliganWaitingText(
  seats: MulliganSeat[],
  viewerIsDeciding: boolean,
): string | null {
  if (viewerIsDeciding) return null;
  const who = mulliganDecider(seats);
  return who ? `Waiting for ${who.name} to decide` : null;
}

// The London mulligan (CR 103.5, ADR 0168): a mulligan draws seven, and
// a keep puts `mulligan_bottom` cards on the bottom of the library. The
// first mulligan of a game that began with more than two players is free
// (CR 103.5c), and so is a two-player table's when the host has turned on
// the free first mulligan (#3387), so it adds nothing to the count.

function cards(n: number): string {
  return n === 1 ? "1 card" : `${n} cards`;
}

/** What keeping costs now, or null when it costs nothing. */
export function keepCostText(bottom: number): string | null {
  return bottom > 0
    ? `Keep: put ${cards(bottom)} from this hand on the bottom of your library.`
    : null;
}

/** How many cards one more mulligan would have the seat bottom at keep. */
export function nextMulliganBottom(taken: number, bottomNow: number, freeFirst: boolean): number {
  const free = taken === 0 && freeFirst;
  return free ? 0 : bottomNow + 1;
}

/** What one more mulligan costs, said under its button. */
export function mulliganCostText(taken: number, bottomNow: number, freeFirst: boolean): string {
  const next = nextMulliganBottom(taken, bottomNow, freeFirst);
  if (next === 0) {
    return "Mulligan: your first mulligan is free. Shuffle this hand away and draw 7.";
  }
  return `Mulligan: shuffle this hand away, draw 7, then put ${cards(next)} on the bottom when you keep.`;
}
