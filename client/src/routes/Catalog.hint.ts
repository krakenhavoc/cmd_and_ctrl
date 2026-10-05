// Catalog.hint.ts — the first-use hint for the catalogue's search (ADR 0125 §3.7).

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "catalog.search",
  version: 1,
  place: "catalog",
  order: 0,
  anchor: { label: L.searchCatalogue },
  title: "Every card the game plays",
  body: "Search by name, type or rules text to see how much of a card the game handles for you.",
};

export default hint;
