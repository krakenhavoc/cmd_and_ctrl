package game

import (
	"testing"

	"github.com/google/uuid"
)

// your_creatures_except_objects_test.go pins S82 OLD b411's "Other creatures
// that player controls can't block this turn" (Mark for Death): a live
// "can't block" on one player's creatures except the named object. The card
// is in cards/effects.

func bindOthersCantBlock(t *testing.T, g *Game, player, spared uuid.UUID) bool {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.OtherCreaturesOfPlayerCantBlockForEffect(uuid.Nil, player, spared, g.UntilEndOfTurnDuration(), "test — other creatures can't block")
	})
	return ok
}

// TestOtherCreaturesOfPlayerCantBlock: the named creature may still block, the
// player's other creatures — and one that arrived afterwards — can't, and no one
// else's creature is touched.
func TestOtherCreaturesOfPlayerCantBlock(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	spared := pushCombatant(t, g, opp, "Spared", 2, 2)
	other := pushCombatant(t, g, opp, "Other", 2, 2)
	mine := pushCombatant(t, g, me, "Mine", 2, 2)
	thirds := pushCombatant(t, g, third, "Third", 2, 2)
	if !bindOthersCantBlock(t, g, opp.ID, spared) {
		t.Fatal("setup: the binding was refused")
	}
	late := pushCombatant(t, g, opp, "Late", 2, 2)
	g.WithWriteLock(func() { g.layerVersion.Add(1) })
	for name, want := range map[string]struct {
		id   uuid.UUID
		bars bool
	}{
		"the spared creature":         {spared, false},
		"another creature of theirs":  {other, true},
		"one that arrived later":      {late, true},
		"the caster's creature":       {mine, false},
		"another opponent's creature": {thirds, false},
	} {
		if got := scopedEffectChar(t, g, want.id).Restrictions.Has(CantBlock); got != want.bars {
			t.Errorf("%s: can't block = %v, want %v", name, got, want.bars)
		}
	}
	if bindOthersCantBlock(t, g, opp.ID, uuid.New()) {
		t.Error("a binding sparing nothing on the battlefield was taken")
	}
}

// TestOtherCreaturesOfPlayerCantBlockRestores: the rule survives a restore
// point, and a spared creature that became a new object (CR 400.7) is no longer
// spared.
func TestOtherCreaturesOfPlayerCantBlockRestores(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spared := pushCombatant(t, g, opp, "Spared", 2, 2)
	other := pushCombatant(t, g, opp, "Other", 2, 2)
	_ = me
	if !bindOthersCantBlock(t, g, opp.ID, spared) {
		t.Fatal("setup: the binding was refused")
	}
	_, restored := roundTrip(t, g)
	restored.WithWriteLock(func() { restored.layerVersion.Add(1) })
	if !scopedEffectChar(t, restored, other).Restrictions.Has(CantBlock) ||
		scopedEffectChar(t, restored, spared).Restrictions.Has(CantBlock) {
		t.Error("the restored rule does not spare exactly the named creature")
	}
	restored.WithWriteLock(func() {
		findBattlefieldCard(restored, spared).ObjectEpoch++
		restored.layerVersion.Add(1)
	})
	if !scopedEffectChar(t, restored, spared).Restrictions.Has(CantBlock) {
		t.Error("a new object was treated as the spared one")
	}
	ref := []ObjectRef{{ID: spared}}
	if p := exceptObjectsScopeProblem(ScopeYourCreaturesExceptObjects, []Mod{OnlyTheseCanAttackMod(ref)}); p != "" {
		t.Errorf("a sound record reported %q", p)
	}
	if p := exceptObjectsScopeProblem(ScopeYourCreaturesExceptObjects, []Mod{AddRestrictionsMod(CantBlock)}); p == "" {
		t.Error("a record naming no object was accepted")
	}
}
