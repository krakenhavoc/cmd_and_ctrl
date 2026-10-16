// outsideCollection.ts — the picker for an "outside_collection" prompt
// (ADR 0198): "a card you own from outside the game" for a seat whose deck
// has no sideboard, chosen from the whole card database.
//
// The prompt carries a few suggested cards (collection_options); the
// player can also search the database (searchOutsideCards), and the
// server filters the search by the effect, so every row is a card the
// pick accepts. The answer is the picked cards' NAMES.

import type { DevCardResult } from "./api";
import type { CardView } from "./protocol";

/** One row of the picker: a suggestion or a search result. */
export interface CollectionRow {
  name: string;
  typeLine: string;
  manaCost: string;
}

/** A search runs once the query has this many characters. */
export const COLLECTION_MIN_QUERY = 2;

/**
 * The rows the picker shows: the search results while the player is
 * searching, the prompt's suggestions otherwise.
 */
export function collectionRows(
  suggestions: CardView[],
  results: DevCardResult[],
  query: string,
): CollectionRow[] {
  if (query.trim().length >= COLLECTION_MIN_QUERY) {
    return results.map((r) => ({ name: r.name, typeLine: r.type_line, manaCost: r.mana_cost }));
  }
  return suggestions.map((c) => ({
    name: c.name ?? "",
    typeLine: c.type_line ?? "",
    manaCost: c.mana_cost ?? "",
  }));
}

/**
 * Whether the same name may be picked more than once. A player may own two
 * copies of a card, so an "any number" pick (Spawnsire of Ulamog's cast,
 * #3955, whose ceiling is 100) takes a name as often as it is clicked. No
 * prompt with a set rule asks for more than two cards ("exactly two with
 * different names"), so a ceiling above two is the "any number" shape and
 * a ceiling of two or less keeps the toggle.
 */
export function allowsRepeatedNames(max: number): boolean {
  return max > 2;
}

/** How many times `name` is picked. */
export function pickedCount(picked: string[], name: string): number {
  return picked.filter((n) => n === name).length;
}

/** Take one copy of `name` back: the last one picked. */
export function unpickOneCollection(picked: string[], name: string): string[] {
  const i = picked.lastIndexOf(name);
  return i < 0 ? picked : [...picked.slice(0, i), ...picked.slice(i + 1)];
}

/**
 * Click a row: where names repeat (an "any number" pick) the click adds
 * another copy while there is room, and a row's own "-" takes one back
 * (unpickOneCollection); otherwise a picked name is unpicked, with a
 * ceiling of one the click replaces the pick, and a name is added while
 * there is room.
 */
export function toggleCollectionPick(picked: string[], name: string, max: number): string[] {
  if (allowsRepeatedNames(max)) return picked.length >= max ? picked : [...picked, name];
  if (picked.includes(name)) return picked.filter((n) => n !== name);
  if (max <= 1) return [name];
  if (picked.length >= max) return picked;
  return [...picked, name];
}

/**
 * Whether the picks can be sent as a "take": at least one card, and
 * within the prompt's bounds. Declining is its own button.
 */
export function canTakeCollection(picked: string[], min: number, max: number): boolean {
  return picked.length >= Math.max(min, 1) && picked.length <= Math.max(max, 1);
}

/** The take button's label: "Take Sol Ring", "Take Sol Ring and Ponder". */
export function takeCollectionLabel(picked: string[]): string {
  return picked.length === 0 ? "Take…" : `Take ${picked.join(" and ")}`;
}
