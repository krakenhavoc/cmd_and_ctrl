// Stable per-seat color palette. Indexed by absolute seat number
// (PlayerView.seat), not the viewer-relative SeatPosition used by the
// PixiJS renderer — chat author badges and the priority indicator
// must look the same to every viewer, so seat 0 is always blue
// regardless of who is looking. The palette is hand-tuned to stay
// legible on the dark `.dev-controls` background (#1a2540) and to
// remain distinguishable for the most common forms of color blindness.
const SEAT_PALETTE: readonly string[] = [
  "#5fb0ff", // seat 0 — blue
  "#c98bff", // seat 1 — purple
  "#ffb45f", // seat 2 — orange
  "#5fd4a4", // seat 3 — green
];

const SPECTATOR_COLOR = "#888888";

// seatColor returns a CSS hex string for the given seat index. Out-of-
// range or negative seats fall back to the spectator gray rather than
// throwing — calling code may legitimately render a spectator chat
// entry with a synthetic seat of -1.
export function seatColor(seat: number): string {
  if (!Number.isInteger(seat) || seat < 0 || seat >= SEAT_PALETTE.length) {
    return SPECTATOR_COLOR;
  }
  return SEAT_PALETTE[seat];
}
