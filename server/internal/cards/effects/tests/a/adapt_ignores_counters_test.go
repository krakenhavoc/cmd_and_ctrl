package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adapt_ignores_counters_test.go — #6239, CR 701.46a: "Adapt N" means
// "If this permanent has no +1/+1 counters on it, put N +1/+1 counters
// on it." Biomancer's Familiar's "the next time target creature adapts
// this turn, it adapts as though it had no +1/+1 counters on it", on a
// test oracle, against a creature with "{0}: Adapt 2".

const (
	testAdaptFamiliarOracle = "test-6239-familiar"
	testAdapterOracle       = "test-6239-adapter"
)

func registerAdaptTestCards(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     testAdaptFamiliarOracle,
		Name:         "Test Familiar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: the next time target creature adapts this turn, it adapts as though it had no +1/+1 counters on it",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NextAdaptIgnoresCounters("Test Familiar"),
		}},
	})
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     testAdapterOracle,
		Name:         "Test Adapter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:  "{0}: Adapt 2",
			Effect: func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 2) },
		}},
	})
}

func plusOnes(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the creature is gone")
	}
	return c.Counters[game.CounterPlusOne]
}

func activateAndResolve(t *testing.T, g *game.Game, me uuid.UUID, src uuid.UUID, targets []game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me, src, 0, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// A creature with two +1/+1 counters adapts 2 and gets two more; the
// next adapt is an ordinary one and does nothing.
func TestAdaptAsThoughItHadNoCounters(t *testing.T) {
	registerAdaptTestCards(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	familiar := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Familiar", TypeLine: "Creature — Rat", OracleID: testAdaptFamiliarOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	adapter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Adapter", TypeLine: "Creature — Beast", OracleID: testAdapterOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 2}})

	// Without the Familiar, adapt does nothing to a creature with counters.
	activateAndResolve(t, g, me.ID, adapter, nil)
	if n := plusOnes(t, g, adapter); n != 2 {
		t.Fatalf("an ordinary adapt with counters on it: %d counters, want 2", n)
	}

	activateAndResolve(t, g, me.ID, familiar, []game.TargetRef{{Kind: game.TargetCard, ID: adapter}})
	activateAndResolve(t, g, me.ID, adapter, nil)
	if n := plusOnes(t, g, adapter); n != 4 {
		t.Fatalf("adapt as though it had no counters: %d counters, want 4", n)
	}
	activateAndResolve(t, g, me.ID, adapter, nil)
	if n := plusOnes(t, g, adapter); n != 4 {
		t.Fatalf("the adapt after that: %d counters, want 4 (the record is spent)", n)
	}
}
