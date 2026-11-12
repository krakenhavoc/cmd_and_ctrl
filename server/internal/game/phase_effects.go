package game

import (
	"fmt"

	"github.com/google/uuid"
)

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
//
// Since #5685 (ADR 0188 amendment 2026-11-11, Decision 14) a waiting
// record may instead name objects, fixed as the spell resolves:
// Illusionist's Gambit's "Each of those creatures attacks that combat if
// able. They can't attack you or planeswalkers you control that combat."
// See DuringCombatPhaseObjectsEffectForEffect.

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

// DuringCombatPhaseObjectsEffectForEffect attaches an attack requirement
// or an attack restriction to combat phase `phaseID` of this turn, over
// the permanents `objects` as the objects they are now: "Each of those
// creatures attacks that combat if able. They can't attack you or
// planeswalkers you control that combat" (Illusionist's Gambit, #5685,
// ADR 0188 Decision 14). The effect is created now and starts as that
// combat begins.
//
// The set is the card's own text ("those creatures"), fixed as the spell
// resolves, so it never grows. Each member is pinned as the object it is
// (CR 400.7) under the player who controls it now. A member whose
// controller changes while the record waits is dropped at once
// (dropPhaseEffectMemberLocked), and as that combat begins so is one
// that has left the battlefield or that another player controls. A
// record left with no member is not registered.
//
// Only a mod that changes no characteristic and no controller is taken
// (phaseObjectsModKinds): CR 611.2c fixes the set of a characteristic-
// changing effect "when that continuous effect begins", which for a
// waiting record is later than the moment the card names its objects.
// Any other kind panics, as a programming error in the card.
//
// The phase is held to DuringCombatPhaseRuleEffectForEffect's rule: a
// combat phase still to come in this turn's plan. Returns false, writing
// nothing, for such a phase id, for no mods, or when none of `objects`
// is on the battlefield. `controller` is the effect's "you". Copy on
// write.
//
// Caller must hold g.mu (write).
func (g *Game) DuringCombatPhaseObjectsEffectForEffect(phaseID int, sourceID uuid.UUID, objects []uuid.UUID, controller uuid.UUID, mods []Mod, label string) bool {
	if len(mods) == 0 || !g.combatPhaseAheadLocked(phaseID) {
		return false
	}
	for _, m := range mods {
		if !phaseObjectsModKinds[m.Kind] {
			panic(fmt.Sprintf("game: phase effect %q: a rule over named objects carries only an attack requirement or restriction, got %q", label, m.Kind))
		}
	}
	var affected []AffectedObject
	seen := make(map[uuid.UUID]bool, len(objects))
	for _, id := range objects {
		c, ok := g.battlefieldCardLocked(id)
		if !ok || seen[id] {
			continue
		}
		seen[id] = true
		pin := PinObject(id, c.EnteredBattlefieldAt)
		pin.Controller = c.Controller
		affected = append(affected, pin)
	}
	if len(affected) == 0 {
		return false
	}
	e := g.buildScopedEffectLocked(sourceID, affected, ScopeNone, controller, mods, IndefiniteDuration(), label, timeNowUnixNano())
	next := clonePhaseEffects(g.PhaseEffects)
	next = append(next, PhaseEffect{Turn: g.Turn.Seq, PhaseID: phaseID, Effect: e})
	g.PhaseEffects = next
	return true
}

// phaseObjectsModKinds are the mods a waiting record over named objects
// may carry: each changes a rule of the declaration (CR 508.1c, 508.1d),
// never a characteristic or a controller.
var phaseObjectsModKinds = map[ModKind]bool{
	ModAddAttackRequirement: true,
	ModCantAttackPlayer:     true,
	ModAddRestrictions:      true,
}

// phaseEffectMembersStillBoundLocked is the affected set of a waiting
// record over named objects as its combat begins: each member that is
// still that object on the battlefield under the controller it was
// pinned with. The rest are dropped (ADR 0188 Decision 14).
//
// Caller must hold g.mu.
func (g *Game) phaseEffectMembersStillBoundLocked(set []AffectedObject) []AffectedObject {
	var out []AffectedObject
	for _, a := range set {
		c, ok := g.battlefieldCardLocked(a.ID)
		if !ok || !memberMatches(a, c) {
			continue
		}
		if a.Controller != uuid.Nil && c.Controller != a.Controller {
			continue
		}
		out = append(out, a)
	}
	return out
}

// dropPhaseEffectMemberLocked takes the permanent `id` out of every
// waiting record that names it, and drops a record left naming nothing.
// Called when a permanent's controller changes (emitControlChangesLocked):
// a control change before that combat begins frees the creature for good
// (ADR 0188 Decision 14). A live-scope record names no objects and is
// untouched. Copy on write.
//
// Caller must hold g.mu (write).
func (g *Game) dropPhaseEffectMemberLocked(id uuid.UUID) {
	named := false
	for _, pe := range g.PhaseEffects {
		for _, a := range pe.Effect.Affected {
			named = named || a.ID == id
		}
	}
	if !named {
		return
	}
	var kept []PhaseEffect
	for _, pe := range clonePhaseEffects(g.PhaseEffects) {
		if pe.Effect.Scope != ScopeNone {
			kept = append(kept, pe)
			continue
		}
		var members []AffectedObject
		for _, a := range pe.Effect.Affected {
			if a.ID != id {
				members = append(members, a)
			}
		}
		if len(members) == 0 {
			continue
		}
		pe.Effect.Affected = members
		kept = append(kept, pe)
	}
	g.PhaseEffects = kept
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
		if e.Scope == ScopeNone {
			// #5685, Decision 14: a record over named objects keeps only
			// the members still bound, and is spent unregistered when
			// none is.
			if e.Affected = g.phaseEffectMembersStillBoundLocked(e.Affected); len(e.Affected) == 0 {
				continue
			}
		}
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
