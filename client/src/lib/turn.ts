// Single source of truth for MTG turn structure on the client. The
// server's game.Step enum is mirrored here so Game.svelte (turn-bar
// label rendering), Settings.svelte (per-step stops UI), and the
// auto-pass / pass-to-next-stop helpers all reach for the same step
// IDs and labels.
//
// Server-side authority: server/internal/game/turn.go. Keep these
// arrays in sync with the canonical `turnSequence` over there.

// STEP_IDS is the ordered server-side step enum. Untap and Cleanup
// belong to the no-priority steps after S13 — the priority indicator
// and the per-step stops UI both use this list verbatim, so the user-
// facing affordances appear in canonical turn order.
export const STEP_IDS = [
  "untap",
  "upkeep",
  "draw",
  "precombat_main",
  "begin_combat",
  "declare_attackers",
  "declare_blockers",
  "first_strike_damage",
  "combat_damage",
  "end_combat",
  "postcombat_main",
  "end",
  "cleanup",
] as const;

export type StepID = (typeof STEP_IDS)[number];

// STEP_LABELS maps each step ID to a short user-facing label. Used
// by the turn bar (the abbreviated forms keep the bar compact) and
// the per-step stops checkbox grid (paired with the longer
// descriptions in STEP_DESCRIPTIONS for tooltips).
export const STEP_LABELS: Record<StepID, string> = {
  untap: "Untap",
  upkeep: "Upkeep",
  draw: "Draw",
  precombat_main: "Main 1",
  begin_combat: "Begin Combat",
  declare_attackers: "Declare Attackers",
  declare_blockers: "Declare Blockers",
  first_strike_damage: "First Strike",
  combat_damage: "Combat Damage",
  end_combat: "End Combat",
  postcombat_main: "Main 2",
  end: "End",
  cleanup: "Cleanup",
};

// NO_PRIORITY_STEPS is the set of steps that do not grant priority
// per CR 502.4 / 514.3. The server lands the cursor on these with
// PriorityHolder = -1; the client uses this set to decide whether
// to render the per-step stops checkbox (it doesn't, since you
// can't stop on a step that doesn't grant you priority anyway) and
// whether to show the priority indicator (it doesn't).
export const NO_PRIORITY_STEPS: ReadonlySet<StepID> = new Set(["untap", "cleanup"]);

// The two combat damage steps share ONE stop. CR 510.4 splits combat
// damage into a first-strike step and a regular step, and the server
// only gives a turn the first of them when first or double strike is
// on the board (`first_strike_damage`, #717). A player who asked to
// stop on combat damage means both, and a grid row for a step most
// turns do not have would be a setting that usually does nothing — so
// `first_strike_damage` reads the `combat_damage` stop and gets no row
// of its own. Manual one-shot pins (priorityStops.ts) are per step and
// are unaffected: you can pin just the first-strike step.
export function stopKeyFor(step: StepID): StepID {
  return step === "first_strike_damage" ? "combat_damage" : step;
}

// hasOwnStop reports whether a step gets its own entry in the per-step
// stops map and its own row in the settings grid: it must grant
// priority, and it must not borrow another step's stop.
export function hasOwnStop(step: StepID): boolean {
  return !NO_PRIORITY_STEPS.has(step) && stopKeyFor(step) === step;
}

// stepLabel resolves a step ID (or unknown string from a future
// server) to a human label, falling back to the raw ID so an
// unrecognised value still renders something meaningful.
export function stepLabel(step: string | undefined | null): string {
  if (!step) return "";
  return STEP_LABELS[step as StepID] ?? step;
}

// grantsPriority lived here until S31 sub-PR 2. Its only caller was
// timing.ts's canPassPriority, which had no callers of its own and
// went with it; priorityStops.ts asks NO_PRIORITY_STEPS directly.
