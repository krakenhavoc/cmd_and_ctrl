// waterbend.ts — #1310 / #1311: the two waterbend payments that are not
// a spell's (CR 701.67a). An activated ability's "Waterbend {N}:" cost
// and a "Ward—Waterbend {N}" pay-or-counter prompt both ship the
// clause in the TapCostView shape a hand card's convoke / waterbend
// uses, so TapCostModal and tapCostLimit serve all three; what lives
// here is only the question each flow asks before opening it, and the
// payload each answer sends.

import type { ActivatedAbilityView, PendingChoiceView, TapCostView } from "./protocol";
import { tapCostLimit } from "./targeting";

// waterbendLimit is how many permanents the picker may take: the
// server's cap, or the announced X for a Waterbend {X}, and never more
// than there are permanents to tap.
export function waterbendLimit(tc: TapCostView, xValue: number | undefined): number {
  const cap = tapCostLimit(tc, xValue);
  const n = tc.options?.cards?.length ?? 0;
  return Math.max(0, Math.min(cap, n));
}

// shouldAskAbilityWaterbend reports whether an activation should stop
// to ask which permanents to tap. Skipped when there is nothing to tap
// or nothing the taps could pay (a Waterbend {X} announced at 0): the
// whole cost is then paid with mana, which tapping none always means.
export function shouldAskAbilityWaterbend(
  ability: ActivatedAbilityView,
  xValue: number | undefined,
): boolean {
  return ability.waterbend !== undefined && waterbendLimit(ability.waterbend, xValue) > 0;
}

// payUnlessAnswer is the resolve_choice payload for a pay_unless
// answer: `{choice_id, apply}`, plus `tap_ids` when a waterbend
// payment taps something. Taps on a "Don't pay" are never sent — the
// server refuses them rather than reading them as a decline.
export function payUnlessAnswer(
  choice: PendingChoiceView,
  apply: boolean,
  tapIDs: string[] = [],
): Record<string, unknown> {
  const out: Record<string, unknown> = { choice_id: choice.id, apply };
  if (apply && choice.tap_cost && tapIDs.length > 0) out.tap_ids = [...tapIDs];
  return out;
}
