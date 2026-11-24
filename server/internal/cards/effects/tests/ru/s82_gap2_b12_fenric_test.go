package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b12_fenric_test.go — S82 GAP2 batch 12: The Curse of Fenric,
// driven through the real cast, chapter, destroy, layer and fight paths.

const gap2B12Fenric = "ef7fce1b-0d95-4f92-a8ad-09d9756fa0e4"

func TestGap2B12FenricIsFull(t *testing.T) {
	spec, ok := effects.Lookup(gap2B12Fenric)
	if !ok {
		t.Fatal("The Curse of Fenric is not registered")
	}
	if spec.Name != "The Curse of Fenric" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func gap2B12Count(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name {
			n++
		}
	}
	return n
}

func TestGap2B12FenricChaptersDestroyRenameAndFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp1.ID, "Their Bear", 2, 2)
	tough := effectstest.PushTestPermanent(g, opp2.ID, game.Card{Name: "Stone Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"indestructible"}})
	victim := effectstest.PushVanillaCreature(g, opp1.ID, "Big Ogre", 3, 3)

	effectstest.CastCatalogSpell(t, g, "The Curse of Fenric", "Enchantment — Saga", gap2B12Fenric, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no target prompt for chapter I: %+v", g.PendingChoices)
	}
	if effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Error("an opponent's creature was offered on your own clause")
	}
	gap2B12AnswerPickTargets(t, g, me.ID, mine, theirs, tough)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("chapter I did not destroy the picks")
	}
	if !g.Battlefield.Contains(tough) {
		t.Fatal("the indestructible creature was destroyed")
	}
	if n := gap2B12Count(g, me.ID, "Mutant"); n != 1 {
		t.Errorf("I have %d Mutants, want 1", n)
	}
	if n := gap2B12Count(g, opp1.ID, "Mutant"); n != 1 {
		t.Errorf("the first opponent has %d Mutants, want 1", n)
	}
	if n := gap2B12Count(g, opp2.ID, "Mutant"); n != 0 {
		t.Errorf("the second opponent has %d Mutants, want 0: their creature was not destroyed", n)
	}

	// Chapter II: the Ogre becomes Fenric. A Mutant token is not a legal pick.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no target prompt for chapter II: %+v", g.PendingChoices)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Mutant" && effectstest.HasID(prompt.PickTargetCards, c.InstanceID) {
			t.Error("a token was offered to chapter II")
		}
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	ogre, ok := g.LookupCardForEffect(victim)
	if !ok {
		t.Fatal("the Ogre is gone")
	}
	eff := ogre.Effective()
	if eff.Name != "Fenric" || eff.Power != 6 || eff.Toughness != 6 {
		t.Errorf("the Ogre is %q %d/%d, want Fenric 6/6", eff.Name, eff.Power, eff.Toughness)
	}
	if !ogre.HasSubtype("Horror") || ogre.HasSubtype("Ogre") || !ogre.HasSupertype("Legendary") {
		t.Errorf("the Ogre's types: %+v %+v %+v", eff.Supertypes, eff.Types, eff.Subtypes)
	}

	// Chapter III: my Mutant fights Fenric; deathtouch kills it.
	var mutant uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Mutant" && c.Controller == me.ID {
			mutant = c.InstanceID
		}
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no target prompt for chapter III: %+v", g.PendingChoices)
	}
	effectstest.PickCard(t, g, me.ID, mutant)
	prompt = effectstest.LatestPickTarget(g, me.ID)
	if prompt != nil && effectstest.HasID(prompt.PickTargetCards, tough) {
		t.Error("a creature not named Fenric was offered as the second target")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("Fenric survived a fight with a deathtouch Mutant")
	}
	if g.Battlefield.Contains(mutant) {
		t.Error("the Mutant survived 6 damage")
	}
	// Fenric's change outlives the Saga, which is sacrificed after III.
	for _, c := range g.Battlefield.Cards {
		if c.Name == "The Curse of Fenric" {
			t.Error("the Saga was not sacrificed after chapter III")
		}
	}
}
