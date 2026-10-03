// libraryCost.ts — ADR 0109 §7 (#1902) and owner decision 3: the pieces
// of the two cost components the activator chooses nothing for, apart
// from Svelte.
//
//   - "Exile the top N cards of your library" (Seasoned Tactician,
//     Arc-Slogger): `library_exile_cost_n`.
//   - "Discard N cards at random" (Pyromancy, Meteor Storm):
//     `discard_cost_random` beside `discard_cost_n`.
//
// Neither sends anything, so there is no picker. They are the costs
// most likely to make a player back out — ten cards off the library, a
// card the player did not choose — so the board confirms them before the
// activation is sent, and says so plainly when the hand or the library
// can't pay (CR 118.3), which the server would refuse.
//
// "Put a card from your hand on top of your library" has a pick and
// reuses the discard picker; topCostNote is its small print.

import type { ActivatedAbilityView, CardView, PlayerView } from "./protocol";

type CostFields = Pick<
  ActivatedAbilityView,
  "discard_cost_n" | "discard_cost_label" | "discard_cost_random" | "library_exile_cost_n"
>;

export interface CostConfirmLine {
  // What the payment does, as the card prints it.
  text: string;
  // Why it can't be paid right now; absent when it can.
  short?: string;
}

const NUMBER_WORDS = [
  "",
  "one",
  "two",
  "three",
  "four",
  "five",
  "six",
  "seven",
  "eight",
  "nine",
  "ten",
];

function cards(n: number): string {
  if (n === 1) return "a card";
  return `${NUMBER_WORDS[n] ?? String(n)} cards`;
}

// randomDiscardCount is how many cards the ability discards at random:
// zero for every ability that does not.
export function randomDiscardCount(a: CostFields): number {
  return a.discard_cost_random ? (a.discard_cost_n ?? 0) : 0;
}

// needsCostConfirm says whether the activation has a cost the board
// confirms rather than picks.
export function needsCostConfirm(a: CostFields): boolean {
  return randomDiscardCount(a) > 0 || (a.library_exile_cost_n ?? 0) > 0;
}

// randomDiscardPool is how many cards in `seat`'s hand a random discard
// could take: everything but the source and the cards the same payment
// already names (CR 601.2h pays the random discard last, from what the
// rest of the cost leaves).
export function randomDiscardPool(
  seat: Pick<PlayerView, "hand"> | undefined,
  sourceID: string,
  taken: string[],
): number {
  const skip = new Set([sourceID, ...taken]);
  return (seat?.hand?.cards ?? []).filter((c: CardView) => !skip.has(c.instance_id)).length;
}

// costConfirmLines is the confirm's body: one line per component, with
// the shortfall when the hand or the library can't pay.
export function costConfirmLines(
  a: CostFields,
  handAvailable: number,
  libraryCount: number,
): CostConfirmLine[] {
  const out: CostConfirmLine[] = [];
  const random = randomDiscardCount(a);
  if (random > 0) {
    const line: CostConfirmLine = {
      text: `Discard ${a.discard_cost_label || `${cards(random)} at random`}.`,
    };
    if (handAvailable < random) {
      line.short = `You have ${cards(handAvailable).replace(/^a card$/, "one card")} in hand to discard.`;
      if (handAvailable === 0) line.short = "You have no cards in hand to discard.";
    }
    out.push(line);
  }
  const library = a.library_exile_cost_n ?? 0;
  if (library > 0) {
    const line: CostConfirmLine = {
      text:
        library === 1
          ? "Exile the top card of your library."
          : `Exile the top ${NUMBER_WORDS[library] ?? String(library)} cards of your library.`,
    };
    if (libraryCount < library) {
      line.short =
        libraryCount === 0
          ? "Your library is empty."
          : `Your library has only ${cards(libraryCount).replace(/^a card$/, "one card")}.`;
    }
    out.push(line);
  }
  return out;
}

// The small print the confirm shows: the random discard is not the
// player's choice, and the library cards are paid after everything else.
export function costConfirmNote(a: CostFields): string {
  if (randomDiscardCount(a) > 0) return "chosen at random · CR 701.9b";
  return "activation cost · CR 602.2b";
}

// topCostNote is the hand-to-library picker's small print.
export const topCostNote = "put on top of your library · not a discard";
