// elimination — a player leaves the game (#4449, ADR 0209).
//
// When a player is eliminated their avatar shatters like rock (a burst
// of shards, sparks and a ring on the avatar, with a heavy sound) and a
// tombstone settles on their panel with their name and their commander,
// where it stays for the rest of the game. This file holds the rules,
// pure and unit-tested:
//
//   - eliminatedSeats: which seats a frame's fresh log entries say left;
//   - the burst's pieces and timings (elimination layer draws them);
//   - the commander the tombstone names. An eliminated player's objects
//     leave the game with them (CR 800.4a) except when their leaving ends
//     it, so the frame that eliminates a seat no longer shows its
//     commander. The server names it instead, on the seat's final board
//     (#4547, `final_board.commanders`), the same for every viewer on every
//     frame, so a spectator or a fresh device that never saw the commander
//     names it too.
//
// EliminationFx.svelte measures, draws and plays; Tombstone.svelte is the
// stone on the panel.

import type { FinalBoardView, GameView, LogEvent, PlayerView } from "./protocol";

// ---- Which seats left ----

/**
 * The seats the log entries say left the game, each once, in log order.
 * One `eliminated` entry is written per departure (ADR 0057), a
 * concession included: it is a player leaving the game all the same.
 */
export function eliminatedSeats(entries: readonly LogEvent[]): number[] {
  const seats: number[] = [];
  for (const e of entries) {
    if (e.kind !== "eliminated") continue;
    if (typeof e.seat !== "number" || e.seat < 0) continue;
    if (!seats.includes(e.seat)) seats.push(e.seat);
  }
  return seats;
}

// ---- Timings, in ms before animations.speed ----

/**
 * Between two players leaving in one frame (a table-wide effect), so
 * each is heard: play() drops a repeat of one name inside 40 ms.
 */
export const ELIMINATION_STAGGER_MS = 180;
/** How long the burst draws before its pieces are removed. */
export const BURST_MS = 900;
/** The tombstone settles this long after the burst starts. */
export const TOMBSTONE_DELAY_MS = 520;
/** The drop of the tombstone, and the fade of one that is still. */
export const TOMBSTONE_DROP_MS = 460;
export const TOMBSTONE_FADE_MS = 600;

// ---- The burst ----

/** One wedge of the avatar, and where it flies. Offsets are in avatar widths. */
export interface Shard {
  clip: string;
  dx: number;
  dy: number;
  rot: number;
}

// The avatar cut into eight wedges meeting at its centre.
const WEDGES = [
  "polygon(50% 50%, 0% 0%, 50% 0%)",
  "polygon(50% 50%, 50% 0%, 100% 0%)",
  "polygon(50% 50%, 100% 0%, 100% 50%)",
  "polygon(50% 50%, 100% 50%, 100% 100%)",
  "polygon(50% 50%, 100% 100%, 50% 100%)",
  "polygon(50% 50%, 50% 100%, 0% 100%)",
  "polygon(50% 50%, 0% 100%, 0% 50%)",
  "polygon(50% 50%, 0% 50%, 0% 0%)",
] as const;

/** The avatar's wedges, each flying out from the centre and a little down. */
export function burstShards(): Shard[] {
  return WEDGES.map((clip, i) => {
    const angle = (i / WEDGES.length) * Math.PI * 2;
    const dist = 1.4 + (i % 2) * 0.8;
    return {
      clip,
      dx: round(Math.cos(angle) * dist),
      dy: round(Math.sin(angle) * dist + 0.4),
      rot: Math.round(angle * 60),
    };
  });
}

/** One spark of the burst. Offsets are in avatar widths. */
export interface Spark {
  dx: number;
  dy: number;
  size: number;
  ember: boolean;
}

export const SPARK_COUNT = 28;

/**
 * The burst's sparks, flung every way and falling. Pseudo-random but
 * fixed by `seed` (the seat's ID), so the same player shatters the same
 * way and a test can say where a spark went.
 */
export function burstSparks(seed: string, count = SPARK_COUNT): Spark[] {
  const rand = mulberry32(hash(seed));
  const out: Spark[] = [];
  for (let i = 0; i < count; i++) {
    const angle = rand() * Math.PI * 2;
    const reach = 0.8 + rand() * 2.2;
    out.push({
      dx: round(Math.cos(angle) * reach),
      // Gravity: whatever goes up comes down a little.
      dy: round(Math.sin(angle) * reach + 0.5),
      size: Math.round((1.5 + rand() * 2.5) * 10) / 10,
      ember: i % 2 === 0,
    });
  }
  return out;
}

function round(n: number): number {
  return Math.round(n * 1000) / 1000;
}

function hash(s: string): number {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

function mulberry32(seed: number): () => number {
  let a = seed;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

// ---- The commander the tombstone names ----

interface CommandCard {
  name?: string;
  type_line?: string;
  is_commander?: boolean;
}

/**
 * The commanders among a command zone's cards: those flagged, else every
 * card but a dungeon or an emblem, which a command zone also holds and
 * which are not commanders.
 */
function commandZoneCommanders<C extends CommandCard>(cards: readonly C[]): C[] {
  const flagged = cards.filter((c) => c.is_commander === true);
  if (flagged.length > 0) return flagged;
  return cards.filter((c) => {
    const t = c.type_line ?? "";
    return !t.includes("Dungeon") && !t.includes("Emblem");
  });
}

function addName(names: string[], name: string | undefined): void {
  if (name && !names.includes(name)) names.push(name);
}

/**
 * The names of a seat's commanders as this frame shows them: those in its
 * command zone first, then any it owns in a shared zone or its graveyard.
 * Empty when none is visible (they left with the player).
 */
export function commanderNames(view: GameView, seat: PlayerView): string[] {
  const names: string[] = [];
  for (const c of commandZoneCommanders(seat.command?.cards ?? [])) addName(names, c.name);
  for (const z of [view.battlefield, view.stack, view.exile, seat.graveyard]) {
    for (const c of z?.cards ?? []) {
      // The owner's: a commander they had stolen is another player's.
      if (c.is_commander && c.owner === seat.id) addName(names, c.name);
    }
  }
  return names;
}

/**
 * The commander names a final board records (#4547): `commanders`, which
 * the server fills with every commander the player owned wherever it was.
 * A record written before that field names the commanders in its command
 * zone. Empty without a record.
 */
export function finalBoardCommanders(fb: FinalBoardView | null | undefined): string[] {
  const names: string[] = [];
  if (!fb) return names;
  if (fb.commanders && fb.commanders.length > 0) {
    for (const n of fb.commanders) addName(names, n);
    return names;
  }
  for (const c of commandZoneCommanders(fb.command ?? [])) addName(names, c.name);
  return names;
}

/**
 * The commander names to put on a seat's tombstone: what its final board
 * records, which every viewer gets on every frame (a spectator or a fresh
 * device that never saw the commander included); else what this frame
 * shows, for a seat restored from a snapshot older than the record (#4912)
 * whose departure ended the game and so left its board in place.
 */
export function tombstoneCommanders(view: GameView, seat: PlayerView): string[] {
  const recorded = finalBoardCommanders(seat.final_board);
  if (recorded.length > 0) return recorded;
  return commanderNames(view, seat);
}

/** The commanders as one line: "A", "A & B". */
export function commanderLine(names: readonly string[]): string {
  return names.join(" & ");
}
