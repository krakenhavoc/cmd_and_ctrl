// payEnergy.ts — ADR 0129 §3: energy paid while a spell or ability
// resolves. Pure helpers for the two prompts that ask for it:
//
//   - a pay_unless with `pay_energy` ("you may pay {E}{E}", "sacrifice it
//     unless you pay {E}"): a yes/no whose Pay is greyed when the seat is
//     short, because the server reads a short "Pay" as Don't pay
//     (CR 118.3);
//   - a pay_amount ("you may pay any amount of {E}"): a number from the
//     prompt's floor to the seat's energy, or nothing, drawn as a stepper
//     in the action dock (owner decision 3).

import type { PayAmountView, PendingChoiceView } from "./protocol";

// energyShortBy is how much energy the seat lacks for a pay_unless's
// energy payment: 0 when it can pay, or when the prompt is not one.
export function energyShortBy(
  c: Pick<PendingChoiceView, "pay_energy"> | null | undefined,
  energy: number,
): number {
  const need = c?.pay_energy;
  if (need === undefined || need === null) return 0;
  return Math.max(0, need - energy);
}

// energyShortReason is the greyed Pay's reason: the server's refusal
// text for an activation the seat cannot pay (ADR 0129 §8).
export function energyShortReason(have: number, need: number): string {
  return `Not enough energy (have ${have}, need ${need})`;
}

// payAmountFloor is the smallest payment: "one or more" is 1, and "any
// amount" still starts the stepper at 1, since 0 is the decline button.
export function payAmountFloor(pa: PayAmountView): number {
  return Math.min(Math.max(pa.min, 1), pa.max);
}

// payAmountStart is where the stepper opens: the card's own threshold
// when it names one, else the smallest payment.
export function payAmountStart(pa: PayAmountView): number {
  const goal = pa.goal ?? 0;
  if (goal >= payAmountFloor(pa) && goal <= pa.max) return goal;
  return payAmountFloor(pa);
}

// payAmountAnswerable reports whether `n` is an answer the server takes:
// a whole number that is 0 or within min..max.
export function payAmountAnswerable(pa: PayAmountView, n: number): boolean {
  if (!Number.isInteger(n)) return false;
  return n === 0 || (n >= pa.min && n <= pa.max);
}

// payAmountClamp keeps a stepped or typed value inside the stepper.
export function payAmountClamp(pa: PayAmountView, n: number): number {
  if (!Number.isFinite(n)) return payAmountFloor(pa);
  return Math.min(Math.max(Math.round(n), payAmountFloor(pa)), pa.max);
}

const UNIT_WORDS: Record<PayAmountView["unit"], string> = {
  damage: "1 damage",
  counters: "1 counter",
  cards: "1 card",
  power: "+1 to the X",
  tax: "{1} more to pay",
  other: "1 more",
};

// payAmountHint is the prompt's hint: the bounds, the seat's energy, the
// card's threshold and what one energy buys.
export function payAmountHint(pa: PayAmountView, sourceName: string): string {
  const goal = pa.goal ?? 0;
  const parts = [
    `Pay ${pa.min > 0 ? "one or more" : "any amount of"} energy, up to ${pa.max}, or nothing.`,
    `Each energy is ${UNIT_WORDS[pa.unit] ?? "1 more"} for ${sourceName}.`,
  ];
  if (goal > 0) parts.push(`The stepper starts at ${goal}, the amount the card is asking for.`);
  return parts.join(" ");
}
