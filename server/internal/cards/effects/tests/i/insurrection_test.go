package testsi

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const insurrectionOracle = "7f7c204f-be1a-47e4-91c6-ba6f906d9012"

// TestInsurrectionStealsUntapsAndHastesEveryCreature — table-wide Act
// of Treason. Gaining control of a creature you already control is a
// harmless no-op; the assertion that matters is the OPPONENTS'
// creatures changing hands.
func TestInsurrectionStealsUntapsAndHastesEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs1 := effectstest.PushTappedVanillaCreature(g, opp1.ID, "Their Bear", 2, 2)
	theirs2 := effectstest.PushTappedVanillaCreature(g, opp2.ID, "Other Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Insurrection", "Sorcery", insurrectionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, theirs1); got != me.ID {
		t.Errorf("opponent 1's creature controller %s, want %s", got, me.ID)
	}
	if got := effectstest.ControllerOf(t, g, theirs2); got != me.ID {
		t.Errorf("opponent 2's creature controller %s, want %s", got, me.ID)
	}
	if got := effectstest.ControllerOf(t, g, mine); got != me.ID {
		t.Errorf("my own creature controller %s, want %s (unchanged)", got, me.ID)
	}
	for _, id := range []uuid.UUID{theirs1, theirs2} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
			t.Errorf("Insurrection did not untap %s", c.Name)
		}
	}
	effectstest.AssertKeywords(t, g, theirs1, "haste")
	effectstest.AssertKeywords(t, g, theirs2, "haste")
}

// The theft reverts at cleanup, same as Act of Treason.
func TestInsurrectionGivesCreaturesBackAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Insurrection", "Sorcery", insurrectionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Fatalf("setup: the theft did not happen")
	}

	effectstest.AdvancePastCleanupForTest(t, g)

	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Errorf("after cleanup: controller %s, want the owner %s", got, opp.ID)
	}
}
