// AutoAnswersSettings.hint.ts — the first-use hint for Settings →
// Gameplay → Automatic answers (ADR 0127 §6, ADR 0125 §3.7). Place
// "settings": it draws inside the dialog, through HintSlot, once the
// Gameplay tab shows the section.

import { L } from "../labels";
import type { Hint } from "../hints/hint";

const hint: Hint = {
  id: "settings.auto-answers",
  version: 1,
  place: "settings",
  order: 10,
  anchor: { label: L.automaticAnswers },
  title: "Answers the game gives for you",
  body: "Prompts you ticked Remember this answer on are listed here. Set one to Ask, or Forget it, to be asked again.",
  // The section is on the Gameplay tab only; the hint waits for it.
  waitsForAnchor: true,
};

export default hint;
