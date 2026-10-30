import type { GameMeta } from "./api";
import { L } from "./labels";
import { TABLE_SEATS, allDecksIn, seatName, seatReady } from "./tablesPage";

// readyMoment.ts is the pure half of the ready-check moment (ADR 0212 §2):
// the instant a lobby table becomes startable by decks, because the last
// seated player's deck landed. Lobby.svelte draws the answers: a brief
// ring on the seat whose deck landed, the checks and the counter landing
// in seat order, one soft ring on the Start game primary, and a polite
// live region that says "Every deck is in. You can start the game." once.
//
// WHEN IT FIRES. The page keeps no socket: it re-reads GET /games every
// 5 s and on focus, and the sit-down answer patches the seats at once.
// observeReady is called with the tables the page shows after every one
// of those writes, and fires for a table only on the EDGE between two
// reads this client made of the same table: the earlier read was not
// allDecksIn (two or more seated, every one ready) and this one is.
//
//   - A table's first read only records it. A page that opens on an
//     already-ready table, or a table that comes into the list ready
//     (another login's seat, the admin's list), says nothing.
//   - A read that is allDecksIn after one that was says nothing. That is
//     every 5 s poll of a ready table, and a seat that leaves or joins
//     already ready while the table stays startable: nothing changed
//     about whether it can start.
//   - A table that leaves allDecksIn (someone sits with no deck yet, or a
//     seat leaves and fewer than two remain) and comes back fires again.
//     Each is a real moment of the table becoming startable.
//   - A table that drops out of the list is forgotten, so coming back
//     counts as a first read. A moment can be missed that way, never
//     invented.
//   - A seat leaving whose absence makes the rest all ready (the one
//     still choosing walks away from three) fires with no seat to ring:
//     the table did just become startable.
//
// No DOM and no fetch, so it is unit-tested (readyMoment.test.ts).

// What the page remembers of one table from its last read.
export interface TableLook {
  // allDecksIn at that read.
  allIn: boolean;
  // The players who held a deck at that read, by player_id: seats are
  // renumbered when one leaves, so a seat number is not an identity.
  ready: ReadonlySet<string>;
}

// The page's memory: the last read of every table it showed.
export type ReadyMemory = ReadonlyMap<string, TableLook>;

export const NO_MEMORY: ReadyMemory = new Map();

export interface ReadyMoment {
  gameID: string;
  // The seat whose deck landed last: the highest-numbered seat that is
  // ready now and was not at the last read (a bot added to a ready table
  // counts: it arrives with its deck). null when no seat did.
  landed: number | null;
}

// observeReady folds one read of the page's tables into its memory and
// answers the moments to show. `tables` is everything the page lists, in
// any state; only a table in the lobby can be allDecksIn.
export function observeReady(
  memory: ReadyMemory,
  tables: readonly GameMeta[],
): { memory: ReadyMemory; moments: ReadyMoment[] } {
  const next = new Map<string, TableLook>();
  const moments: ReadyMoment[] = [];
  for (const g of tables) {
    const look: TableLook = {
      allIn: allDecksIn(g),
      ready: new Set(g.players.filter(seatReady).map((p) => p.player_id)),
    };
    next.set(g.id, look);
    const before = memory.get(g.id);
    if (!before || before.allIn || !look.allIn) continue;
    const arrived = g.players
      .filter((p) => seatReady(p) && !before.ready.has(p.player_id))
      .map((p) => p.seat);
    moments.push({ gameID: g.id, landed: arrived.length > 0 ? Math.max(...arrived) : null });
  }
  return { memory: next, moments };
}

// --- motion ------------------------------------------------------------

// "full" draws the moment; "still" draws none of it and the checks, the
// counter and the button simply change (DESIGN.md §4).
export type ReadyMotion = "full" | "still";

// readyMotion folds the signals. Any one of them makes the moment still:
// the animations master switch off, the setting `accessibility.reduceMotion`
// (which sets data-reduce-motion), or the OS's `prefers-reduced-motion`.
// The caller passes every signal, since accessibility.reduceMotion is
// never pushed into animations.ts. The CSS carries the media query again
// for the frames before the settings load.
export function readyMotion(s: {
  enabled: boolean;
  reduceMotion: boolean;
  osReduceMotion?: boolean;
}): ReadyMotion {
  return s.enabled && !s.reduceMotion && s.osReduceMotion !== true ? "full" : "still";
}

// How long the moment's classes stay on the card: the checks' wave, the
// ring on the tile that landed and the ring on the primary, and a little
// over, so the last keyframe has played. It is UI chrome, not a game
// animation, so animations.speed does not scale it.
export const READY_MOMENT_MS = 2400;

// How long the live region keeps its line before it is emptied, so a
// stale "Every deck is in" is not left for a screen reader's browse mode
// after the table has started.
export const READY_ANNOUNCE_HOLD_MS = 10_000;

// The wave of checks and counter segments: one seat after the next.
export const READY_STAGGER_MS = 110;

// The delay of the n-th seat's part of the wave, and of the ring on the
// primary, which waits for the last seat's.
export const readyDelayMs = (seat: number): number =>
  Math.min(Math.max(seat, 0), TABLE_SEATS - 1) * READY_STAGGER_MS;
export const READY_PRIMARY_DELAY_MS = TABLE_SEATS * READY_STAGGER_MS;

// --- the live region -----------------------------------------------------

// readyAnnouncement is the line the polite live region says once (§2):
// "Every deck is in. You can start the game." for whoever may start, else
// "Every deck is in. Waiting for <host> to start." `named` adds the
// table's name, for a person who takes part in more than one lobby
// table, so the line says which one.
export function readyAnnouncement(g: GameMeta, o: { startable: boolean; named: boolean }): string {
  const host = g.players.find((p) => p.is_host);
  return L.everyDeckIn(o.startable, host ? seatName(host) : "the host", o.named ? g.name : "");
}
