// Roadmap.hint.ts — the first-use hint for the roadmap's search (ADR 0125 §3.7).

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "roadmap.search",
  version: 1,
  place: "roadmap",
  order: 0,
  anchor: { label: L.searchRoadmap },
  title: "What comes next",
  body: "Search a mechanic, keyword or card name to see what it is waiting on before the game plays it for you.",
};

export default hint;
