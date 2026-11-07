package game

import "github.com/google/uuid"

// phase_effects.go — a rules effect bound to ONE future combat phase of
// this turn (#5277, ADR 0188 amendment 2026-11-07, Decision 13).
//
// Bumi, Unleashed: "After this phase, there is an additional combat
// phase. Only land creatures can attack during that combat phase." The
// restriction is created during the first combat, but it applies only in
// the combat Bumi added, even when a newer added combat runs first
// (CR 500.8: the most recently created phase occurs first; the Bumi
// ruling of 2025-10-02).
//
// It is ADR 0188 Decision 3b's shape (an extra turn's deferred effects)
// for a phase: the record is built and validated when the ability
// resolves, waits on Game.PhaseEffects keyed by the turn (Turn.Seq) and
// the phase instance (PlannedStep.PhaseID), and is registered on the
// ScopedEffects registry as that combat phase begins, lasting until the
// end of that combat (UntilEndOfCombat, CR 511.3). No duration is
// dormant until a future phase, and nothing goes on the stack that the
// card does not print (a delayed trigger, Decision 5's binding, would).
//
// A record whose combat never begins (the turn ends early, CR 724.1, or
// every step of the phase is skipped) is dropped as the next turn
// begins, beside the delayed triggers bound to a phase of that turn.

// PhaseEffect is one deferred rules effect waiting for a phase instance.
type PhaseEffect struct {
	// Turn is the Turn.Seq the phase belongs to.
	Turn int `json:"turn"`
	// PhaseID is the phase instance (PlannedStep.PhaseID).
	PhaseID int `json:"phaseId"`
	// Effect is the record, built and validated when the card resolved
	// and stamped with the creation timestamp (CR 613.7b). Its Duration
	// is a placeholder: registration stamps UntilEndOfCombat.
	Effect ScopedEffect `json:"effect"`
}

// DuringCombatPhaseRuleEffectForEffect attaches a rules effect to combat
// phase `phaseID` of this turn: "Only land creatures can attack during
// that combat phase" (Bumi, Unleashed). The effect is created now and
// starts as that combat begins.
//
// The phase must be a combat phase still to come in this turn's plan,
// as AddPhasesForEffect returned it. The combat in progress is refused:
// "during that combat phase" names a phase that has not begun, and an
// effect for the current combat is RegisterScopedRuleEffectForEffect
// with UntilEndOfCombatDuration.
//
// scope must not be ScopeNone: an effect that modifies objects fixes its
// affected set when it begins (CR 611.2c), which a deferred record cannot
// honour, so the API takes a scope and never affected objects, as
// DuringExtraTurnRuleEffectForEffect does. `controller` is the effect's
// "you".
//
// Returns false, writing nothing, for an unknown or unusable scope, no
// mods, or a phase id that is not a combat phase still ahead. Copy on
// write: the list is rebuilt, so an undo snapshot sharing the old
// backing array is not touched.
//
// Caller must hold g.mu (write).
func (g *Game) DuringCombatPhaseRuleEffectForEffect(phaseID int, sourceID uuid.UUID, scope AffectedScope, controller uuid.UUID, mods []Mod, label string) bool {
	if scope == ScopeNone || !KnownAffectedScope(scope) || len(mods) == 0 {
		return false
	}
	if !g.combatPhaseAheadLocked(phaseID) {
		return false
	}
	panicOnLayerModUnderLayerScope(scope, mods, label)
	e := g.buildScopedEffectLocked(sourceID, nil, scope, controller, mods, IndefiniteDuration(), label, timeNowUnixNano())
	next := clonePhaseEffects(g.PhaseEffects)
	next = append(next, PhaseEffect{Turn: g.Turn.Seq, PhaseID: phaseID, Effect: e})
	g.PhaseEffects = next
	return true
}

// combatPhaseAheadLocked reports whether `id` names a combat phase still
// to come in this turn: an entry of the plan, not the phase in progress.
//
// Caller must hold g.mu (write: the plan may be rebuilt).
func (g *Game) combatPhaseAheadLocked(id int) bool {
	if id == 0 || id == g.Turn.PhaseID {
		return false
	}
	g.ensureTurnPlanLocked()
	for _, ps := range g.TurnPlan {
		if ps.PhaseID == id {
			return PhaseOf(ps.Step) == PhaseCombat
		}
	}
	return false
}

// registerPhaseEffectsLocked puts the deferred effects bound to the
// combat phase that has just begun on the registry, each lasting until
// the end of this combat (CR 511.3) and keeping the timestamp it was
// created with (CR 613.7b). Called from finishStepEntryLocked when the
// first step of a combat phase really begins, before the step is
// announced, so the rule is in force for everything that phase does.
//
// Caller must hold g.mu (write).
func (g *Game) registerPhaseEffectsLocked() {
	if len(g.PhaseEffects) == 0 {
		return
	}
	d, ok := g.UntilEndOfCombatDuration()
	if !ok {
		return
	}
	var kept []PhaseEffect
	for _, pe := range g.PhaseEffects {
		if pe.Turn != g.Turn.Seq || pe.PhaseID != g.Turn.PhaseID {
			kept = append(kept, pe)
			continue
		}
		e := pe.Effect
		e.Duration = d
		e.Affected = append([]AffectedObject(nil), e.Affected...)
		e.Mods = cloneMods(e.Mods)
		g.appendBuiltScopedEffectLocked(e)
	}
	g.PhaseEffects = clonePhaseEffects(kept)
}

// sweepUnreachablePhaseEffectsLocked drops every deferred effect bound to
// a phase of another turn: its combat never began (CR 724.1 ended the
// turn, or the phase was skipped), so there is nothing for it to last
// through. Called as a turn begins (sweepUnreachableBoundTriggersLocked).
//
// Caller must hold g.mu (write).
func (g *Game) sweepUnreachablePhaseEffectsLocked() {
	if len(g.PhaseEffects) == 0 {
		return
	}
	var kept []PhaseEffect
	for _, pe := range g.PhaseEffects {
		if pe.Turn == g.Turn.Seq {
			kept = append(kept, pe)
		}
	}
	if len(kept) == len(g.PhaseEffects) {
		return
	}
	g.PhaseEffects = clonePhaseEffects(kept)
}

// clonePhaseEffects deep-copies the list: each record carries slices of
// its own (Mods, Affected), and an undo snapshot must not share them with
// the live game. Nil for an empty list, so a file without the field and a
// game that never used it capture the same.
func clonePhaseEffects(in []PhaseEffect) []PhaseEffect {
	if len(in) == 0 {
		return nil
	}
	out := make([]PhaseEffect, len(in))
	for i, pe := range in {
		pe.Effect.Affected = append([]AffectedObject(nil), pe.Effect.Affected...)
		pe.Effect.Mods = cloneMods(pe.Effect.Mods)
		out[i] = pe
	}
	return out
}
