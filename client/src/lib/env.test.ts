import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { get } from "svelte/store";
import {
  parseConfig,
  loadAppConfig,
  resetAppConfigForTests,
  appConfig,
  isDev,
  devFeature,
  PROD_CONFIG,
} from "./env";

describe("parseConfig", () => {
  it("treats anything that isn't env:'dev' as production", () => {
    expect(parseConfig({ env: "prod", features: { card_spawn: true } })).toEqual(PROD_CONFIG);
    expect(parseConfig({ env: "staging", features: { card_spawn: true } })).toEqual(PROD_CONFIG);
    expect(parseConfig({ features: { card_spawn: true } })).toEqual(PROD_CONFIG);
    expect(parseConfig(null)).toEqual(PROD_CONFIG);
    expect(parseConfig("dev")).toEqual(PROD_CONFIG);
  });

  it("reads dev features, defaulting unknown/missing ones to false", () => {
    const got = parseConfig({ env: "dev", features: { card_spawn: true, seat_swap: "yes" } });
    expect(got.env).toBe("dev");
    expect(got.features.card_spawn).toBe(true);
    // Non-boolean truthy values do not enable a feature.
    expect(got.features.seat_swap).toBe(false);
    expect(got.features.frame_inspector).toBe(false);
  });

  it("survives a dev response with no features object", () => {
    expect(parseConfig({ env: "dev" }).features.card_spawn).toBe(false);
  });
});

describe("loadAppConfig", () => {
  beforeEach(() => resetAppConfigForTests());
  afterEach(() => vi.unstubAllGlobals());

  it("defaults to production before any fetch resolves", () => {
    expect(get(appConfig)).toEqual(PROD_CONFIG);
    expect(get(isDev)).toBe(false);
  });

  it("populates the store from a dev response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ env: "dev", features: { frame_inspector: true } }),
      }),
    );
    await loadAppConfig();
    expect(get(isDev)).toBe(true);
    expect(get(devFeature("frame_inspector"))).toBe(true);
    expect(get(devFeature("card_spawn"))).toBe(false);
  });

  // The rollout-order case: prod can serve this client before the
  // reverse proxy routes /config. A 404 must not surface dev UI.
  it("falls back to production on a non-ok response", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 404 }));
    await loadAppConfig();
    expect(get(appConfig)).toEqual(PROD_CONFIG);
  });

  it("falls back to production when fetch throws", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));
    await loadAppConfig();
    expect(get(appConfig)).toEqual(PROD_CONFIG);
  });

  it("fetches once even when called concurrently", async () => {
    const spy = vi.fn().mockResolvedValue({ ok: true, json: async () => ({ env: "dev" }) });
    vi.stubGlobal("fetch", spy);
    await Promise.all([loadAppConfig(), loadAppConfig(), loadAppConfig()]);
    expect(spy).toHaveBeenCalledTimes(1);
  });
});
