package game

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// scoped_restrictions_test.go pins #1650's engine half: a restriction
// record whose affected set is a live rule (ScopedEffect.Scope) is
// applied after the layer pass, reaches permanents that did not exist
// when it began, and reads the finished characteristics. The cards are
// in cards/effects/live_mass_restriction_test.go.

func registerRuleRestrictionForTest(t *testing.T, g *Game, scope AffectedScope, controller uuid.UUID, r Restriction) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.RegisterScopedRuleEffectForEffect(uuid.Nil, scope, controller,
			[]Mod{AddRestrictionsMod(r)}, g.UntilEndOfTurnDuration(), "test restriction")
	})
	if !ok {
		t.Fatal("RegisterScopedRuleEffectForEffect registered nothing")
	}
}

// Each new scope matches what it names, read against the record's
// controller, and reaches a creature pushed after registration.
func TestRuleScopedRestrictionScopes(t *testing.T) {
	cases := []struct {
		scope AffectedScope
		// mine, theirs, theirFlyer, theirLate
		want [4]bool
	}{
		{ScopeYourCreatures, [4]bool{true, false, false, false}},
		{ScopeOpponentsCreatures, [4]bool{false, true, true, true}},
		{ScopeCreaturesWithoutFlying, [4]bool{true, true, false, true}},
	}
	for _, c := range cases {
		t.Run(string(c.scope), func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := pushCombatant(t, g, me, "Mine", 2, 2)
			theirs := pushCombatant(t, g, opp, "Theirs", 2, 2)
			flyer := pushCombatant(t, g, opp, "Their Flyer", 1, 1, "flying")

			registerRuleRestrictionForTest(t, g, c.scope, me.ID, CantBlock)
			late := pushCombatant(t, g, opp, "Their Late Arrival", 2, 2)

			for i, id := range []uuid.UUID{mine, theirs, flyer, late} {
				got := scopedEffectChar(t, g, id).Restrictions.Has(CantBlock)
				if got != c.want[i] {
					t.Errorf("creature %d: CantBlock = %v, want %v", i, got, c.want[i])
				}
			}
		})
	}
}

// A live-rule restriction is not a layer effect: the adapter hands the
// layer engine nothing for it, and the fold after the pass applies it.
func TestRuleScopedRestrictionIsFoldedAfterTheLayers(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	registerRuleRestrictionForTest(t, g, ScopeYourCreatures, me.ID, CantBeBlocked)
	var n int
	g.WithWriteLock(func() { n = len(g.scopedEffectContinuousEffectsLocked()) })
	if n != 0 {
		t.Errorf("the layer adapter built %d effects for a live-rule restriction, want 0", n)
	}
	mine := pushCombatant(t, g, me, "Mine", 2, 2)
	if !scopedEffectChar(t, g, mine).Restrictions.Has(CantBeBlocked) {
		t.Error("the fold did not apply the restriction")
	}
}

// ScopeCreaturesWithoutFlying reads a layer-6 result, so a layer mod on
// it would read the keyword half-way through the pass. Registration
// refuses it.
func TestCreaturesWithoutFlyingScopeCarriesOnlyRestrictions(t *testing.T) {
	g := newActiveGame(t)
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), "carries only addRestrictions") {
			t.Errorf("want a registration panic, got %v", r)
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeCreaturesWithoutFlying, g.Seats[0].ID,
			[]Mod{ModifyPTMod(1, 1)}, g.UntilEndOfTurnDuration(), "test")
	})
}
