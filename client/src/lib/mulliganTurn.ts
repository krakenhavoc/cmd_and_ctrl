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
