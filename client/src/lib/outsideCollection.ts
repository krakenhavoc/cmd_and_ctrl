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
 * copies of a card, so a pick of several cards (Research's "up to four",
 * Spawnsire of Ulamog's "any number", #3955) takes a name as often as it is
 * clicked. A pick "with different names" says so (`distinct_names`, #4947:
 * Turtles Forever's four, Extrapolate the Impossible's pair) and keeps the
 * toggle whatever its ceiling, as a ceiling of two or less does.
 */
export function allowsRepeatedNames(max: number, distinctNames = false): boolean {
  return !distinctNames && max > 2;
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
 * Click a row: where names repeat (allowsRepeatedNames) the click adds
 * another copy while there is room, and a row's own "-" takes one back
 * (unpickOneCollection); otherwise a picked name is unpicked, with a
 * ceiling of one the click replaces the pick, and a name is added while
 * there is room.
 */
export function toggleCollectionPick(
  picked: string[],
  name: string,
  max: number,
  distinctNames = false,
): string[] {
  if (allowsRepeatedNames(max, distinctNames)) {
    return picked.length >= max ? picked : [...picked, name];
  }
  if (picked.includes(name)) return picked.filter((n) => n !== name);
  if (max <= 1) return [name];
  if (picked.length >= max) return picked;
  return [...picked, name];
}

/**
 * #4962: on a sideboard pick (a "choose_cards" prompt) that carries
 * `distinct_names`, whether the card `id` shares a name with a different
 * card already picked, so it can't be picked too. Names compare as the
 * server compares them (game.CardNamesDiffer): case-insensitively, with
 * surrounding space trimmed, and a card with no name matches nothing. A
 * picked card is never blocked, so it can always be taken back.
 */
export function nameAlreadyPicked(
  options: CardView[],
  selected: ReadonlySet<string>,
  id: string,
): boolean {
  if (selected.has(id)) return false;
  const key = (c: CardView | undefined): string => (c?.name ?? "").trim().toLowerCase();
  const name = key(options.find((c) => c.instance_id === id));
  if (name === "") return false;
  return options.some(
    (c) => c.instance_id !== id && selected.has(c.instance_id) && key(c) === name,
  );
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
