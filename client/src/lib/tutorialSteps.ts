// tutorialSteps.ts — the tutorial's script (ADR 0076 §2.1).
//
// The walking skeleton (#1079, sub-PR 3) wires the two button steps only:
// step 1 opens, step 11 hands off. Sub-PR 4 (#1081) puts steps 2–10
// between them; each is a TutorialStep (tutorial.ts) with its anchor, its
// predicate and its copy. The copy is the CMD CTRL Tutorial canvas's. It
// assumes the player knows Magic: it explains the client, never a rule.
//
// Anchors for the middle steps, checked against the DOM after ADR 0111:
//   your hand                      { label: "your hand" }
//   lands / creatures              { label: "lands", within: "your board" }
//                                  (every opponent's panel has the same lists)
//   step 8, the action dock        { label: "actions" }
//   step 9, the attention strip    { label: "attention" }
//   step 7, one card               { cardID }

import type { TutorialStep } from "./tutorial";

export const WELCOME: TutorialStep = {
  id: "welcome",
  n: 1,
  kind: "opening",
  title: "A five-minute practice game",
  body: "You are seated against a practice bot. Mana is not enforced and you can undo, so nothing here can go wrong.",
};

export const HANDOFF: TutorialStep = {
  id: "handoff",
  n: 11,
  kind: "done",
  title: "That is the whole interface",
  body: ({ helpKey, settingsKey }) => {
    const help = helpKey ? `Press ${helpKey} for the keymap` : "The keymap is in Settings";
    const set = settingsKey ? ` and ${settingsKey} for settings` : "";
    return `${help}${set}. Your zone piles are on the rail: library draws, graveyard and exile open a browser.`;
  },
};

/** The script, in order. */
export const TUTORIAL_STEPS: TutorialStep[] = [WELCOME, HANDOFF];
