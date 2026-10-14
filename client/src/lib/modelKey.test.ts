import { describe, expect, it } from "vitest";
import {
  DEFAULT_LIMITS,
  MODEL_KEY_OFF,
  describeKey,
  formFor,
  keyRequired,
  normalizeModelKeyState,
  type ModelKeyState,
  type ModelKeyView,
} from "./modelKey";
import { ownKeyTierChoices, tierForOwnKey, type TierChoice } from "./botTier";

const view: ModelKeyView = {
  provider: "anthropic",
  provider_label: "Anthropic",
  model: "claude-haiku-4-5",
  last4: "cdef",
  games_per_day: 3,
  calls_per_game: 40,
  games_today: 1,
  created_at: "2026-10-14T00:00:00Z",
  updated_at: "2026-10-14T00:00:00Z",
};

const on = (key: ModelKeyView | null): ModelKeyState => ({
  available: true,
  key,
  providers: [
    { id: "anthropic", label: "Anthropic", transport: "anthropic" },
    { id: "custom", label: "Other (OpenAI-compatible)", transport: "openai" },
  ],
  limits: DEFAULT_LIMITS,
});

describe("normalizeModelKeyState", () => {
  it("is off unless the server says available", () => {
    expect(normalizeModelKeyState(undefined)).toEqual(MODEL_KEY_OFF);
    expect(normalizeModelKeyState({ available: false, key: view })).toEqual(MODEL_KEY_OFF);
  });
  it("keeps the view and fills missing limits", () => {
    const got = normalizeModelKeyState({ available: true, key: view, providers: [] });
    expect(got.key).toEqual(view);
    expect(got.limits).toEqual(DEFAULT_LIMITS);
  });
});

describe("the form", () => {
  it("starts from the saved key with an empty key field", () => {
    const f = formFor(on(view));
    expect(f).toMatchObject({
      provider: "anthropic",
      model: "claude-haiku-4-5",
      key: "",
      games_per_day: 3,
    });
  });
  it("starts from the defaults with no key saved", () => {
    expect(formFor(on(null))).toMatchObject({
      provider: "anthropic",
      games_per_day: DEFAULT_LIMITS.default_games_per_day,
      calls_per_game: DEFAULT_LIMITS.default_calls_per_game,
    });
  });
  it("needs a key typed only for a new key, a new provider or a new custom address", () => {
    expect(keyRequired(on(null), formFor(on(null)))).toBe(true);
    expect(keyRequired(on(view), formFor(on(view)))).toBe(false);
    expect(keyRequired(on(view), { ...formFor(on(view)), provider: "custom" })).toBe(true);
    const custom = { ...view, provider: "custom", base_url: "https://a.example/v1" };
    expect(keyRequired(on(custom), { ...formFor(on(custom)) })).toBe(false);
    expect(
      keyRequired(on(custom), { ...formFor(on(custom)), base_url: "https://b.example/v1" }),
    ).toBe(true);
  });
  it("describes a key by provider, model and last four only", () => {
    expect(describeKey(view)).toBe("Anthropic · claude-haiku-4-5 · key ending cdef");
  });
});

describe("own-key tiers", () => {
  const t = (tier: string, available: boolean): TierChoice => ({
    tier,
    label: tier,
    description: "",
    available,
  });
  it("offers the model tiers, available, even when the server has no model", () => {
    const got = ownKeyTierChoices([
      t("random", true),
      t("heuristic", true),
      t("assisted", false),
      t("strong", false),
    ]);
    expect(got.map((x) => [x.tier, x.available])).toEqual([
      ["assisted", true],
      ["strong", true],
    ]);
  });
  it("keeps a model tier and moves off any other", () => {
    expect(tierForOwnKey("strong")).toBe("strong");
    expect(tierForOwnKey("heuristic")).toBe("assisted");
  });
});
