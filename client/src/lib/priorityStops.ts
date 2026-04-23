// Manual one-time priority stops.
//
// Distinct from the persistent `settings.gameplay.stepStops` grid:
// those express the player's *default* intent ("always hold the
// cursor at my declare-attackers"). Manual stops are one-shot:
// click a phase icon to pin a stop at that step for the next time
// priority lands on the viewer there, then the pin auto-clears.
//
// Semantics in order of precedence (strongest first):
//   1. Manual stop set on this step → cursor holds. Overrides
//      autoPassPriority, stepStops, and smartAutoPass. This is the
//      "fake a game action" case: the viewer wants to think /
//      respond / bluff even though the engine sees nothing to do.
//   2. stepStops[step] === true  → cursor holds, subject to
//      smartAutoPass (S13.6) which skips when no legal response.
//   3. stepStops[step] === false → cursor auto-passes.
//
// Consumption: the manual pin clears when the snapshot step
// advances away from the pinned step. One-time behaviour means the
// next cycle of that step will not be held unless re-pinned. Only
// priority-granting steps can be pinned — Untap / Cleanup don't
// grant priority so pinning them would be a footgun.
//
// Scope: module-level Svelte store. Intentionally not persisted —
// manual stops are a now-intent, not a preference. Cleared on page
// reload, cleared on consume, cleared on toggle-off.

import { writable, get, type Readable } from "svelte/store";
import type { StepID } from "./turn";
import { NO_PRIORITY_STEPS } from "./turn";

const manual = writable<Set<StepID>>(new Set());

// manualStops is the read-only view used by reactive consumers
// (PhaseDisplay highlights, Game.svelte auto-pass override).
export const manualStops: Readable<Set<StepID>> = { subscribe: manual.subscribe };

// canManuallyStop reports whether the step is a legal target for a
// manual stop. Untap and Cleanup don't grant priority per CR 502.4
// / 514.3 so pinning them would just be confusing — the cursor
// moves past them via the step-entry hook, not via pass_priority.
export function canManuallyStop(step: StepID): boolean {
  return !NO_PRIORITY_STEPS.has(step);
}

// toggleManualStop adds / removes the step from the manual-stop
// set. No-ops on steps that aren't priority-granting. Returns the
// post-toggle membership so callers can e.g. flash a "saved" hint.
export function toggleManualStop(step: StepID): boolean {
  if (!canManuallyStop(step)) return false;
  let result = false;
  manual.update((set) => {
    const next = new Set(set);
    if (next.has(step)) {
      next.delete(step);
      result = false;
    } else {
      next.add(step);
      result = true;
    }
    return next;
  });
  return result;
}

// hasManualStop is the non-reactive read used by the auto-pass
// effect. Reactive consumers should subscribe to `manualStops`.
// Accepts `string | null | undefined` so callers (e.g. Game.svelte
// reading `view.turn.step`) don't have to cast; unknown step IDs
// simply never match.
export function hasManualStop(step: string | undefined | null): boolean {
  if (!step) return false;
  return get(manual).has(step as StepID);
}

// consumeManualStop drops a step from the set, called when the
// cursor has passed through the pinned step (the pin fired once).
// Safe to call for a step that isn't pinned or for an unknown step
// ID (defensive against stale snapshots).
export function consumeManualStop(step: string | undefined | null): void {
  if (!step) return;
  manual.update((set) => {
    if (!set.has(step as StepID)) return set;
    const next = new Set(set);
    next.delete(step as StepID);
    return next;
  });
}

// clearAllManualStops empties the set. Used on game end / session
// reset. Not wired to the session store currently — manual stops
// already clear on reload; exposed for future reuse.
export function clearAllManualStops(): void {
  manual.set(new Set());
}

// _resetForTests is the vitest teardown hook. Not for prod use.
export function _resetForTests(): void {
  manual.set(new Set());
}
