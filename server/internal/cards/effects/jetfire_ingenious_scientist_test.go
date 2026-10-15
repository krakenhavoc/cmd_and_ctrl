package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func jetfireRow() cards.Card {
	return transformerRow(jetfireOracleID,
		disturbFace{"Jetfire, Ingenious Scientist", "Legendary Artifact Creature — Robot", "{4}{U}", "3", "4"},
		disturbFace{"Jetfire, Air Guardian", "Legendary Artifact — Vehicle", "", "3", "4"},
		"More Than Meets the Eye {3}{U}\nFlying\nRemove one or more +1/+1 counters from among artifacts you control: Target player adds that much {C}. This mana can't be spent to cast nonartifact spells. Convert Jetfire.",
		"Living metal\nFlying\n{U}{U}{U}: Convert Jetfire, then adapt 3.",
		"Living metal", "Convert", "Flying", "Adapt")
}

func jetfirePoolC(p *game.Player) int {
	n := 0
	for _, tok := range p.ManaPool.Units() {
		if tok.Color == "C" {
			n++
		}
	}
	return n
}

// The cost is a total AND a split: two counters off Jetfire, one off
// another artifact, and the player adds {C}{C}{C}, restricted.
func TestJetfireRemovesACountedSplitAndAddsThatMuchRestrictedMana(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, jetfireRow(), me)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	pfAddCounters(g, id, "+1/+1", 3)
	pfAddCounters(g, rock, "+1/+1", 2)
	g.WithWriteLock(func() { me.ManaPool = nil })

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{id, rock}, CounterCounts: []int{2, 1},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})

	wantFace(t, g, id, 1, "Jetfire, Air Guardian")
	if c := findBattlefieldCardForTest(g, id); c.Counters["+1/+1"] != 1 {
		t.Errorf("Jetfire has %d counters, want 1 (3 minus 2)", c.Counters["+1/+1"])
	}
	if c := findBattlefieldCardForTest(g, rock); c.Counters["+1/+1"] != 1 {
		t.Errorf("Rock has %d counters, want 1 (2 minus 1)", c.Counters["+1/+1"])
	}
	if n := jetfirePoolC(me); n != 3 {
		t.Fatalf("%d {C} floating, want 3", n)
	}
	for _, tok := range me.ManaPool.Units() {
		if len(tok.Restrictions) != 1 || tok.Restrictions[0] != game.ManaRestrictNotNonartifactSpell {
			t.Errorf("mana restrictions = %v, want only nonartifact-spell", tok.Restrictions)
		}
	}
}

// "Target player adds": the mana goes to the chosen player, not the
// activator.
func TestJetfireManaGoesToTheTargetPlayer(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, jetfireRow(), me)
	pfAddCounters(g, id, "+1/+1", 2)
	g.WithWriteLock(func() { me.ManaPool = nil; opp.ManaPool = nil })

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{id}, CounterCounts: []int{2},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})

	if n := jetfirePoolC(opp); n != 2 {
		t.Errorf("target player has %d {C}, want 2", n)
	}
	if n := jetfirePoolC(me); n != 0 {
		t.Errorf("activator has %d {C}, want 0", n)
	}
}

// The whole payment is validated as a set before anything comes off.
func TestJetfireRefusesBadPayments(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, jetfireRow(), me)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	for _, x := range []uuid.UUID{id, rock, bear, theirs} {
		pfAddCounters(g, x, "+1/+1", 2)
	}
	tgt := []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}

	for name, p := range map[string]game.ActivateAbilityParams{
		"none removed":        {CounterSourceIDs: []uuid.UUID{id}, CounterCounts: []int{0}, Targets: tgt},
		"no payment":          {Targets: tgt},
		"more than it holds":  {CounterSourceIDs: []uuid.UUID{id}, CounterCounts: []int{3}, Targets: tgt},
		"a non-artifact":      {CounterSourceIDs: []uuid.UUID{bear}, CounterCounts: []int{1}, Targets: tgt},
		"an opponent's":       {CounterSourceIDs: []uuid.UUID{theirs}, CounterCounts: []int{1}, Targets: tgt},
		"one permanent twice": {CounterSourceIDs: []uuid.UUID{id, id}, CounterCounts: []int{1, 1}, Targets: tgt},
		"counts mismatch":     {CounterSourceIDs: []uuid.UUID{id, rock}, CounterCounts: []int{1}, Targets: tgt},
	} {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, p); err == nil {
			t.Errorf("%s: activation was accepted", name)
		}
	}
	for _, x := range []uuid.UUID{id, rock, bear, theirs} {
		if c := findBattlefieldCardForTest(g, x); c.Counters["+1/+1"] != 2 {
			t.Errorf("a refused payment still removed counters (%s has %d)", c.Name, c.Counters["+1/+1"])
		}
	}
	wantFace(t, g, id, 0, "Jetfire, Ingenious Scientist")
}

// Back face: convert, then adapt 3 puts counters on only if it has none.
func TestJetfireAirGuardianConvertsThenAdapts(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerConverted(t, g, jetfireRow(), me)
	for range 3 {
		b06AddMana(me, "U")
	}

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})

	c := wantFace(t, g, id, 0, "Jetfire, Ingenious Scientist")
	if c.Counters["+1/+1"] != 3 {
		t.Errorf("%d counters, want 3 from adapt", c.Counters["+1/+1"])
	}
}

func TestJetfireAirGuardianAdaptDoesNothingWithCounters(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerConverted(t, g, jetfireRow(), me)
	pfAddCounters(g, id, "+1/+1", 1)
	for range 3 {
		b06AddMana(me, "U")
	}

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})

	c := wantFace(t, g, id, 0, "Jetfire, Ingenious Scientist")
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("%d counters, want the 1 it had (adapt does nothing)", c.Counters["+1/+1"])
	}
}
