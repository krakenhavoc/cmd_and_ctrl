package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const reinsOfPowerOracle = "96381388-1e82-411c-8290-1f3e909d3b5f"

// TestReinsOfPowerSwapsControlBothWaysAndHastesEveryone pins the
// two-way swap: the caster's creatures go to the chosen opponent and
// vice versa, both sides untap, and everything that changed hands
// gains haste until end of turn. A third player's creature is
// untouched.
func TestReinsOfPowerSwapsControlBothWaysAndHastesEveryone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := pushTappedVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushTappedVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	bystander := pushVanillaCreature(g, other.ID, "Bystander Bear", 2, 2)

	castCatalogSpell(t, g, "Reins of Power", "Instant", reinsOfPowerOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, mine); got != opp.ID {
		t.Errorf("my creature's controller %s, want the opponent %s", got, opp.ID)
	}
	if got := controllerOf(t, g, theirs); got != me.ID {
		t.Errorf("their creature's controller %s, want me %s", got, me.ID)
	}
	if got := controllerOf(t, g, bystander); got != other.ID {
		t.Errorf("a third player's creature changed hands: controller %s, want %s", got, other.ID)
	}
	if c, _ := battlefieldCard(g, mine); c.Tapped {
		t.Error("Reins of Power did not untap my creature")
	}
	if c, _ := battlefieldCard(g, theirs); c.Tapped {
		t.Error("Reins of Power did not untap the opponent's creature")
	}
	assertKeywords(t, g, mine, "haste")
	assertKeywords(t, g, theirs, "haste")
}

// The swap reverts at cleanup, same as every other until-end-of-turn
// control change.
func TestReinsOfPowerGivesCreaturesBackAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	castCatalogSpell(t, g, "Reins of Power", "Instant", reinsOfPowerOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, mine); got == me.ID {
		t.Fatalf("setup: the swap did not happen")
	}

	advancePastCleanupForTest(t, g)

	if got := controllerOf(t, g, mine); got != me.ID {
		t.Errorf("after cleanup: my creature's controller %s, want me %s", got, me.ID)
	}
	if got := controllerOf(t, g, theirs); got != opp.ID {
		t.Errorf("after cleanup: their creature's controller %s, want them %s", got, opp.ID)
	}
}
