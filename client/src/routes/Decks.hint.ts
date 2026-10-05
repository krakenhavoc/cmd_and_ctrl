// Decks.hint.ts — the first-use hint for checking a deck (ADR 0125 §3.7).

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "decks.check",
  version: 1,
  place: "decks",
  order: 0,
  anchor: { label: L.deckLink },
  title: "Check a deck",
  body: "Paste a link or a list. The report shows which cards the game plays for you and which you play by hand.",
};

export default hint;
