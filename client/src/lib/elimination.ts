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
//     commander. The name is remembered from the frames before it.
//
// EliminationFx.svelte measures, draws and plays; Tombstone.svelte is the
// stone on the panel.

import type { CardView, GameView, LogEvent, PlayerView } from "./protocol";

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

function isDungeonOrEmblem(c: CardView): boolean {
  const t = c.type_line ?? "";
  return t.includes("Dungeon") || t.includes("Emblem");
}

/**
 * The names of a seat's commanders as this frame shows them: those in its
 * command zone first, then any it owns or controls in a shared zone or its
 * graveyard. Empty when none is visible (they left with the player).
 */
export function commanderNames(view: GameView, seat: PlayerView): string[] {
  const names: string[] = [];
  const add = (c: CardView) => {
    if (c.name && !names.includes(c.name)) names.push(c.name);
  };
  const command = seat.command?.cards ?? [];
  const flagged = command.filter((c) => c.is_commander === true);
  // A command zone also holds dungeons and emblems; they are not commanders.
  for (const c of flagged.length > 0 ? flagged : command.filter((c) => !isDungeonOrEmblem(c))) {
    add(c);
  }
  for (const z of [view.battlefield, view.stack, view.exile, seat.graveyard]) {
    for (const c of z?.cards ?? []) {
      if (c.is_commander && (c.owner === seat.id || c.controller === seat.id)) add(c);
    }
  }
  return names;
}

const MEMORY_KEY = "cmdctrl.commanders";
const MEMORY_MAX_GAMES = 8;

// game ID -> seat ID -> names. Insertion-ordered, so the oldest game is
// the first key and is dropped past MEMORY_MAX_GAMES.
const memory = new Map<string, Map<string, string[]>>();
let loaded = false;

function load(): void {
  if (loaded) return;
  loaded = true;
  try {
    const raw = typeof sessionStorage === "undefined" ? null : sessionStorage.getItem(MEMORY_KEY);
    if (!raw) return;
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return;
    for (const entry of parsed) {
      if (!Array.isArray(entry) || typeof entry[0] !== "string" || !Array.isArray(entry[1]))
        continue;
      const seats = new Map<string, string[]>();
      for (const pair of entry[1]) {
        if (!Array.isArray(pair) || typeof pair[0] !== "string" || !Array.isArray(pair[1]))
          continue;
        seats.set(
          pair[0],
          pair[1].filter((n): n is string => typeof n === "string"),
        );
      }
      memory.set(entry[0], seats);
    }
  } catch {
    // Storage blocked or damaged: the memory lives for this page only.
  }
}

function save(): void {
  try {
    if (typeof sessionStorage === "undefined") return;
    const dump = [...memory].map(([game, seats]) => [game, [...seats]]);
    sessionStorage.setItem(MEMORY_KEY, JSON.stringify(dump));
  } catch {
    // Same: a private window still remembers in memory.
  }
}

/**
 * Remember the seat's commanders while they can be seen, so the tombstone
 * can name them after they have gone. Only ever adds: the frame that
 * eliminates a seat shows none, and that must not forget them. The
 * remembered names survive a reload of the page (sessionStorage).
 */
export function rememberCommanders(view: GameView, seat: PlayerView): void {
  const names = commanderNames(view, seat);
  if (names.length === 0) return;
  load();
  let seats = memory.get(view.id);
  if (!seats) {
    seats = new Map();
    memory.set(view.id, seats);
    if (memory.size > MEMORY_MAX_GAMES) {
      const oldest = memory.keys().next().value;
      if (oldest !== undefined) memory.delete(oldest);
    }
  }
  const before = seats.get(seat.id) ?? [];
  const merged = [...before, ...names.filter((n) => !before.includes(n))];
  if (merged.length === before.length) return;
  seats.set(seat.id, merged);
  save();
}

/** The commander names to put on a seat's tombstone: what is visible now, else what was remembered. */
export function tombstoneCommanders(view: GameView, seat: PlayerView): string[] {
  const now = commanderNames(view, seat);
  if (now.length > 0) return now;
  load();
  return memory.get(view.id)?.get(seat.id) ?? [];
}

/** Forget everything remembered. For tests. */
export function resetCommanderMemory(): void {
  memory.clear();
  loaded = false;
  try {
    if (typeof sessionStorage !== "undefined") sessionStorage.removeItem(MEMORY_KEY);
  } catch {
    // ignore
  }
}

/** The commanders as one line: "A", "A & B". */
export function commanderLine(names: readonly string[]): string {
  return names.join(" & ");
}
