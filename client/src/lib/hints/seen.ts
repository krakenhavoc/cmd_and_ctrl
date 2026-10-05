// seen.ts — `settings.help.seen`, the hints a person has dismissed
// (ADR 0125 §3.3, §4).
//
// The map is hint id → the version last dismissed. A hint is unseen when
// its id is missing or the stored version is lower than the hint's own,
// so bumping a hint's `version` offers it once more to everyone.
//
// Pure, and free of the hint collection, so settings.ts and
// settingsSync.ts can import it.

import { RETIRED_HINT_IDS } from "./retired";

export type SeenMap = Record<string, number>;

const RETIRED: ReadonlySet<string> = new Set(RETIRED_HINT_IDS);

/** isSeenVersion is a stored version the map may hold: a positive integer. */
function isSeenVersion(v: unknown): v is number {
  return typeof v === "number" && Number.isInteger(v) && v > 0;
}

/**
 * normalizeSeen is the stored map, checked: string ids with a positive
 * integer version, retired ids dropped, everything else kept, unknown
 * ids included (an older tab must not prune a newer client's hints).
 */
export function normalizeSeen(raw: unknown, retired: ReadonlySet<string> = RETIRED): SeenMap {
  const out: SeenMap = {};
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return out;
  for (const [id, v] of Object.entries(raw as Record<string, unknown>)) {
    if (id === "" || retired.has(id) || !isSeenVersion(v)) continue;
    out[id] = v;
  }
  return out;
}

/** isUnseen reports whether a hint at `version` should still be offered. */
export function isUnseen(seen: SeenMap, id: string, version: number): boolean {
  const v = seen[id];
  return v === undefined || v < version;
}

/**
 * unionSeen merges two maps: every id either has, with the higher
 * version where both do. It cannot lose anything either side had, which
 * is why `help.seen` merges this way at sign-in and on a 412 (§4).
 */
export function unionSeen(a: SeenMap, b: SeenMap): SeenMap {
  const out: SeenMap = { ...a };
  for (const [id, v] of Object.entries(b)) {
    const have = out[id];
    out[id] = have === undefined ? v : Math.max(have, v);
  }
  return out;
}

/** withSeen marks `id` dismissed at `version`, never lowering a stored version. */
export function withSeen(seen: SeenMap, id: string, version: number): SeenMap {
  const have = seen[id];
  if (have !== undefined && have >= version) return seen;
  return { ...seen, [id]: version };
}

/** withoutSeen forgets the given ids, so their hints are offered again. */
export function withoutSeen(seen: SeenMap, ids: Iterable<string>): SeenMap {
  const drop = new Set(ids);
  const out: SeenMap = {};
  for (const [id, v] of Object.entries(seen)) if (!drop.has(id)) out[id] = v;
  return out;
}
