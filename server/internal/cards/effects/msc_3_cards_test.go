package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_3_cards_test.go — Marvel Super Heroes Commander, third batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	msc3AntManOracle     = "1e4cf4a4-50a4-4856-bf7e-20899e8affd1"
	msc3BoomerangOracle  = "bbd84035-3efd-4301-9778-913898efd57a"
	msc3SkirmisherOracle = "8f1c21d6-bf85-4e31-a978-6423c40650d2"
	msc3EvolutionOracle  = "6401d0ae-4635-497d-91ef-ce2eabefb893"
)

func TestMsc3EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc3AntManOracle:     "Ant-Man, Scott Lang",
		msc3BoomerangOracle:  "Boomerang, Blade Flinger",
		msc3SkirmisherOracle: "Atlantean Skirmisher",
		msc3EvolutionOracle:  "Accelerated Evolution",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestMsc3AntManPutsACounterOnItselfForFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ant := b12Push(g, me.ID, "Ant-Man, Scott Lang", "Legendary Creature — Human Hero", msc3AntManOracle, 2, 2)
	b06AddMana(me, "C", "C", "C", "C")
	b16Activate(t, g, me.ID, ant, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ant, game.CounterPlusOne); got != 1 {
		t.Fatalf("{4} puts one +1/+1 counter on Ant-Man: %d", got)
	}
	if got := effectivePowerWithCounters(t, g, ant); got != 3 {
		t.Errorf("a 2/2 with a counter is a 3/3: power %d", got)
	}
}

func TestMsc3BoomerangPingsEachOpponentAndGainsALifeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	// Power 0 so that the only damage in the test is the trigger's.
	boomerang := b12Push(g, me.ID, "Boomerang, Blade Flinger", "Legendary Creature — Human Villain", msc3BoomerangOracle, 0, 3)
	life, oppLife := me.Life, opp.Life
	declareAttack(t, g, opp.ID, boomerang)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("opponent: %d → %d, want one damage from the trigger", oppLife, opp.Life)
	}
	if me.Life != life+1 {
		t.Errorf("you gain 1 life: %d → %d", life, me.Life)
	}
}

func TestMsc3AtlanteanSkirmisherConnivesWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	skirmisher := pushCatalogPermanent(g, me.ID, "Atlantean Skirmisher", "Creature — Merfolk Rogue", msc3SkirmisherOracle, false)
	seedLibrary(me, "One", "Two", "Three")
	lib := me.Library.Size()
	declareAttack(t, g, opp.ID, skirmisher)
	passPriorityAroundTable(t, g)
	if got := lib - me.Library.Size(); got != 1 {
		t.Errorf("connive draws one card: library fell by %d", got)
	}
}

func TestMsc3AcceleratedEvolutionGivesPlusTwoAndHexproofOnEnter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "G", "C", "C")
	castCatalogSpell(t, g, "Accelerated Evolution", "Enchantment — Aura", msc3EvolutionOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("enchanted Bear is %d/%d, want 4/4", p, tg)
	}
	assertKeywords(t, g, bear, "hexproof")
}
