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

// ---------------------------------------------------------------------
// Repeated phases and turn skips (ADR 0188, #3567). Pure derivations
// over the wire fields the server already sends, so the board only
// renders them.

/** The first phase id an effect can add; the template uses 1..5 (turn_plan.go). */
const TEMPLATE_PHASE_COUNT = 5;

/**
 * The caption of a track group, given the nth phase of its family this
 * turn (TurnView.phase_ordinal, which only describes the phase in
 * progress). Precombat and postcombat main are one family, so a third
 * main reads "Main 3". An ordinal of 1, or none, leaves the printed
 * caption alone.
 */
export function phaseCaption(
  family: "main" | "combat" | "other",
  base: string,
  ordinal: number | undefined,
): string {
  if (family === "main" && ordinal !== undefined && ordinal >= 1) return `Main ${ordinal}`;
  if (family === "combat" && ordinal !== undefined && ordinal > 1) return `Combat ${ordinal}`;
  return base;
}

/**
 * One "+ Combat" / "+ Main" marker per phase an effect has added and
 * the turn has not reached yet, in turn order. The phase in progress
 * is not a marker: its own caption already says "Combat 2".
 */
export function addedPhaseMarkers(
  upcoming: ReadonlyArray<{ step: string; phase_id: number }> | undefined,
  currentPhaseId: number | undefined,
): string[] {
  const out: string[] = [];
  const seen = new Set<number>();
  for (const s of upcoming ?? []) {
    if (s.phase_id <= TEMPLATE_PHASE_COUNT || s.phase_id === currentPhaseId) continue;
    if (seen.has(s.phase_id)) continue;
    seen.add(s.phase_id);
    if (s.step === "postcombat_main" || s.step === "precombat_main") out.push("+ Main");
    else if (s.step === "begin_combat") out.push("+ Combat");
    else if (s.step === "upkeep" || s.step === "untap" || s.step === "draw") out.push("+ Begin");
    else if (s.step === "end" || s.step === "cleanup") out.push("+ End");
  }
  return out;
}

interface SeatLite {
  name?: string;
  seat: number;
  turn_skips?: number;
}

/**
 * The "Next:" line. A queued extra turn wins (it is taken before
 * rotation resumes). Otherwise the line appears only when next_seat
 * jumps over a seat that owes a skip: the plain case has no line, and a
 * seat that merely left is not announced as a skip.
 */
export function nextTurnLine(
  turn: { active_seat: number; next_seat?: number; extra_turns?: number[] },
  seats: readonly SeatLite[],
): string | null {
  const nameOf = (i: number): string => seats.find((p) => p.seat === i)?.name ?? `seat ${i}`;
  const extra = turn.extra_turns?.[0];
  if (extra !== undefined) return `Next: ${nameOf(extra)} (extra)`;
  const next = turn.next_seat;
  const n = seats.length;
  if (next === undefined || n === 0) return null;
  const skipped: string[] = [];
  let i = (turn.active_seat + 1) % n;
  for (let step = 0; step < n && i !== next; step++, i = (i + 1) % n) {
    const p = seats.find((s) => s.seat === i);
    if ((p?.turn_skips ?? 0) > 0) skipped.push(nameOf(i));
  }
  if (skipped.length === 0) return null;
  return `Next: ${nameOf(next)} (skips ${skipped.join(", ")})`;
}

/** The player panel's chip for a seat that will skip turns, or null. */
export function turnSkipChip(skips: number | undefined): string | null {
  if (!skips || skips < 1) return null;
  return skips === 1 ? "Skips next turn" : `Skips next ${skips} turns`;
}
