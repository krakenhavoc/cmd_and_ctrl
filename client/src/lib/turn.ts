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

// stepLabel resolves a step ID (or unknown string from a future
// server) to a human label, falling back to the raw ID so an
// unrecognised value still renders something meaningful.
export function stepLabel(step: string | undefined | null): string {
  if (!step) return "";
  return STEP_LABELS[step as StepID] ?? step;
}

// grantsPriority reports whether the named step grants priority. The
// auto-pass and pass-to-next-stop helpers consult this so they don't
// queue pass_priority calls on Untap / Cleanup (the server would
// reject them with ErrNoPriority anyway).
export function grantsPriority(step: string | undefined | null): boolean {
  if (!step) return false;
  return !NO_PRIORITY_STEPS.has(step as StepID);
}
