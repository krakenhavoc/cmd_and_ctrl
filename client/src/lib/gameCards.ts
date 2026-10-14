// gameCards.ts — the pure half of the "Cards from this game" panel
// (ADR 0179 §7): the wire shape of GET /me/games/{id}/cards and the
// list a "Buy these cards" cart is built from. No DOM, no fetch.

import type { CardLine } from "./shopLinks";

/** One card an opponent cast, and how often. */
export interface GameCard {
  name: string;
  casts: number;
}

/** One opponent's line. It names no one (ADR 0165 §9). */
export interface GameCardsSeat {
  commanders: string[];
  cards: GameCard[];
}

/** The body of GET /me/games/{id}/cards. */
export interface GameCards {
  /** False while the game's record is not written yet. */
  ready: boolean;
  opponents: GameCardsSeat[];
}

/** "Opponent 1", "Opponent 2" ...: seats are told apart, never named. */
export function opponentLabel(index: number): string {
  return `Opponent ${index + 1}`;
}

/**
 * boughtList is every distinct commander and card the opponents played,
 * one copy each (a player buys one of a card, however often it was
 * cast), commanders first, the rest in first-seen order.
 */
export function boughtList(g: GameCards): CardLine[] {
  const seen = new Set<string>();
  const out: CardLine[] = [];
  const add = (name: string): void => {
    const key = name.toLowerCase();
    if (!name || seen.has(key)) return;
    seen.add(key);
    out.push({ name, count: 1 });
  };
  for (const o of g.opponents) o.commanders.forEach(add);
  for (const o of g.opponents) o.cards.forEach((c) => add(c.name));
  return out;
}
