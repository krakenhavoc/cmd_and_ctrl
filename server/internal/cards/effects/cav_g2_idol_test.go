package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Idol of False Gods keeps being a creature when it loses all abilities
// (2024-06-07 ruling): the type change is a layer 4 effect and the removal
// is layer 6, which reaches only forwards (CR 613.6, ADR 0067 §2). What it
// loses is the layer 6 annihilator, which an effect started after the
// removal's timestamp does not outlive.
func TestIdolOfFalseGodsStaysACreatureWhenItLosesItsAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	idol := pushCatalogPermanent(g, me.ID, "Idol of False Gods", "Kindred Artifact — Eldrazi", idolOfFalseGodsOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(idol, game.CounterPlusOne, 8); err != nil {
			t.Fatalf("add counters: %v", err)
		}
	})
	before := annCard(t, g, idol)
	if !before.IsCreature() || len(game.AnnihilatorAmounts(&before)) != 1 {
		t.Fatalf("setup: creature %v, annihilator %v", before.IsCreature(), game.AnnihilatorAmounts(&before))
	}

	castCatalogSpell(t, g, "Sudden Spoiling", "Instant", suddenSpoilingOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)

	after := annCard(t, g, idol)
	if !after.IsCreature() {
		t.Error("the Idol stopped being a creature when it lost its abilities")
	}
	if got := game.AnnihilatorAmounts(&after); len(got) != 0 {
		t.Errorf("annihilator %v after losing all abilities, want none", got)
	}
}

// When the removal is OLDER than the Idol (a Sludge Monster out first), the
// Idol's layer 7b half still applies after it (CR 613.6): its 0/0 base beats
// the Sludge Monster's older base 2/2, so eight counters make an 8/8 and not
// a 10/10. It stays a creature.
func TestIdolOfFalseGodsBasePTOutlivesAnOlderRemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sludge Monster", "Creature — Horror", "2802fc92-49a0-43c8-bc10-24882eeee3f1", false)
	idol := pushCatalogPermanent(g, me.ID, "Idol of False Gods", "Kindred Artifact — Eldrazi", idolOfFalseGodsOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(idol, game.CounterPlusOne, 8); err != nil {
			t.Fatalf("add counters: %v", err)
		}
		if err := g.AddCounterForEffect(idol, "slime", 1); err != nil {
			t.Fatalf("add slime counter: %v", err)
		}
	})
	c := annCard(t, g, idol)
	if !c.IsCreature() {
		t.Error("the Idol stopped being a creature")
	}
	if c.CurrentPower() != 8 || c.CurrentToughness() != 8 {
		t.Errorf("Idol is %d/%d, want 8/8 (base 0/0 outlives the older removal)", c.CurrentPower(), c.CurrentToughness())
	}
}
