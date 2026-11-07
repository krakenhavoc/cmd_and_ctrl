package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// phase_effects_test.go pins #5277's engine half (ADR 0188 Decision 13):
// a rules effect bound to one future combat phase waits on
// Game.PhaseEffects, is registered as that combat begins, ends with it,
// and is not in force in any other combat, a newer added one included.
// The card is in cards/effects/bumi_unleashed_test.go.

// pushLandCreatureForTest drops a land creature (Dryad Arbor's type line)
// under `owner`, able to attack.
func pushLandCreatureForTest(t *testing.T, g *Game, owner *Player, name string) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.TypeLine = "Land Creature — Forest Dryad"
	c.Power, c.Toughness = 1, 1
	g.Battlefield.PushTop(c)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == c.InstanceID {
			g.Battlefield.Cards[i].effective = printedEffectiveWith(c)
		}
	}
	return c.InstanceID
}

// bindOnlyLandCreatures binds "only land creatures can attack" to phase
// `phase` and reports whether the engine took it.
func bindOnlyLandCreatures(t *testing.T, g *Game, phase int) bool {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.DuringCombatPhaseRuleEffectForEffect(phase, uuid.Nil, ScopeCreaturesThatArentLands,
			g.Seats[g.Turn.ActiveSeat].ID, []Mod{AddRestrictionsMod(CantAttack)}, "test — only land creatures can attack")
	})
	return ok
}

// TestPhaseEffectAppliesOnlyInItsCombat: bound during the first combat to
// the combat it added, the rule is not in force in the first combat, nor
// in a newer added combat that runs first (CR 500.8), and is in force in
// its own, where a creature that isn't a land can't be declared as an
// attacker and a land creature can (CR 508.1c). It ends with that combat.
func TestPhaseEffectAppliesOnlyInItsCombat(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepBeginCombat)
	bear := pushCombatant(t, g, me, "Bear", 2, 2)
	arbor := pushLandCreatureForTest(t, g, me, "Dryad Arbor")

	bumis := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if len(bumis) != 1 || !bindOnlyLandCreatures(t, g, bumis[0]) {
		t.Fatalf("setup: ids %v, or the binding was refused", bumis)
	}
	if len(g.PhaseEffects) != 1 || scopedEffectCount(g) != 0 {
		t.Fatalf("waiting %d, registered %d; want one waiting and none registered", len(g.PhaseEffects), scopedEffectCount(g))
	}
	if scopedEffectChar(t, g, bear).Restrictions.Has(CantAttack) {
		t.Fatal("the rule is in force in the combat that created it")
	}

	// A newer combat added after this one runs first.
	newer := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if len(newer) != 1 {
		t.Fatalf("setup: ids %v", newer)
	}
	advanceToPhase(t, g, newer[0])
	advanceIntoStep(t, g, StepDeclareAttackers)
	if g.Turn.PhaseID != newer[0] {
		t.Fatalf("in phase %d, want the newer combat %d", g.Turn.PhaseID, newer[0])
	}
	if scopedEffectChar(t, g, bear).Restrictions.Has(CantAttack) {
		t.Fatal("the rule is in force in a newer added combat")
	}
	if len(g.PhaseEffects) != 1 {
		t.Fatalf("the waiting rule was spent on the wrong combat: %+v", g.PhaseEffects)
	}

	advanceToPhase(t, g, bumis[0])
	advanceIntoStep(t, g, StepDeclareAttackers)
	if len(g.PhaseEffects) != 0 || scopedEffectCount(g) != 1 {
		t.Fatalf("in its combat: waiting %d, registered %d; want none waiting and one registered",
			len(g.PhaseEffects), scopedEffectCount(g))
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	if err := g.DeclareAttacker(bear, opp); !errors.Is(err, ErrCantAttack) {
		t.Errorf("a creature that isn't a land attacked in the bound combat: err = %v", err)
	}
	if err := g.DeclareAttacker(arbor, opp); err != nil {
		t.Errorf("a land creature could not attack in the bound combat: %v", err)
	}

	advanceIntoStep(t, g, StepPostcombatMain)
	if scopedEffectCount(g) != 0 || scopedEffectChar(t, g, bear).Restrictions.Has(CantAttack) {
		t.Error("the rule outlived its combat")
	}
}

// TestPhaseEffectRefusesWhatIsNotACombatAhead: only a combat phase still
// to come in this turn can be bound; the combat in progress, a main
// phase, an unknown id and a scope that names nothing are refused.
func TestPhaseEffectRefusesWhatIsNotACombatAhead(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepBeginCombat)
	for _, id := range []int{g.Turn.PhaseID, 4, 99, 0} {
		if bindOnlyLandCreatures(t, g, id) {
			t.Errorf("phase %d was bound", id)
		}
	}
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	ok := true
	g.WithWriteLock(func() {
		ok = g.DuringCombatPhaseRuleEffectForEffect(ids[0], uuid.Nil, ScopeNone, g.Seats[0].ID,
			[]Mod{AddRestrictionsMod(CantAttack)}, "test")
	})
	if ok {
		t.Error("ScopeNone was bound")
	}
	if len(g.PhaseEffects) != 0 {
		t.Errorf("a refusal wrote %+v", g.PhaseEffects)
	}
}

// TestPhaseEffectOfAnotherTurnIsDropped: a rule whose combat never began
// is not registered in a later turn's phase with the same id, and the
// turn-begin sweep drops it.
func TestPhaseEffectOfAnotherTurnIsDropped(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepBeginCombat)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if !bindOnlyLandCreatures(t, g, ids[0]) {
		t.Fatal("setup: the binding was refused")
	}
	g.PhaseEffects[0].Turn = g.Turn.Seq - 1 // as if left over from the turn before
	advanceToPhase(t, g, ids[0])
	if scopedEffectCount(g) != 0 {
		t.Error("a rule of another turn was registered")
	}
	g.WithWriteLock(func() { g.sweepUnreachablePhaseEffectsLocked() })
	if len(g.PhaseEffects) != 0 {
		t.Errorf("the sweep kept %+v", g.PhaseEffects)
	}
}

// TestPhaseEffectCloneAndSnapshot: the waiting rule clones without
// aliasing, comes back from a restore point, and is registered in its
// combat of the restored game.
func TestPhaseEffectCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepBeginCombat)
	bear := pushCombatant(t, g, me, "Bear", 2, 2)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if !bindOnlyLandCreatures(t, g, ids[0]) {
		t.Fatal("setup: the binding was refused")
	}

	cl := g.Clone()
	if len(cl.PhaseEffects) != 1 || cl.PhaseEffects[0].PhaseID != ids[0] {
		t.Fatalf("clone lost the rule: %+v", cl.PhaseEffects)
	}
	cl.PhaseEffects[0].Effect.Mods[0].Restrictions = CantBlock
	if g.PhaseEffects[0].Effect.Mods[0].Restrictions != CantAttack {
		t.Error("the clone aliases the live rule's mods")
	}

	snap, restored := roundTrip(t, g)
	if len(snap.PhaseEffects) != 1 || len(restored.PhaseEffects) != 1 {
		t.Fatalf("snapshot %d, restored %d waiting rules; want 1 and 1", len(snap.PhaseEffects), len(restored.PhaseEffects))
	}
	if got := restored.PhaseEffects[0]; got.Turn != g.Turn.Seq || got.PhaseID != ids[0] ||
		got.Effect.Scope != ScopeCreaturesThatArentLands {
		t.Errorf("restore lost the binding: %+v", got)
	}
	advanceToPhase(t, restored, ids[0])
	if !scopedEffectChar(t, restored, bear).Restrictions.Has(CantAttack) {
		t.Error("the restored rule was not registered in its combat")
	}
}

// TestCreaturesThatArentLandsScopeCarriesOnlyRestrictions: the scope
// reads a layer-4 result, so a layer mod on it is refused at
// registration, waiting or live.
func TestCreaturesThatArentLandsScopeCarriesOnlyRestrictions(t *testing.T) {
	for _, deferred := range []bool{false, true} {
		g := newActiveGame(t)
		advanceIntoStep(t, g, StepBeginCombat)
		ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
		func() {
			defer func() {
				r := recover()
				if s, _ := r.(string); !strings.Contains(s, "carries only addRestrictions") {
					t.Errorf("deferred=%v: want a registration panic, got %v", deferred, r)
				}
			}()
			g.WithWriteLock(func() {
				if deferred {
					g.DuringCombatPhaseRuleEffectForEffect(ids[0], uuid.Nil, ScopeCreaturesThatArentLands, g.Seats[0].ID,
						[]Mod{ModifyPTMod(1, 1)}, "test")
					return
				}
				g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeCreaturesThatArentLands, g.Seats[0].ID,
					[]Mod{ModifyPTMod(1, 1)}, g.UntilEndOfTurnDuration(), "test")
			})
		}()
	}
}

// TestCreaturesThatArentLandsScopeReadsTheFinishedType: a land creature
// is outside the scope and a creature that is not a land is inside it,
// including one that arrives after the record began (CR 611.2c).
func TestCreaturesThatArentLandsScopeReadsTheFinishedType(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbor := pushLandCreatureForTest(t, g, me, "Dryad Arbor")
	registerRuleRestrictionForTest(t, g, ScopeCreaturesThatArentLands, me.ID, CantAttack)
	late := pushCombatant(t, g, opp, "Their Late Arrival", 2, 2)
	if scopedEffectChar(t, g, arbor).Restrictions.Has(CantAttack) {
		t.Error("a land creature is under the restriction")
	}
	if !scopedEffectChar(t, g, late).Restrictions.Has(CantAttack) {
		t.Error("a creature that arrived later is not under the restriction")
	}
}
