// deckGrid.ts — the Decks page's grid of saved decks (ADR 0158 §2): the
// sort, the coverage ring's numbers and tone, the commander art, the
// colour-identity aura and the notes line. Pure functions, so the rules
// have unit tests; DeckCard.svelte only draws what these return.

import { coverageDetail, type MyDeckCommander, type MyDeckInfo } from "./myDecks";

export type DeckSort = "recent" | "name" | "coverage";

export const DECK_SORTS: readonly { key: DeckSort; label: string }[] = [
  { key: "recent", label: "Recent" },
  { key: "name", label: "Name" },
  { key: "coverage", label: "Coverage" },
];

/** The ring turns green at this share of the deck playing as printed. */
export const RING_GOOD_AT = 0.96;

export interface DeckRing {
  /** Copies that play exactly as printed. */
  n: number;
  /** The deck's size, unresolved copies included. */
  total: number;
  /** n / total, 0..1. */
  share: number;
  tone: "good" | "warn";
}

/** deckRing is the coverage ring's numbers, or null with no coverage. */
export function deckRing(d: MyDeckInfo): DeckRing | null {
  const c = d.coverage;
  if (!c || c.deck_size <= 0) return null;
  const share = Math.min(1, Math.max(0, c.as_printed_copies / c.deck_size));
  return {
    n: c.as_printed_copies,
    total: c.deck_size,
    share,
    tone: share >= RING_GOOD_AT ? "good" : "warn",
  };
}

const byName = (a: MyDeckInfo, b: MyDeckInfo): number =>
  a.name.localeCompare(b.name, undefined, { sensitivity: "base" });

/**
 * sortDecks returns a sorted copy. Recent keeps the list's own order:
 * the server's, newest updated first, with a just-saved deck put on top
 * by the page. Name is
 * A to Z; Coverage is the best-covered first, with a deck the server
 * could not grade last. Ties fall back to the name.
 */
export function sortDecks(decks: readonly MyDeckInfo[], sort: DeckSort): MyDeckInfo[] {
  const out = [...decks];
  switch (sort) {
    case "name":
      return out.sort(byName);
    case "coverage":
      return out.sort((a, b) => {
        const ra = deckRing(a)?.share ?? -1;
        const rb = deckRing(b)?.share ?? -1;
        return rb - ra || byName(a, b);
      });
    default:
      return out;
  }
}

export interface DeckArt {
  name: string;
  scryfallId: string;
}

/**
 * deckArt lists the commanders with a printing to draw, partners side by
 * side. A saved deck, a pre-built deck and a bot deck all carry
 * `commander_cards` the same way (GET /me/decks, GET /decks and GET
 * /bot/options), so each is drawn by the one rule.
 */
export function deckArt(d: { commander_cards?: MyDeckCommander[] }): DeckArt[] {
  return (d.commander_cards ?? [])
    .filter((c) => !!c.scryfall_id)
    .map((c) => ({ name: c.name, scryfallId: c.scryfall_id! }));
}

const WUBRG = ["W", "U", "B", "R", "G"] as const;

/**
 * deckIdentity is the deck's colour identity for the aura: WUBRG
 * letters, [] for a colourless deck, or null when no commander resolved
 * (the aura then stays neutral rather than guessing).
 */
export function deckIdentity(d: MyDeckInfo): string[] | null {
  const cmds = d.commander_cards ?? [];
  if (!cmds.some((c) => Array.isArray(c.color_identity))) return null;
  const have = new Set(d.color_identity ?? []);
  return WUBRG.filter((l) => have.has(l));
}

/**
 * auraColors names the fixed identity tokens the aura runs through:
 * one per colour, steel for colourless, none when unknown. The tokens
 * are app.css's --identity-*, which no skin overrides.
 */
export function auraColors(identity: string[] | null): string[] {
  if (identity === null) return [];
  if (identity.length === 0) return ["var(--identity-c)"];
  return identity.map((l) => `var(--identity-${l.toLowerCase()})`);
}

/**
 * auraGradient is the gradient the border, the glow and a plain band
 * use: a single colour is a flat run of itself, more make a sweep
 * through them. Empty for an unknown identity.
 */
export function auraGradient(identity: string[] | null): string {
  const colors = auraColors(identity);
  if (colors.length === 0) return "";
  const stops = colors.length === 1 ? [colors[0], colors[0]] : colors;
  return `linear-gradient(120deg, ${stops.join(", ")})`;
}

/** deckNotes is the notes line before the source: what is simplified or by hand. */
export function deckNotes(d: MyDeckInfo): string {
  if (!d.coverage) return "";
  return coverageDetail(d) || "every card plays as printed";
}
