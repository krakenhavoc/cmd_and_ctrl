package effects

import (
	"testing"

	"github.com/google/uuid"
)

const insurrectionOracle = "7f7c204f-be1a-47e4-91c6-ba6f906d9012"

// TestInsurrectionStealsUntapsAndHastesEveryCreature — table-wide Act
// of Treason. Gaining control of a creature you already control is a
// harmless no-op; the assertion that matters is the OPPONENTS'
// creatures changing hands.
func TestInsurrectionStealsUntapsAndHastesEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs1 := pushTappedVanillaCreature(g, opp1.ID, "Their Bear", 2, 2)
	theirs2 := pushTappedVanillaCreature(g, opp2.ID, "Other Bear", 2, 2)

	castCatalogSpell(t, g, "Insurrection", "Sorcery", insurrectionOracle, nil)
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, theirs1); got != me.ID {
		t.Errorf("opponent 1's creature controller %s, want %s", got, me.ID)
	}
	if got := controllerOf(t, g, theirs2); got != me.ID {
		t.Errorf("opponent 2's creature controller %s, want %s", got, me.ID)
	}
	if got := controllerOf(t, g, mine); got != me.ID {
		t.Errorf("my own creature controller %s, want %s (unchanged)", got, me.ID)
	}
	for _, id := range []uuid.UUID{theirs1, theirs2} {
		if c, _ := battlefieldCard(g, id); c.Tapped {
			t.Errorf("Insurrection did not untap %s", c.Name)
		}
	}
	assertKeywords(t, g, theirs1, "haste")
	assertKeywords(t, g, theirs2, "haste")
}

// The theft reverts at cleanup, same as Act of Treason.
func TestInsurrectionGivesCreaturesBackAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	castCatalogSpell(t, g, "Insurrection", "Sorcery", insurrectionOracle, nil)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("setup: the theft did not happen")
	}

	advancePastCleanupForTest(t, g)

	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("after cleanup: controller %s, want the owner %s", got, opp.ID)
	}
}
