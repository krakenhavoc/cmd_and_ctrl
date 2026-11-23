package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_cost_tax_test.go — #5948: ModSpellsCostMoreForOpponents, the timed
// "spells your opponents cast cost {N} more" record behind Tax Collector's Tax
// bullet. The catalog half is in cards/effects.

func gapTaxPrice(t *testing.T, g *Game, caster uuid.UUID) string {
	t.Helper()
	card := Card{InstanceID: uuid.New(), Name: "Test Spell", TypeLine: "Sorcery", ManaCost: "{1}{R}", Owner: caster, Controller: caster}
	base, err := ParseCost("{1}{R}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	var out ParsedCost
	g.WithWriteLock(func() {
		var err error
		out, err = g.ApplyCostModifiersForEffect(base, CostQuery{Game: g, Card: card, Controller: caster, FromZone: ZoneHand})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
	})
	return out.String()
}

func TestSpellsCostMoreForOpponentsTaxesOnlyOpponentsCasts(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := gapTaxPrice(t, g, opp.ID); got != "{1}{R}" {
		t.Fatalf("baseline price %q", got)
	}
	g.WithWriteLock(func() {
		if !g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeGame, me.ID, []Mod{SpellsCostMoreForOpponentsMod(1)}, g.UntilEndOfTurnDuration(), "test tax") {
			t.Fatal("the tax was not registered")
		}
	})
	if got := gapTaxPrice(t, g, opp.ID); got != "{2}{R}" {
		t.Errorf("an opponent's spell costs %q, want {2}{R}", got)
	}
	if got := gapTaxPrice(t, g, me.ID); got != "{1}{R}" {
		t.Errorf("the taxer's own spell costs %q, want {1}{R}", got)
	}
}

func TestSpellsCostMoreForOpponentsNeedsAnAmount(t *testing.T) {
	if scopedCostTaxModProblem(SpellsCostMoreForOpponentsMod(0)) == "" {
		t.Error("a zero tax is accepted")
	}
	if scopedCostTaxModProblem(SpellsCostMoreForOpponentsMod(2)) != "" {
		t.Error("a sound tax is refused")
	}
}
