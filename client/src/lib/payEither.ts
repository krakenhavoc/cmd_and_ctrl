// payEither.ts — ADR 0187 §5, §8 (#3575): an either payment. One
// "pay_unless" offer with two or more ways to pay (Worms of the Earth:
// "sacrifice two lands or take 5 damage"). The prompt ships `pay_either`,
// one entry per branch, and the answer names the branch it takes as
// `pay_option` beside `apply: true`, with `card_ids` for a branch that
// names cards:
//
//   { apply: true, pay_option: 1 }                          take the damage
//   { apply: true, pay_option: 0, card_ids: [a, b] }        sacrifice two lands
//
// The card picker is the one a plain card payment uses (payCards.ts);
// this module only adapts a branch to it.

import type { PayCardsView, PayOptionView, PendingChoiceView } from "./protocol";

// eitherBranches is the prompt's branches, or none when it is not an
// either payment.
export function eitherBranches(c: Pick<PendingChoiceView, "pay_either"> | null | undefined) {
  return c?.pay_either ?? [];
}

// branchNamesCards reports whether a branch is answered by naming cards
// (a sacrifice, a discard, a graveyard exile) rather than by yes alone
// (damage, life, energy).
export function branchNamesCards(b: PayOptionView): boolean {
  return b.action === "sacrifice" || b.action === "discard" || b.action === "exile_graveyard";
}

// branchPicker is a card-naming branch in the shape the shared card
// picker reads (payCardsOptions, togglePayCard, payCardsReady,
// payCardsVerb).
export function branchPicker(b: PayOptionView): PayCardsView {
  return {
    action: b.action as PayCardsView["action"],
    count: b.count ?? 0,
    options: b.options ?? [],
  };
}

// defaultPickerBranch is the card branch whose picker is open before the
// player chooses one: the first they can pay, or -1 when none names
// cards or none can be paid.
export function defaultPickerBranch(branches: PayOptionView[]): number {
  return branches.findIndex((b) => branchNamesCards(b) && b.payable);
}

// branchLabel is a branch's button: "Sacrifice two lands", "Take 5
// damage".
export function branchLabel(b: PayOptionView): string {
  const w = b.words || b.action;
  return w[0].toUpperCase() + w.slice(1);
}

// branchReady reports whether pressing the branch's button now answers:
// it is payable, and a branch that names cards has exactly its count
// picked.
export function branchReady(b: PayOptionView, picks: string[]): boolean {
  if (!b.payable) return false;
  return branchNamesCards(b) ? picks.length === (b.count ?? 0) : true;
}

// branchWhyNot is why a branch can't be pressed, for its tooltip, or
// undefined when it can.
export function branchWhyNot(b: PayOptionView): string | undefined {
  if (b.payable) return undefined;
  if (b.action === "life") return "You can't pay that much life.";
  if (b.action === "energy") return "You don't have that much energy.";
  return `You don't have enough to ${b.words || "pay this"}.`;
}
