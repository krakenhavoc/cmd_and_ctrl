package effects

import (
	"testing"

	"github.com/google/uuid"

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

// #3862: a removal OLDER than the Idol's effect (a Sludge Monster out
// first) does not take the annihilator either. The layer 6 half is the
// same effect as the layer 4 one, so CR 613.6 carries it past the
// removal in the removal's own layer, and timestamp order puts the
// older removal first.
func TestIdolOfFalseGodsKeepsAnnihilatorUnderAnOlderRemoval(t *testing.T) {
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
	if !c.HasLostAllAbilities() {
		t.Fatal("setup: the Sludge Monster did not take the Idol's abilities")
	}
	if got := game.AnnihilatorAmounts(&c); len(got) != 1 || got[0] != 2 {
		t.Errorf("annihilator %v under an older removal, want [2]", got)
	}
}

// The Idol's ruling: "If Idol of False Gods loses its abilities before
// gaining its eighth +1/+1 counter, it will have annihilator 2;
// otherwise, it won't." The effect begins at the eighth counter, so a
// removal that arrived at seven is older than it.
func TestIdolOfFalseGodsThatLostItsAbilitiesBeforeTheEighthCounterHasAnnihilator(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	idol := pushCatalogPermanent(g, me.ID, "Idol of False Gods", "Kindred Artifact — Eldrazi", idolOfFalseGodsOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(idol, game.CounterPlusOne, 7); err != nil {
			t.Fatalf("add counters: %v", err)
		}
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(idol),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
	})
	if c := annCard(t, g, idol); c.IsCreature() || !c.HasLostAllAbilities() {
		t.Fatalf("setup at seven counters: creature %v, lost abilities %v", c.IsCreature(), c.HasLostAllAbilities())
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(idol, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("add the eighth counter: %v", err)
		}
	})
	c := annCard(t, g, idol)
	if !c.IsCreature() || c.CurrentPower() != 8 || c.CurrentToughness() != 8 {
		t.Errorf("at eight counters: creature %v, %d/%d; want an 8/8 creature", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	if got := game.AnnihilatorAmounts(&c); len(got) != 1 || got[0] != 2 {
		t.Errorf("annihilator %v, want [2]: it lost its abilities before its eighth counter", got)
	}
}

// "Otherwise, it won't": a removal after the eighth counter takes the
// annihilator, and a ninth counter does not start the effect again.
func TestIdolOfFalseGodsANinthCounterDoesNotOutrunALaterRemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	idol := pushCatalogPermanent(g, me.ID, "Idol of False Gods", "Kindred Artifact — Eldrazi", idolOfFalseGodsOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(idol, game.CounterPlusOne, 8); err != nil {
			t.Fatalf("add counters: %v", err)
		}
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(idol),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
		if err := g.AddCounterForEffect(idol, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("add the ninth counter: %v", err)
		}
	})
	c := annCard(t, g, idol)
	if !c.IsCreature() || c.CurrentPower() != 9 {
		t.Errorf("creature %v, power %d; want a 9/9 creature", c.IsCreature(), c.CurrentPower())
	}
	if got := game.AnnihilatorAmounts(&c); len(got) != 0 {
		t.Errorf("annihilator %v, want none: it lost its abilities after its eighth counter", got)
	}
}
