// castPreview.ts — #696. The announce-time half of a cast, in the
// shape `/games/:id/auto-tap-preview` prices it.
//
// The preview endpoint answers "can I pay for this cast, and which
// permanents would tap", and before this it was only ever told the
// card. Everything the caster has already chosen by then changes the
// price: which zone the cast comes out of (the commander tax, the
// cost modifiers, the granted permission that opens it), which CR
// 118.9 alternative cost is being claimed, which optional additional
// costs are being paid (ADR 0073), which permanents are convoked, and
// which printed face is being cast (ADR 0034). A preview that skips
// them prices a different cast from the one the confirm button sends,
// and the modal then disables a cast the server would have allowed —
// or plans taps for one it will refuse.
//
// Two builders because the choices reach the two callers in two
// shapes, and both must produce the same query:
//
//   - castPreviewParams reads the in-flight CastChoices bundle, which
//     is what the X picker and the Phyrexian stepper hold while the
//     cast is still being assembled;
//   - castPreviewParamsFromPayload reads a dispatched cast_spell
//     payload back off Game.svelte's per-card stash, which is what
//     the auto-tap retry has after the server rejected the cast.

import type { CastChoices } from "./targeting";

// AutoTapCastParams is the preview's cast-shaped options, in the
// camelCase the api.ts fetchers take. api.ts turns them into the
// snake_case query the server reads.
export interface AutoTapCastParams {
  fromZone?: string;
  alternativeCost?: string;
  optionalCosts?: number[];
  tapIDs?: string[];
  face?: number;
  // #1242: the cards and permanents named to the additional cost.
  // They do not change the price; they change what the auto-tapper
  // may spend on it — the server will not crack the Eldrazi Spawn a
  // cast offers to Village Rites, so the preview must not either.
  sacrificeIDs?: string[];
  discardIDs?: string[];
}

// castPreviewParams projects the choices announced so far onto the
// preview's query. Undefined and empty are dropped rather than sent:
// the server reads a missing `from_zone` as the hand and a missing
// `face` as the front, exactly as the cast path does, so an omitted
// field and its default are the same request.
export function castPreviewParams(choices: CastChoices | null | undefined): AutoTapCastParams {
  const out: AutoTapCastParams = {};
  if (!choices) return out;
  if (choices.fromZone) out.fromZone = choices.fromZone;
  if (choices.altCost) out.alternativeCost = choices.altCost;
  if (choices.optionalCosts && choices.optionalCosts.length > 0) {
    out.optionalCosts = [...choices.optionalCosts];
  }
  if (choices.tapIDs && choices.tapIDs.length > 0) out.tapIDs = [...choices.tapIDs];
  if (choices.face) out.face = choices.face;
  if (choices.sacrificeIDs && choices.sacrificeIDs.length > 0) {
    out.sacrificeIDs = [...choices.sacrificeIDs];
  }
  if (choices.discardIDs && choices.discardIDs.length > 0) out.discardIDs = [...choices.discardIDs];
  return out;
}

function stringIDs(v: unknown): string[] {
  if (!Array.isArray(v)) return [];
  return v.filter((s): s is string => typeof s === "string" && s !== "");
}

// castPreviewParamsFromPayload reads the same fields back off a
// dispatched `cast_spell` payload — the wire object Game.svelte
// stashes per instance ID so a retry can replay it.
//
// Defensive about types because the stash is a
// Record<string, unknown>: a field of the wrong shape is dropped
// rather than sent, since a 400 from the preview would leave the
// player with no readout at all, and the cast itself is validated by
// the server either way.
export function castPreviewParamsFromPayload(
  payload: Record<string, unknown> | null | undefined,
): AutoTapCastParams {
  const out: AutoTapCastParams = {};
  if (!payload) return out;
  if (typeof payload.from_zone === "string" && payload.from_zone !== "") {
    out.fromZone = payload.from_zone;
  }
  if (typeof payload.alternative_cost === "string" && payload.alternative_cost !== "") {
    out.alternativeCost = payload.alternative_cost;
  }
  if (Array.isArray(payload.optional_costs)) {
    const costs = payload.optional_costs.filter(
      (n): n is number => typeof n === "number" && Number.isInteger(n) && n >= 0,
    );
    if (costs.length > 0) out.optionalCosts = costs;
  }
  if (Array.isArray(payload.tap_ids)) {
    const ids = payload.tap_ids.filter((s): s is string => typeof s === "string" && s !== "");
    if (ids.length > 0) out.tapIDs = ids;
  }
  if (typeof payload.face === "number" && Number.isInteger(payload.face) && payload.face > 0) {
    out.face = payload.face;
  }
  const sacs = stringIDs(payload.sacrifice_ids);
  if (sacs.length > 0) out.sacrificeIDs = sacs;
  const discards = stringIDs(payload.discard_ids);
  if (discards.length > 0) out.discardIDs = discards;
  return out;
}
