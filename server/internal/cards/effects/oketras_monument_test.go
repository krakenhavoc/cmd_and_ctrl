package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oketrasMonumentOracle = "0370afa0-07d3-4787-8a5b-10272cb3a486"

// TestOketrasMonumentDiscountsYourWhiteCreatureSpellsOnly — the
// discount is "WHITE CREATURE spells YOU cast": all three words are
// predicates, and each one is checked by a spell that fails only it.
func TestOketrasMonumentDiscountsYourWhiteCreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Oketra's Monument", "Legendary Artifact", oketrasMonumentOracle, false)

	if got := priceInHand(t, g, me, "White Bear", "Creature — Bear", "{1}{W}"); got != 1 {
		t.Errorf("own white creature spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Green Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("own green creature spell: %d, want 2 (untouched)", got)
	}
	if got := priceInHand(t, g, me, "White Instant", "Instant", "{1}{W}"); got != 2 {
		t.Errorf("own white noncreature spell: %d, want 2 (untouched)", got)
	}
	if got := priceInHand(t, g, them, "Their White Bear", "Creature — Bear", "{1}{W}"); got != 2 {
		t.Errorf("opponent's white creature spell: %d, want 2 (undiscounted)", got)
	}
}

// TestOketrasMonumentMakesAWarriorForAnyCreatureSpellYouCast — the
// trigger is NOT colour-restricted: any creature spell you cast makes
// a 1/1 white Warrior with vigilance. An opponent's creature spell
// makes nothing.
func TestOketrasMonumentMakesAWarriorForAnyCreatureSpellYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Oketra's Monument", "Legendary Artifact", oketrasMonumentOracle, false)

	castCatalogSpell(t, g, "Green Bear", "Creature — Bear", "test-green-bear", nil)
	passPriorityAroundTable(t, g)

	if got := countTokensControlled(g, me.ID, "Warrior"); got != 1 {
		t.Fatalf("Warrior tokens after one creature spell = %d, want 1", got)
	}
	var warrior game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Warrior" && c.Controller == me.ID {
			warrior = c
		}
	}
	if warrior.Power != 1 || warrior.Toughness != 1 || !warrior.HasColor("W") {
		t.Errorf("Warrior = %d/%d colours %v, want a 1/1 white", warrior.Power, warrior.Toughness, warrior.Colors)
	}
	if !hasEffectiveKeyword(t, g, warrior.InstanceID, "vigilance") {
		t.Error("the Warrior token has no vigilance")
	}

	// A noncreature spell does not trigger it.
	castCatalogSpell(t, g, "Some Instant", "Instant", "test-some-instant", nil)
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Warrior"); got != 1 {
		t.Errorf("Warrior tokens after a noncreature spell = %d, want still 1", got)
	}

	// Nor does an opponent's creature spell.
	advanceToPrecombatMainOf(t, g, 1)
	castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "test-their-bear", nil)
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Warrior"); got != 1 {
		t.Errorf("Warrior tokens after an opponent's creature spell = %d, want still 1", got)
	}
	if got := countTokensControlled(g, them.ID, "Warrior"); got != 0 {
		t.Errorf("the opponent got %d Warriors from someone else's Monument", got)
	}
}
