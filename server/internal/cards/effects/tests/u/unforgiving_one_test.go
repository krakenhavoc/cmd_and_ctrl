package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const unforgivingOneOracle = "9ec2114c-023b-45c3-a44c-6a9f6462bc85"

func TestUnforgivingOneIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(unforgivingOneOracle)
	if !ok || spec.Name != "Unforgiving One" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestUnforgivingOneReturnsACreatureUpToTheModifiedCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.B12Push(g, me.ID, "Unforgiving One", "Creature — Spirit", unforgivingOneOracle, 2, 3)
	// Unforgiving One itself carries a counter, so it is modified; X = 2 with the bear.
	effectstest.S82SetCounters(g, one, map[string]int{"+1/+1": 1})
	effectstest.ModifiedBear(g, me.ID)
	small := effectstest.GraveCreatureMV(me, "Two Drop", "{1}{B}")
	big := effectstest.GraveCreatureMV(me, "Three Drop", "{2}{B}")

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(one, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, big) || !effectstest.HasID(pick.PickTargetCards, small) {
		t.Fatalf("offered = %v, want only the two-drop", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, small) {
		t.Error("the two-drop did not return")
	}
	if !effects.OnBattlefield(g, small) || effects.OnBattlefield(g, big) {
		t.Error("the three-drop returned")
	}
}

func TestUnforgivingOneWithNothingModifiedOnlyReturnsAZeroDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.B12Push(g, me.ID, "Unforgiving One", "Creature — Spirit", unforgivingOneOracle, 2, 3)
	one1 := effectstest.GraveCreatureMV(me, "One Drop", "{B}")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(one, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, one1) {
		t.Error("a mana value 1 creature returned with no modified creatures")
	}
}
