// Catalog.hint.ts — the first-use hint for the catalogue's search (ADR 0125 §3.7).
//
// #6120 paged the grid. The body says so, since a search is now the quick
// way to a card rather than scrolling for it. Still version 1: nothing the
// hint said before has become wrong (§3.3), so a returning player is not
// shown it again.

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "catalog.search",
  version: 1,
  place: "catalog",
  order: 0,
  anchor: { label: L.searchCatalogue },
  title: "Every card the game plays",
  body: "Search by name, type or rules text to see how much of a card the game handles for you. Results come a page at a time.",
};

export default hint;
