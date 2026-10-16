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
 * Click a row: a picked name is unpicked; with a ceiling of one the click
 * replaces the pick; otherwise a name is added while there is room.
 */
export function toggleCollectionPick(picked: string[], name: string, max: number): string[] {
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
