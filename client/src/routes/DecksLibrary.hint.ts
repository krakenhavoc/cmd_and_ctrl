// DecksLibrary.hint.ts — the first-use hint for the saved decks (ADR 0125 §3.7).
// The list only exists once the person has a saved deck, so the hint
// waits for one: its anchor does not resolve before that.

import { L } from "../lib/labels";
import type { Hint } from "../lib/hints/hint";

const hint: Hint = {
  id: "decks.library",
  version: 1,
  place: "decks",
  order: 1,
  anchor: { label: L.yourDecks },
  when: (c) => c.signedIn,
  title: "Your decks",
  body: "Save a checked deck here and it is offered when you sit down at a table.",
};

export default hint;
