package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// only_these_can_attack_test.go pins #6087's "Only the chosen creatures
// can attack during that combat phase" (Last Night Together): a live
// "can't attack" on every creature but the named objects, bound to one
// added combat. The card is in cards/effects.

func bindOnlyThese(t *testing.T, g *Game, phase int, ids ...uuid.UUID) bool {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.OnlyTheseCanAttackDuringCombatPhaseForEffect(phase, uuid.Nil, g.Seats[g.Turn.ActiveSeat].ID, ids, "test — only the chosen creatures can attack")
	})
	return ok
}

// TestOnlyTheseCanAttackInItsCombat: not in force before its combat; in
// it, the chosen creature attacks and every other creature can't — an
// opponent's, and one that arrived after the rule was made; it ends with
// that combat.
func TestOnlyTheseCanAttackInItsCombat(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceIntoStep(t, g, StepBeginCombat)
	chosen := pushCombatant(t, g, me, "Chosen", 2, 2)
	other := pushCombatant(t, g, me, "Other", 2, 2)
	theirs := pushCombatant(t, g, opp, "Theirs", 2, 2)

	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if len(ids) != 1 || !bindOnlyThese(t, g, ids[0], chosen) {
		t.Fatalf("setup: ids %v, or the binding was refused", ids)
	}
	if scopedEffectChar(t, g, other).Restrictions.Has(CantAttack) {
		t.Fatal("the rule is in force before its combat")
	}

	advanceToPhase(t, g, ids[0])
	late := pushCombatant(t, g, me, "Late", 2, 2)
	g.layerVersion.Add(1) // an engine entry bumps the layers; the helper does not
	advanceIntoStep(t, g, StepDeclareAttackers)
	if g.Turn.PhaseID != ids[0] {
		t.Fatalf("in phase %d, want the bound combat %d", g.Turn.PhaseID, ids[0])
	}
	for name, id := range map[string]uuid.UUID{"another creature": other, "a creature that arrived later": late, "an opponent's creature": theirs} {
		if !scopedEffectChar(t, g, id).Restrictions.Has(CantAttack) {
			t.Errorf("%s may attack in the bound combat", name)
		}
	}
	if err := g.DeclareAttacker(other, opp.ID); !errors.Is(err, ErrCantAttack) {
		t.Errorf("another creature attacked in the bound combat: err = %v", err)
	}
	if err := g.DeclareAttacker(chosen, opp.ID); err != nil {
		t.Errorf("the chosen creature could not attack: %v", err)
	}

	advanceIntoStep(t, g, StepPostcombatMain)
	if scopedEffectChar(t, g, other).Restrictions.Has(CantAttack) {
		t.Error("the rule outlived its combat")
	}
}

// TestOnlyTheseCanAttackNamesTheObject: a chosen creature that has become
// a new object (CR 400.7) is not chosen any more.
func TestOnlyTheseCanAttackNamesTheObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepBeginCombat)
	chosen := pushCombatant(t, g, me, "Chosen", 2, 2)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if !bindOnlyThese(t, g, ids[0], chosen) {
		t.Fatal("setup: the binding was refused")
	}
	g.WithWriteLock(func() {
		findBattlefieldCard(g, chosen).ObjectEpoch++
		g.layerVersion.Add(1)
	})
	advanceToPhase(t, g, ids[0])
	if !scopedEffectChar(t, g, chosen).Restrictions.Has(CantAttack) {
		t.Error("a new object was treated as the chosen one")
	}
}

// TestOnlyTheseCanAttackRefusesAndRestores: nothing on the battlefield to
// spare is refused; the waiting rule survives a restore point; and the
// scope's shape is checked.
func TestOnlyTheseCanAttackRefusesAndRestores(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepBeginCombat)
	other := pushCombatant(t, g, me, "Other", 2, 2)
	chosen := pushCombatant(t, g, me, "Chosen", 2, 2)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if bindOnlyThese(t, g, ids[0], uuid.New()) {
		t.Error("a binding sparing nothing on the battlefield was taken")
	}
	if !bindOnlyThese(t, g, ids[0], chosen) {
		t.Fatal("setup: the binding was refused")
	}
	_, restored := roundTrip(t, g)
	if len(restored.PhaseEffects) != 1 || restored.PhaseEffects[0].Effect.Scope != ScopeCreaturesExceptObjects {
		t.Fatalf("restore lost the binding: %+v", restored.PhaseEffects)
	}
	advanceToPhase(t, restored, ids[0])
	if !scopedEffectChar(t, restored, other).Restrictions.Has(CantAttack) ||
		scopedEffectChar(t, restored, chosen).Restrictions.Has(CantAttack) {
		t.Error("the restored rule does not spare exactly the chosen creature")
	}

	ref := []ObjectRef{{ID: chosen}}
	for name, c := range map[string]struct {
		scope AffectedScope
		mods  []Mod
	}{
		"objects outside the scope": {ScopeCreatures, []Mod{OnlyTheseCanAttackMod(ref)}},
		"no object":                 {ScopeCreaturesExceptObjects, []Mod{AddRestrictionsMod(CantAttack)}},
		"another kind":              {ScopeCreaturesExceptObjects, []Mod{AddKeywordsMod("haste")}},
	} {
		if exceptObjectsScopeProblem(c.scope, c.mods) == "" {
			t.Errorf("%s: no problem reported", name)
		}
	}
	if p := exceptObjectsScopeProblem(ScopeCreaturesExceptObjects, []Mod{OnlyTheseCanAttackMod(ref)}); p != "" {
		t.Errorf("a sound record reported %q", p)
	}
}
