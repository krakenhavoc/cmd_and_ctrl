package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phasing_extras_test.go — #6201: Teferi's Imp's "whenever this creature
// phases out" and "whenever this creature phases in" on a test oracle,
// and Spatial Binding's target, through the real trigger harvest.
//
//	603.10b Abilities that trigger when a permanent phases out look back
//	        in time.

const (
	testPhasingImpOracle     = "test-6201-phasing-imp"
	testPhasingBindingOracle = "test-6201-binding"
)

func registerPhasingTestCards(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:        testPhasingImpOracle,
		Name:            "Test Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"phasing"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisPhasesOut("Test Imp — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
			effects.WheneverThisPhasesIn("Test Imp — you gain 5 life", effects.Do(effects.GainLife{Amount: 5})),
		},
	})
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     testPhasingBindingOracle,
		Name:         "Test Binding",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: Until your next upkeep, target permanent can't phase out.",
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.TargetCantPhaseOutUntilYourNextUpkeep("Test Binding"),
		}},
	})
}

func TestAPermanentsOwnPhasesOutAndInTriggersFire(t *testing.T) {
	registerPhasingTestCards(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	imp := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Imp", TypeLine: "Creature — Imp",
		OracleID: testPhasingImpOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"phasing"}})
	life := me.Life
	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, imp) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPhasedOutForEffect(imp) {
		t.Fatal("the imp did not phase out")
	}
	if me.Life != life+2 {
		t.Fatalf("life = %d, want %d: its own \"whenever this phases out\" did not fire (CR 603.10b)", me.Life, life+2)
	}
	g.WithWriteLock(func() { _ = g.PhaseInForEffect(imp) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+7 {
		t.Fatalf("life = %d, want %d: \"whenever this phases in\" did not fire", me.Life, life+7)
	}
}

func TestTargetCantPhaseOutBlocksAnEffectsPhaseOut(t *testing.T) {
	registerPhasingTestCards(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	binding := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Binding", TypeLine: "Enchantment",
		OracleID: testPhasingBindingOracle, Owner: me.ID, Controller: me.ID})
	imp := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Imp", TypeLine: "Creature — Imp",
		OracleID: testPhasingImpOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"phasing"}})
	if err := g.ActivateCatalogAbility(me.ID, binding, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: imp}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.CantPhaseOutForEffect(imp) {
		t.Fatal("the target can phase out")
	}
	life := me.Life
	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, imp) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.IsPhasedOutForEffect(imp) || me.Life != life {
		t.Fatal("a permanent that can't phase out phased out (and its trigger fired)")
	}
}
