package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestOuroboroidBoardSaturatesAtTheCeiling is the issue's board: a lone
// Ouroboroid doubles itself every combat (X is its own power, and it
// puts X counters on itself). An unclamped board is past 2^31 at the 32nd
// combat and past 2^63 at the 64th, where it wraps negative. Seventy combats must
// finish without a panic and leave power AT the ceiling, never past it
// and never negative.
func TestOuroboroidBoardSaturatesAtTheCeiling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ouro := effectstest.PushCatalogPermanent(g, me.ID, "Ouroboroid", "Creature — Plant Wurm", effectstest.OuroboroidOracle, false)
	// Seventy combats is seventy turns each: keep the table from decking.
	for _, seat := range g.Seats {
		for i := 0; i < 100; i++ {
			effectstest.PushLibraryCardForTest(seat, game.NewCard("basic-filler", uuid.Nil))
		}
	}
	// A second creature, so the per-creature placement loop runs too.
	bear := effectstest.OuroboroidCreature(g, me.ID, "Bear", 1, 1)

	for combat := 1; combat <= 70; combat++ {
		effectstest.OuroboroidToCombat(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		got := effectstest.CounterCount(g, ouro, game.CounterPlusOne)
		if got < 0 || got > game.MaxCount {
			t.Fatalf("combat %d: Ouroboroid has %d counters, outside [0, MaxCount=%d]", combat, got, game.MaxCount)
		}
		if p := effectstest.EffectivePowerWithCounters(t, g, ouro); p < 0 || p > game.MaxCount {
			t.Fatalf("combat %d: Ouroboroid power %d is outside [0, MaxCount=%d]", combat, p, game.MaxCount)
		}
	}
	if got := effectstest.CounterCount(g, ouro, game.CounterPlusOne); got != game.MaxCount {
		t.Errorf("Ouroboroid ended with %d counters, want exactly the ceiling %d", got, game.MaxCount)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != game.MaxCount {
		t.Errorf("the bear ended with %d counters, want the ceiling %d", got, game.MaxCount)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, ouro); p != game.MaxCount {
		t.Errorf("Ouroboroid power = %d, want the ceiling %d", p, game.MaxCount)
	}
	if effectstest.CountCappedEvents(g, "counters") == 0 {
		t.Error("the clamp was never announced: no EventCountCapped for the counters")
	}
}

// TestHugePlacementThroughDoublersCannotWrapTheDelta: a placement asked
// for at 2^62 and then doubled three times would wrap `int` (to 0 at
// 2^64) if the replacement pipeline did not pin CounterDelta between
// passes. It must land at the ceiling.
func TestHugePlacementThroughDoublersCannotWrapTheDelta(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	target := effectstest.OuroboroidCreature(g, me.ID, "Bear", 1, 1)
	for i := 0; i < 3; i++ {
		effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	}
	if err := g.AddCounter(target, game.CounterPlusOne, 1<<62); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := effectstest.CounterCount(g, target, game.CounterPlusOne); got != game.MaxCount {
		t.Errorf("a 2^62 placement through three doublers left %d counters, want the ceiling %d", got, game.MaxCount)
	}
}

// TestMarwynAtTheCeilingTapsForABoundedPool is the allocation half: a
// Marwyn at MaxCount power used to build a 2^30-slot string (and at 2^40
// a 3.3 TB one). It must tap, add exactly MaxProducedMana green mana,
// and say it capped the production.
func TestMarwynAtTheCeilingTapsForABoundedPool(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	marwyn := effectstest.PushCatalogPermanent(g, me.ID, "Marwyn, the Nurturer", "Legendary Creature — Elf Druid", effectstest.B10MarwynOracle, false)
	effectstest.OuroboroidSetCounters(g, marwyn, game.MaxCount)

	if err := g.ActivateManaAbility(me.ID, marwyn, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.Total(); got != game.MaxProducedMana {
		t.Errorf("pool holds %d mana, want exactly the production ceiling %d", got, game.MaxProducedMana)
	}
	for i, tok := range me.ManaPool.Units() {
		if tok.Color != "G" {
			t.Fatalf("pool[%d] is %q, want G", i, tok.Color)
		}
	}
	if effectstest.CountCappedEvents(g, "mana") != 1 {
		t.Errorf("want exactly one mana EventCountCapped, got %d", effectstest.CountCappedEvents(g, "mana"))
	}
}

// TestUnboundedProducedFuncsStayBounded covers the string builders: any
// card that repeats a symbol by a board quantity goes through
// game.RepeatMana, so a count of 2^40 cannot allocate.
func TestUnboundedProducedFuncsStayBounded(t *testing.T) {
	huge := 1 << 40
	// One slot past the ceiling, so the engine's cut can see and announce it.
	if got := len(game.RepeatMana("{G}", huge)); got != 3*(game.MaxProducedMana+1) {
		t.Errorf("RepeatMana at 2^40 built %d bytes, want %d", got, 3*(game.MaxProducedMana+1))
	}
	if got := game.RepeatMana("{G}", -5); got != "" {
		t.Errorf("RepeatMana of a negative count = %q, want empty", got)
	}
	// The engine-side parse bounds a string it was handed anyway.
	slots, err := game.ParseProducedMana("{G999999999999999999}")
	if err != nil {
		t.Fatalf("ParseProducedMana of an out-of-range amount: %v", err)
	}
	if len(slots) != game.MaxProducedMana {
		t.Errorf("{G<huge>} parsed to %d slots, want %d", len(slots), game.MaxProducedMana)
	}
	one, err := game.ParseProducedMana(effects.OneColorOfAmount(huge))
	if err != nil {
		t.Fatalf("ParseProducedMana(OneColorOfAmount): %v", err)
	}
	if len(one) != 1 || one[0].AmountFor("G") != game.MaxProducedMana {
		t.Errorf("one-colour amount 2^40 parsed to %+v, want one slot of %d", one, game.MaxProducedMana)
	}
}

// TestPlayerCounterTotalSaturates: the player-counter writer shares the
// ceiling, so two enormous placements land at it rather than past it.
func TestPlayerCounterTotalSaturates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 2; i++ {
		if err := g.AddPlayerCounterForEffect(me.ID, "experience", 1<<62); err != nil {
			t.Fatalf("AddPlayerCounterForEffect: %v", err)
		}
	}
	got := 0
	g.ReadSnapshot(func() { got = me.Counters["experience"] })
	if got != game.MaxCount {
		t.Errorf("experience counters = %d, want the ceiling %d", got, game.MaxCount)
	}
}
