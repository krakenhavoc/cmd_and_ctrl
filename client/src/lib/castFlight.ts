// castFlight — a spell travelling to the stack from where it was cast
// (#3381).
//
// The stack linger (ADR 0119 §3) shows a spell LEAVING the stack and
// flying to where it went. Arriving, it just appeared in the stack
// panel: an opponent's spell gave no sign of whose hand it came out of,
// and a cast from a graveyard (flashback), from exile (warp, foretell,
// adventure) or from the command zone looked the same as one from hand.
//
// So a copy of the spell starts where it came from, small, and grows
// into its place on the stack:
//
//   - from a hand: the caster's hand zone (`data-hand-owner`), else the
//     caster's seat;
//   - from a graveyard, exile or a library: the caster's pile of that
//     name, else the seat;
//   - from the command zone, or anywhere else: the caster's seat.
//
// DISPLAY ONLY, like the linger: the board already shows the spell on
// the stack. THE SIGNAL is the public log's `cast` entry (its `card_id`
// is the spell's stack item ID; `old_zone` is set only when the origin
// is not a hand). A spell that is no longer on the stack by the frame
// that brings its cast entry (it resolved in the same frame) does not
// fly. PRIMING: the first frame, a reconnect, a replay toggle or an
// undo flies nothing (the combatBeats.ts watermark).
//
// Plain functions over plain data, unit-tested without a component.
// CastFlight.svelte measures, draws and flies.

import type { GameView, LogEvent } from "./protocol";
import { seatSelector } from "./boardAnchor";
import { cardImageURL } from "./cardImage";
import { emptyBeatTracker, track, type BeatTracker } from "./combatBeats";
import { pileSelector } from "./stackLinger";

/** The flight, before animations.speed. */
export const CAST_FLIGHT_MS = 420;
/** At most this many casts fly from one frame; a burst beyond it just appears. */
export const CAST_FLIGHT_MAX = 4;

export type CastOrigin = "hand" | "graveyard" | "exile" | "library" | "command" | "other";

export interface CastFlight {
  /** The spell's stack item ID (its card's instance ID). */
  itemID: string;
  /** The caster's seat ID. */
  seat: string;
  from: CastOrigin;
  /** The spell's face, or a card back when this viewer may not see it. */
  image: string;
  /** The log entry's seq, for ordering. */
  seq: number;
}

export interface CastTracker {
  log: BeatTracker;
}

export function emptyCastTracker(): CastTracker {
  return { log: emptyBeatTracker() };
}

export interface CastFrame {
  tracker: CastTracker;
  flights: CastFlight[];
}

function originOf(e: LogEvent): CastOrigin {
  switch (e.old_zone) {
    case undefined:
    case "":
    case "hand":
      return "hand";
    case "graveyard":
    case "exile":
    case "library":
    case "command":
      return e.old_zone;
    default:
      return "other";
  }
}

/**
 * trackCasts folds one frame into the tracker: the spells the frame's
 * new `cast` entries put on the stack, oldest first, that are still on
 * it. The input tracker is not mutated.
 */
export function trackCasts(
  prev: CastTracker,
  view: GameView,
  opts: { reprime?: boolean } = {},
): CastFrame {
  const logged = track(prev.log, view.log, opts.reprime === true);
  const tracker: CastTracker = { log: logged.tracker };
  if (logged.primed || logged.rewound) return { tracker, flights: [] };
  const onStack = new Map((view.stack?.cards ?? []).map((c) => [c.instance_id, c]));
  const flights: CastFlight[] = [];
  for (const e of logged.fresh) {
    if (e.kind !== "cast" || !e.card_id) continue;
    const card = onStack.get(e.card_id);
    const seat = view.seats?.[e.seat]?.id;
    if (!card || !seat) continue;
    flights.push({
      itemID: e.card_id,
      seat,
      from: originOf(e),
      image: cardImageURL(card, "normal") ?? "/card-back.jpg",
      seq: e.seq,
    });
  }
  flights.sort((a, b) => a.seq - b.seq);
  return { tracker, flights: flights.slice(Math.max(0, flights.length - CAST_FLIGHT_MAX)) };
}

function escapeAttr(s: string): string {
  return s.replace(/["\\]/g, "\\$&");
}

/** Where to look for a flight's start, best first. */
export function castOriginSelectors(f: Pick<CastFlight, "seat" | "from">): string[] {
  const seat = seatSelector(f.seat);
  switch (f.from) {
    case "hand":
      return [`[data-hand-owner="${escapeAttr(f.seat)}"]`, seat];
    case "graveyard":
    case "exile":
    case "library":
      return [pileSelector(f.from, f.seat), seat];
    default:
      return [seat];
  }
}

/** The selector for a stack item's drawn element (the pile, a compact row, a lane tile). */
export function stackItemSelector(itemID: string): string {
  return `[data-stack-item-id="${escapeAttr(itemID)}"]`;
}
