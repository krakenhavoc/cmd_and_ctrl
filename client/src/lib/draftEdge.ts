// draftEdge.ts — the slight coloured edge on a draft card (#5704).
//
// A card's edge says its colour identity at a glance: one colour is that
// colour, two or more are split into equal diagonal bands in WUBRG
// order (so the edge still shows which colours), and a colourless card
// gets the neutral. A land has no colours of its own, so its colour
// identity (the printing's, on the wire as `color_identity`) is used;
// every other card falls back to `colors` when the wire sent no
// identity.
//
// The inks are the colour-identity tokens (DESIGN.md §2), which no skin
// overrides. The edge is a redundant signal: the mana symbols, the
// card's own art and the pool's colour groups say the same thing, and
// the element that wears it is hidden from assistive technology.

import type { DraftCard } from "./draft";

export const EDGE_COLOURS = ["W", "U", "B", "R", "G"] as const;
export type EdgeColour = (typeof EDGE_COLOURS)[number];

const INK: Record<EdgeColour, string> = {
  W: "var(--identity-w)",
  U: "var(--identity-u)",
  B: "var(--identity-b)",
  R: "var(--identity-r)",
  G: "var(--identity-g)",
};
/** Colourless. */
const NEUTRAL = "var(--identity-c)";

/** The colours of a card's edge, in WUBRG order; empty for a colourless card. */
export function edgeColours(card: Pick<DraftCard, "colors" | "color_identity">): EdgeColour[] {
  const source =
    card.color_identity && card.color_identity.length > 0 ? card.color_identity : card.colors;
  const have = new Set(source ?? []);
  return EDGE_COLOURS.filter((c) => have.has(c));
}

/**
 * The CSS background of an edge, a gradient made of the identity
 * tokens. One colour is a flat fill; several are equal hard-edged
 * bands on a diagonal, first colour at the top left.
 */
export function edgeBackground(colours: readonly EdgeColour[]): string {
  if (colours.length === 0) return `linear-gradient(${NEUTRAL}, ${NEUTRAL})`;
  if (colours.length === 1) return `linear-gradient(${INK[colours[0]]}, ${INK[colours[0]]})`;
  const step = 100 / colours.length;
  const stops = colours.map((c, i) => {
    const from = Math.round(i * step * 100) / 100;
    const to = Math.round((i + 1) * step * 100) / 100;
    return `${INK[c]} ${from}% ${to}%`;
  });
  return `linear-gradient(135deg, ${stops.join(", ")})`;
}

export interface CardEdge {
  /** "W", "UR", "" for colourless: for a test or a data attribute. */
  key: string;
  /** The CSS `background` value. */
  background: string;
}

/** The edge of a card. */
export function cardEdge(card: Pick<DraftCard, "colors" | "color_identity">): CardEdge {
  const colours = edgeColours(card);
  return { key: colours.join(""), background: edgeBackground(colours) };
}
