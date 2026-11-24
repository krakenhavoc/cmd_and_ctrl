package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_cost_reduction_test.go — #6093: ModSpellsCostLessForPlayer, the
// timed "spells <a player> casts cost {N} less" record behind Commander
// Liara Portyr's and Will Kenrith's reductions. The catalog half is in
// cards/effects.

func reductionPrice(t *testing.T, g *Game, caster uuid.UUID, typeLine, cost string, from ZoneKind) string {
	t.Helper()
	card := Card{InstanceID: uuid.New(), Name: "Test Spell", TypeLine: typeLine, ManaCost: cost, Owner: caster, Controller: caster}
	base, err := ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	var out ParsedCost
	g.WithWriteLock(func() {
		var err error
		out, err = g.ApplyCostModifiersForEffect(base, CostQuery{Game: g, Card: card, Controller: caster, FromZone: from})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
	})
	return out.String()
}

func TestSpellsCostLessForPlayerReducesOnlyThatPlayersSpellsOfThoseTypes(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if !g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeGame, me.ID,
			[]Mod{SpellsCostLessForPlayerMod(opp.ID, 2, "instant", "sorcery", "planeswalker")}, g.UntilYourNextTurnDuration(me.ID), "test reduction") {
			t.Fatal("the reduction was not registered")
		}
	})
	if got := reductionPrice(t, g, opp.ID, "Instant", "{3}{U}", ZoneHand); got != "{1}{U}" {
		t.Errorf("the player's instant costs %q, want {1}{U}", got)
	}
	if got := reductionPrice(t, g, opp.ID, "Legendary Planeswalker — Test", "{1}{U}", ZoneHand); got != "{U}" {
		t.Errorf("the player's planeswalker costs %q, want {U} (floored at the generic part)", got)
	}
	if got := reductionPrice(t, g, opp.ID, "Creature — Bear", "{3}{G}", ZoneHand); got != "{3}{G}" {
		t.Errorf("the player's creature costs %q, want it unreduced", got)
	}
	if got := reductionPrice(t, g, me.ID, "Instant", "{3}{U}", ZoneHand); got != "{3}{U}" {
		t.Errorf("another player's instant costs %q, want it unreduced", got)
	}
	// Until the controller's next turn: gone once that turn begins.
	advanceToStepOfSeat(t, g, 0, StepUpkeep)
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	advanceToStepOfSeat(t, g, 0, StepUpkeep)
	if got := reductionPrice(t, g, opp.ID, "Instant", "{3}{U}", ZoneHand); got != "{3}{U}" {
		t.Errorf("the instant costs %q after the controller's next turn began, want it unreduced", got)
	}
}

func TestSpellsCostLessForPlayerFromExileOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeGame, me.ID,
			[]Mod{ExiledSpellsCostLessForPlayerMod(me.ID, 3)}, g.UntilEndOfTurnDuration(), "test reduction")
	})
	if got := reductionPrice(t, g, me.ID, "Creature — Bear", "{4}{G}", ZoneExile); got != "{1}{G}" {
		t.Errorf("a spell cast from exile costs %q, want {1}{G}", got)
	}
	if got := reductionPrice(t, g, me.ID, "Creature — Bear", "{4}{G}", ZoneHand); got != "{4}{G}" {
		t.Errorf("a spell cast from hand costs %q, want it unreduced", got)
	}
}

func TestSpellsCostLessForPlayerModProblems(t *testing.T) {
	if scopedCostReductionModProblem(SpellsCostLessForPlayerMod(uuid.Nil, 2)) == "" {
		t.Error("a reduction naming no player is accepted")
	}
	if scopedCostReductionModProblem(ExiledSpellsCostLessForPlayerMod(uuid.New(), -1)) == "" {
		t.Error("a negative reduction is accepted")
	}
	if scopedCostReductionModProblem(ExiledSpellsCostLessForPlayerMod(uuid.New(), 2)) != "" {
		t.Error("a sound reduction is refused")
	}
}
