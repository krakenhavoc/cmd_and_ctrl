// The lobby bot picker's tier choices (#2686). `random` is the engine's
// fuzzer, not an opponent, so it is never the default and is marked as
// testing-only wherever it is listed.

export interface TierChoice {
  tier: string;
  label: string;
  description: string;
  available: boolean;
}

export const RANDOM_TIER = "random";
export const DEFAULT_BOT_TIER = "heuristic";

// defaultBotTier prefers the heuristic tier when it is available, else the
// first available tier, else "" (nothing to offer).
export function defaultBotTier(tiers: readonly TierChoice[]): string {
  const available = tiers.filter((t) => t.available);
  return (available.find((t) => t.tier === DEFAULT_BOT_TIER) ?? available[0])?.tier ?? "";
}

// tierOptionLabel appends the testing mark to random's server label.
export function tierOptionLabel(t: TierChoice): string {
  const base = t.tier === RANDOM_TIER ? `${t.label} (testing only)` : t.label;
  return t.available ? base : `${base} — not built yet`;
}

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
