// Deployment identity and dev-only feature flags, fetched once from
// GET /config at app start.
//
// This module decides what the UI *draws*. It is deliberately not a
// security boundary: every dev-only route and action is independently
// refused by the server when it isn't a dev deployment (see
// server/internal/lobby/devconfig.go requireDev). Flipping a flag in
// devtools reveals a button that 404s.
//
// Failure posture is fail-closed: a missing endpoint, a 500, a
// network error, or a malformed body all resolve to production with
// no features. That matters for rollout order — production can serve
// the new client before its reverse proxy learns to route /config,
// and the only consequence is that the (correct) prod defaults apply.

import { type Readable } from "svelte/store";

import { guardedDerived, guardedWritable } from "./guardedStore";

export type EnvName = "prod" | "dev";

// DevFeatures mirrors appenv.Features in the Go server. Keys are the
// JSON tags, not the Go field names.
export interface DevFeatures {
  card_spawn: boolean;
  seat_swap: boolean;
  frame_inspector: boolean;
  replay_scrubber: boolean;
}

export interface AppConfig {
  env: EnvName;
  features: DevFeatures;
}

export const NO_FEATURES: DevFeatures = {
  card_spawn: false,
  seat_swap: false,
  frame_inspector: false,
  replay_scrubber: false,
};

// PROD_CONFIG is both the initial value and the value every failure
// path resolves to.
export const PROD_CONFIG: AppConfig = { env: "prod", features: NO_FEATURES };

const store = guardedWritable<AppConfig>(PROD_CONFIG, "appConfig");

// appConfig is the live deployment config. Read-only to consumers —
// only loadAppConfig writes it.
export const appConfig: Readable<AppConfig> = { subscribe: store.subscribe };

// isDev is true only on a dev deployment. Drives the env banner.
// The `false` fallback is the same promise PROD_CONFIG makes: every
// failure path here resolves to production, which is the answer that
// hides dev-only surfaces rather than exposing them.
export const isDev: Readable<boolean> = guardedDerived(
  store,
  ($c) => $c.env === "dev",
  "isDev",
  false,
);

// devFeature returns a store for one flag. Components should prefer
// this over reading appConfig directly so the call site names the
// feature it depends on:
//
//   const showInspector = devFeature("frame_inspector");
//   {#if $showInspector} <FrameInspector /> {/if}
export function devFeature(name: keyof DevFeatures): Readable<boolean> {
  return guardedDerived(store, ($c) => $c.features[name] === true, `devFeature:${name}`, false);
}

// parseConfig narrows an untrusted JSON body onto AppConfig. Anything
// unexpected degrades to the production default rather than throwing,
// so a server that grows a new field (or an older one that lacks a
// feature this client knows about) never breaks the boot path.
export function parseConfig(body: unknown): AppConfig {
  if (typeof body !== "object" || body === null) return PROD_CONFIG;
  const b = body as { env?: unknown; features?: unknown };
  // Only the exact string "dev" opts in. Unknown env names are prod.
  if (b.env !== "dev") return PROD_CONFIG;
  const raw = (typeof b.features === "object" && b.features !== null ? b.features : {}) as Record<
    string,
    unknown
  >;
  return {
    env: "dev",
    features: {
      card_spawn: raw.card_spawn === true,
      seat_swap: raw.seat_swap === true,
      frame_inspector: raw.frame_inspector === true,
      replay_scrubber: raw.replay_scrubber === true,
    },
  };
}

let inflight: Promise<AppConfig> | null = null;

// loadAppConfig fetches /config once per page load. Concurrent calls
// share the same request; a settled result is not re-fetched, because
// the deployment cannot change identity under a running tab.
export async function loadAppConfig(): Promise<AppConfig> {
  if (inflight) return inflight;
  inflight = (async () => {
    try {
      const res = await fetch("/config", { method: "GET", credentials: "same-origin" });
      if (!res.ok) return PROD_CONFIG;
      return parseConfig(await res.json());
    } catch {
      return PROD_CONFIG;
    }
  })();
  const cfg = await inflight;
  store.set(cfg);
  return cfg;
}

// resetAppConfigForTests clears the memoised fetch and the store.
// Exported for unit tests only.
export function resetAppConfigForTests(): void {
  inflight = null;
  store.set(PROD_CONFIG);
}
