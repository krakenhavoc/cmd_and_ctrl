// payLifeForMana.ts — the "Pay life for {B}…" row (ADR 0131 §4, #2531).
//
// Under K'rrik, Son of Yawgmoth every {B} in the viewer's costs can be
// paid with 2 life. A cast asks about that only when mana falls short
// (phyrexianLife.ts shouldAskPhyrexianLife), so a card whose mana the
// viewer HAS would never show the stepper. The card's menu keeps this
// row for the player who wants to pay life anyway: choosing it starts
// the ordinary cast chain with the Phyrexian stepper forced open.
//
// The row lives on a hand card's popover, two components deep under the
// Board, and the stepper needs the Board's cast chain, so the request
// goes through this store rather than a callback threaded through every
// panel: the same shape as castAnyway.ts. Board watches it, clears it, and
// starts the chain.

import { type Writable } from "svelte/store";

import { guardedWritable } from "./guardedStore";
import type { CardView } from "./protocol";
import { L } from "./labels";

// The row's text and accessible name: a label contract (AGENTS.md §5).
export const PAY_LIFE_LABEL = L.payLifeForMana;
// The row's title.
export const PAY_LIFE_TITLE = "Choose how many {B} to pay with 2 life each instead of mana.";

// payLifeRequest is the card whose cast should open with the stepper,
// or null.
export const payLifeRequest: Writable<CardView | null> = guardedWritable(null, "payLifeRequest");

// requestPayLife asks the Board to start `card`'s cast with the life
// stepper open. It sends nothing itself.
export function requestPayLife(card: CardView): void {
  payLifeRequest.set(card);
}

// grantedLifeOffered says whether a card gets the row: it has a symbol
// payable with life only because of a grant.
export function grantedLifeOffered(card: CardView): boolean {
  return (card.phyrexian_granted ?? 0) > 0;
}
