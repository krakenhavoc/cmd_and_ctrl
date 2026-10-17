// dungeon.ts — what the player panel says about a dungeon (ADR 0203,
// #4096, CR 309). The engine sends the whole card on PlayerView.dungeon:
// its rooms in printed order and the room the venture marker is in. This
// file turns that into the chip's name, its screen-reader description and
// the room list the card draws, so the component holds no rules.

import type { DungeonRoomView, DungeonView } from "./protocol";

/** One room as the card lists it. */
export interface DungeonRoomLine {
  key: string;
  name: string;
  text: string;
  /** The venture marker is in this room. */
  current: boolean;
  /** An arrow from the current room points here: a room the next venture may enter. */
  next: boolean;
  /** The names of the rooms its arrows point to, for "leads to". */
  leadsTo: string[];
}

/** The room the marker is in, or undefined. */
export function currentRoom(d: DungeonView): DungeonRoomView | undefined {
  return d.rooms.find((r) => r.key === d.room);
}

/** The rooms in printed order, marked current and next. */
export function dungeonRoomLines(d: DungeonView): DungeonRoomLine[] {
  const byKey = new Map(d.rooms.map((r) => [r.key, r]));
  const next = new Set(currentRoom(d)?.leads_to ?? []);
  return d.rooms.map((r) => ({
    key: r.key,
    name: r.name,
    text: r.text,
    current: r.key === d.room,
    next: next.has(r.key),
    leadsTo: (r.leads_to ?? []).map((k) => byKey.get(k)?.name ?? k),
  }));
}

/** The chip's accessible name: "Undercity, in Forge". */
export function dungeonChipName(d: DungeonView): string {
  const room = currentRoom(d);
  return room ? `${d.name}, in ${room.name}` : d.name;
}

/**
 * "Completed Undercity twice" / "Completed Undercity and Tomb of
 * Annihilation". Empty for none.
 */
export function completedPhrase(completed: readonly string[]): string {
  if (completed.length === 0) return "";
  const counts = new Map<string, number>();
  for (const name of completed) counts.set(name, (counts.get(name) ?? 0) + 1);
  const parts = [...counts].map(([name, n]) =>
    n === 1 ? name : n === 2 ? `${name} twice` : `${name} ${n} times`,
  );
  const list =
    parts.length === 1
      ? parts[0]
      : `${parts.slice(0, -1).join(", ")} and ${parts[parts.length - 1]}`;
  return `Completed ${list}`;
}

/** The chip's description: where the marker is, what each room does, where it can go next. */
export function dungeonDescription(
  d: DungeonView | undefined,
  completed: readonly string[],
): string {
  const parts: string[] = [];
  if (d) {
    const room = currentRoom(d);
    if (room) {
      parts.push(`In ${room.name}: ${room.text}`);
      const next = dungeonRoomLines(d).filter((l) => l.next);
      parts.push(
        next.length === 0
          ? `The last room: the next venture completes ${d.name}.`
          : `Next: ${next.map((l) => `${l.name} (${l.text})`).join(" or ")}.`,
      );
    }
  }
  const done = completedPhrase(completed);
  if (done) parts.push(`${done}.`);
  return parts.join(" ");
}
