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
//
// ADR 0129's amendment of 2026-10-09 (#1941) widened pay_amount to any
// number a resolving effect asks for: life ("pay any amount of life",
// Necrodominance, Phyrexian Processor), or a number that is chosen and
// not paid ("an amount of damage of your choice", Volcano Hellion), with
// or without a ceiling. The helpers below read `resource` for that.

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

// payAmountResource is what each point of a pay_amount costs: energy
// when the server sends none (every prompt from before #1941), life, or
// none for a number that is chosen and not paid.
export function payAmountResource(pa: PayAmountView): "energy" | "life" | "none" {
  return pa.resource ?? "energy";
}

// payAmountPays reports whether the prompt is a payment, which may
// always be declined with 0 (CR 118.12), rather than a number.
export function payAmountPays(pa: PayAmountView): boolean {
  return payAmountResource(pa) !== "none";
}

// payAmountFloor is the stepper's floor. A payment's is the smallest
// payment: "one or more" is 1, and "any amount" still starts the stepper
// at 1, since 0 is the decline button. A number that is not paid starts
// at its own minimum, 0 included: there is no decline.
export function payAmountFloor(pa: PayAmountView): number {
  if (!payAmountPays(pa)) return pa.min;
  return Math.min(Math.max(pa.min, 1), pa.max);
}

// payAmountStart is where the stepper opens: the card's own threshold
// when it names one, else the floor.
export function payAmountStart(pa: PayAmountView): number {
  const goal = pa.goal ?? 0;
  if (goal >= payAmountFloor(pa) && goal <= pa.max) return goal;
  return payAmountFloor(pa);
}

// payAmountAnswerable reports whether `n` is an answer the server takes:
// a whole number within min..max, or 0 for a payment.
export function payAmountAnswerable(pa: PayAmountView, n: number): boolean {
  if (!Number.isInteger(n)) return false;
  return (n === 0 && payAmountPays(pa)) || (n >= pa.min && n <= pa.max);
}

// payAmountClamp keeps a stepped or typed value inside the stepper.
export function payAmountClamp(pa: PayAmountView, n: number): number {
  if (!Number.isFinite(n)) return payAmountFloor(pa);
  return Math.min(Math.max(Math.round(n), payAmountFloor(pa)), pa.max);
}

// payAmountCanSubmit is whether the primary answers `n`: an answerable
// number, and for a payment one above 0 (0 is the decline button).
export function payAmountCanSubmit(pa: PayAmountView, n: number): boolean {
  return payAmountAnswerable(pa, n) && (n > 0 || !payAmountPays(pa));
}

const UNIT_WORDS: Record<PayAmountView["unit"], string> = {
  damage: "1 damage",
  counters: "1 counter",
  cards: "1 card",
  power: "+1 to the X",
  tax: "{1} more to pay",
  other: "1 more",
};

// payAmountHint is the prompt's hint: the bounds, what is paid, the
// card's threshold and what one point buys.
export function payAmountHint(pa: PayAmountView, sourceName: string): string {
  const goal = pa.goal ?? 0;
  const amount = pa.min > 0 ? "one or more" : "any amount of";
  const each = UNIT_WORDS[pa.unit] ?? "1 more";
  let parts: string[];
  switch (payAmountResource(pa)) {
    case "life":
      parts = [
        `Pay ${amount} life, up to ${pa.max}, or nothing.`,
        `Each life is ${each} for ${sourceName}.`,
      ];
      break;
    case "none":
      parts = [
        pa.no_max
          ? `Choose any number from ${pa.min} up.`
          : `Choose a number from ${pa.min} to ${pa.max}.`,
        `Each point is ${each} for ${sourceName}${pa.self_damage ? ", and you are dealt the same amount" : ""}.`,
      ];
      break;
    default:
      parts = [
        `Pay ${amount} energy, up to ${pa.max}, or nothing.`,
        `Each energy is ${each} for ${sourceName}.`,
      ];
  }
  if (goal > 0) parts.push(`The stepper starts at ${goal}, the amount the card is asking for.`);
  return parts.join(" ");
}
