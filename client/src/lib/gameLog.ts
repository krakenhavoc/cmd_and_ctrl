// gameLog.ts — presentation logic for the public game log (S31
// sub-PR 0). Pure functions, unit-tested; GameLogPanel.svelte is the
// thin rendering shell over them.
//
// The server hands the client an already-rendered, already-redacted
// `text` per entry (protocol/log.go). Nothing here reconstructs a
// sentence or looks up a card name — an entry that reads "a card" is
// one this viewer is not entitled to see named, and dressing it up
// from the board state would walk straight around the visibility
// filter the log went through.

import type { GameView, LogEvent, LogKind, PlayerView } from "./protocol";

// LogGroup is one turn/step block: the `step` entry's text as a
// header, and everything the engine emitted inside that step.
export interface LogGroup {
  // Stable key for the keyed each — the seq of the step entry, or 0
  // for the pre-history block.
  key: number;
  // Header line ("Turn 7 — Aang · precombat main"), or null for
  // entries that arrived before the first step announcement in the
  // window the ring happens to hold.
  header: string | null;
  // Seat index the header's turn belongs to, for the accent colour.
  seat: number;
  entries: LogEvent[];
}

// groupLog folds a flat log into turn/step blocks, oldest first.
//
// A ring buffer can start mid-turn, so the first block may have no
// header at all; that is normal, not an error, and it renders as a
// plain "earlier" block rather than being dropped.
export function groupLog(entries: readonly LogEvent[] | undefined): LogGroup[] {
  if (!entries || entries.length === 0) return [];
  const out: LogGroup[] = [];
  let current: LogGroup | null = null;
  for (const e of entries) {
    if (e.kind === "step") {
      current = { key: e.seq, header: e.text, seat: e.seat, entries: [] };
      out.push(current);
      continue;
    }
    if (!current) {
      current = { key: 0, header: null, seat: -1, entries: [] };
      out.push(current);
    }
    current.entries.push(e);
  }
  // A step that announced and produced nothing is noise in a panel —
  // a four-player turn is twelve steps and most of them are empty.
  // The most recent block survives even when empty, because it is
  // what says where the game currently is.
  return out.filter((g, i) => g.entries.length > 0 || i === out.length - 1);
}

// LOG_TONE maps a kind onto the panel's accent classes. Kept as data
// so the stylesheet and the logic can't drift apart silently.
const LOG_TONE: Record<LogKind, string> = {
  step: "tone-step",
  cast: "tone-cast",
  resolve: "tone-resolve",
  fizzle: "tone-bad",
  counter: "tone-bad",
  zone: "tone-zone",
  draw: "tone-quiet",
  life: "tone-life",
  damage: "tone-damage",
  attack: "tone-damage",
  block: "tone-combat",
  token: "tone-zone",
  sacrifice: "tone-bad",
  eliminated: "tone-bad",
};

export function logTone(kind: LogKind): string {
  return LOG_TONE[kind] ?? "tone-quiet";
}

// seatName resolves a seat index to a display name, or null when the
// index names nobody (the -1 "no actor" sentinel, or a seat the view
// no longer carries).
export function seatName(view: GameView | null, seat: number | undefined): string | null {
  if (!view || seat === undefined || seat < 0) return null;
  const p: PlayerView | undefined = view.seats[seat];
  return p ? p.name : null;
}

// entrySeats lists the seat indices an entry involves, for the
// "only show my seat" filter and for accent colouring.
export function entrySeats(e: LogEvent): number[] {
  const seats: number[] = [];
  if (e.seat >= 0) seats.push(e.seat);
  if (e.target_seat !== undefined && e.target_seat >= 0 && !seats.includes(e.target_seat)) {
    seats.push(e.target_seat);
  }
  return seats;
}

// filterBySeat narrows a log to the entries that involve `seat`.
// Step entries always survive — without them the remaining lines lose
// the only thing that says when they happened.
export function filterBySeat(entries: readonly LogEvent[], seat: number | null): LogEvent[] {
  if (seat === null || seat < 0) return [...entries];
  return entries.filter((e) => e.kind === "step" || entrySeats(e).includes(seat));
}
