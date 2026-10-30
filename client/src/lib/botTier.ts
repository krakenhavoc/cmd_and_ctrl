// The lobby bot picker's tier choices (#2686). `random` is the engine's
// fuzzer, not an opponent, so it is never the default and is marked as
// testing-only wherever it is listed.

export interface TierChoice {
  tier: string;
  label: string;
  description: string;
  available: boolean;
  /** Why an unavailable tier is unavailable, from the server. */
  reason?: string;
}

export const RANDOM_TIER = "random";
export const DEFAULT_BOT_TIER = "heuristic";

// defaultBotTier prefers the heuristic tier when it is available, else the
// first available tier, else "" (nothing to offer).
export function defaultBotTier(tiers: readonly TierChoice[]): string {
  const available = tiers.filter((t) => t.available);
  return (available.find((t) => t.tier === DEFAULT_BOT_TIER) ?? available[0])?.tier ?? "";
}

// tierLabel is a tier's segment in the bot picker (ADR 0212 owner
// decision 8): the server's label, with the testing mark on random. An
// unavailable tier is a disabled segment, and unavailableNotes says why.
export function tierLabel(t: TierChoice): string {
  return t.tier === RANDOM_TIER ? `${t.label} (testing only)` : t.label;
}

// unavailableNotes is the line under the tier control for the tiers it
// cannot offer, one per reason the server gave: "Assisted and Strong:
// this server has no model API key". A tier with no reason is "not
// available on this server".
export function unavailableNotes(tiers: readonly TierChoice[]): string[] {
  const byReason = new Map<string, string[]>();
  for (const t of tiers) {
    if (t.available) continue;
    const why = t.reason?.trim() || "not available on this server";
    byReason.set(why, [...(byReason.get(why) ?? []), tierLabel(t)]);
  }
  return [...byReason].map(([why, labels]) => `${andList(labels)}: ${why}`);
}

const andList = (names: readonly string[]): string =>
  names.length <= 1
    ? (names[0] ?? "")
    : `${names.slice(0, -1).join(", ")} and ${names[names.length - 1]}`;

export const RANDOM_TIER_WARNING =
  "Random picks any legal move. It's the engine's test fuzzer, not an opponent.";

// The tiers that call a model, and so the ones a player's own model key
// can run (ADR 0182). Mirrors aiseat.Tier.NeedsModel.
export const MODEL_TIERS: readonly string[] = ["assisted", "strong"];

// ownKeyTierChoices is the picker's tier list while "run on my model
// key" is ticked: the model tiers only, each available, whether or not
// the server has a model of its own.
export function ownKeyTierChoices<T extends TierChoice>(tiers: readonly T[]): T[] {
  return tiers.filter((t) => MODEL_TIERS.includes(t.tier)).map((t) => ({ ...t, available: true }));
}

// tierForOwnKey keeps the selected tier when it can run on a key, and
// otherwise picks the first model tier.
export function tierForOwnKey(selected: string): string {
  return MODEL_TIERS.includes(selected) ? selected : MODEL_TIERS[0];
}

// tierNote is the caption under the picker for the selected tier: the
// server's description, flagged as a warning for random.
export function tierNote(
  tiers: readonly TierChoice[],
  selected: string,
): { text: string; warning: boolean } | null {
  const t = tiers.find((x) => x.tier === selected);
  if (!t) return null;
  if (t.tier === RANDOM_TIER) {
    return { text: `${RANDOM_TIER_WARNING} ${t.description}`.trim(), warning: true };
  }
  return t.description ? { text: t.description, warning: false } : null;
}
