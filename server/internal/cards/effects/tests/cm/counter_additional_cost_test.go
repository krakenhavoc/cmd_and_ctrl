package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #6121: Hierophant Bio-Titan's Frenzied Metabolism — "As an additional
// cost to cast this spell, you may remove any number of +1/+1 counters
// from among creatures you control. This spell costs {2} less to cast for
// each counter removed this way." CR 601.2b announces the counters, CR
// 601.2f prices the cast with them, CR 601.2h removes them.

const titanOracle = "test-6121-frenzied-metabolism"

func registerTitan(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:       titanOracle,
		Name:           titanOracle,
		AdditionalCost: effects.RemoveAnyNumberOfCountersAmongCost(game.CounterPlusOne, "creatures you control", effects.Creature()),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessPerCounterRemoved("{2}", "This spell costs {2} less to cast for each counter removed this way"),
		},
	})
}

func counteredCreature(g *game.Game, owner uuid.UUID, name string, n int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: owner, Controller: owner,
			Counters: map[string]int{game.CounterPlusOne: n},
		})
	})
	return id
}

func plusOnes(g *game.Game, id uuid.UUID) int {
	return effectstest.FindBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne]
}

func TestRemovingCountersPaysTwoLessPerCounter(t *testing.T) {
	registerTitan(t)
	g, me, _ := effectstest.TapTable(t)
	titan := effectstest.HandCardOf(g, me, "Titan", "Creature — Tyranid", "{10}{G}{G}", titanOracle)
	a := counteredCreature(g, me.ID, "Bear A", 2)
	b := counteredCreature(g, me.ID, "Bear B", 3)
	g.WithWriteLock(func() {
		effectstest.FillPoolColored(me, "G", 2)
		effectstest.FillPool(me, 2)
	})

	// Three counters take {6} off {10}: {4}{G}{G} is more than the pool.
	three := game.CastSpellParams{Strict: true, CounterSourceIDs: []uuid.UUID{a, b}, CounterCounts: []int{1, 2}}
	if err := g.CastSpell(me.ID, titan, three); err == nil {
		t.Fatal("three counters and four mana paid {10}{G}{G}")
	}
	if plusOnes(g, a) != 2 || plusOnes(g, b) != 3 {
		t.Fatal("a refused cast removed counters")
	}
	// Five counters take {10} off: {G}{G} is left, and {2} of the pool is
	// over.
	five := game.CastSpellParams{Strict: true, CounterSourceIDs: []uuid.UUID{a, b}, CounterCounts: []int{2, 3}}
	if err := g.CastSpell(me.ID, titan, five); err != nil {
		t.Fatalf("five counters and {G}{G}: %v", err)
	}
	if plusOnes(g, a) != 0 || plusOnes(g, b) != 0 {
		t.Errorf("counters left: %d and %d, want none", plusOnes(g, a), plusOnes(g, b))
	}
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("%d mana left in the pool, want 2 — the cast cost {G}{G}", n)
	}
	item := g.StackMeta[titan]
	if item == nil || item.Paid.CountersRemoved != 5 {
		t.Errorf("paid record = %+v, want five counters removed", item)
	}
}

// "You may remove any number" includes none: the cast is legal at full
// price with no counters named, and a creature can't give more counters
// than it has.
func TestRemovingNoCountersIsFullPrice(t *testing.T) {
	registerTitan(t)
	g, me, _ := effectstest.TapTable(t)
	titan := effectstest.HandCardOf(g, me, "Titan", "Creature — Tyranid", "{10}{G}{G}", titanOracle)
	a := counteredCreature(g, me.ID, "Bear A", 1)
	g.WithWriteLock(func() {
		effectstest.FillPoolColored(me, "G", 2)
		effectstest.FillPool(me, 10)
	})
	tooMany := game.CastSpellParams{Strict: true, CounterSourceIDs: []uuid.UUID{a}, CounterCounts: []int{2}}
	if err := g.CastSpell(me.ID, titan, tooMany); err == nil {
		t.Fatal("two counters came off a creature with one")
	}
	if err := g.CastSpell(me.ID, titan, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("no counters at full price: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want the whole {10}{G}{G} spent", n)
	}
	if plusOnes(g, a) != 1 {
		t.Error("a counter came off with none named")
	}
}

// The view ships the counter block on the hand card's additional cost.
func TestTheCastViewShipsTheCounterBlock(t *testing.T) {
	registerTitan(t)
	g, me, _ := effectstest.TapTable(t)
	titan := effectstest.HandCardOf(g, me, "Titan", "Creature — Tyranid", "{10}{G}{G}", titanOracle)
	a := counteredCreature(g, me.ID, "Bear A", 2)
	v := protocol.ViewOfGameFor(g, me.ID.String())
	var ac *protocol.AdditionalCostView
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if c := &s.Hand.Cards[i]; c.InstanceID == titan.String() {
				ac = c.AdditionalCost
			}
		}
	}
	if ac == nil {
		t.Fatal("no additional cost on the hand card's view")
	}
	if !ac.CounterCostAmong || !ac.CounterCostVariable || ac.CounterCostN != 0 || ac.CounterCostKind != game.CounterPlusOne {
		t.Errorf("counter block = %+v, want a variable among +1/+1 block with no floor", ac.CounterCostView)
	}
	if len(ac.CounterCostOptions) != 1 || ac.CounterCostOptions[0].CardID != a.String() {
		t.Errorf("counter options = %+v, want the one creature with counters", ac.CounterCostOptions)
	}
}

func TestRegisterRefusesAnOptionalCounterCost(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an optional counter-removal cost registered")
		}
		delete(effects.Registry, "test-6121-bad-counters")
	}()
	oc := *effects.RemoveAnyNumberOfCountersAmongCost(game.CounterPlusOne, "creatures you control", effects.Creature())
	oc.Optional, oc.Key = true, "test"
	effects.Register(effects.Spec{OracleID: "test-6121-bad-counters", Name: "test-6121-bad-counters", OptionalCosts: []game.AdditionalCost{oc}})
}
