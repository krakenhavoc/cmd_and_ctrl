package testsr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const reinsOfPowerOracle = "96381388-1e82-411c-8290-1f3e909d3b5f"

// TestReinsOfPowerSwapsControlBothWaysAndHastesEveryone pins the
// two-way swap: the caster's creatures go to the chosen opponent and
// vice versa, both sides untap, and everything that changed hands
// gains haste until end of turn. A third player's creature is
// untouched.
func TestReinsOfPowerSwapsControlBothWaysAndHastesEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.PushTappedVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushTappedVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	bystander := effectstest.PushVanillaCreature(g, other.ID, "Bystander Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Reins of Power", "Instant", reinsOfPowerOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, mine); got != opp.ID {
		t.Errorf("my creature's controller %s, want the opponent %s", got, opp.ID)
	}
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Errorf("their creature's controller %s, want me %s", got, me.ID)
	}
	if got := effectstest.ControllerOf(t, g, bystander); got != other.ID {
		t.Errorf("a third player's creature changed hands: controller %s, want %s", got, other.ID)
	}
	if c, _ := effectstest.BattlefieldCard(g, mine); c.Tapped {
		t.Error("Reins of Power did not untap my creature")
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Tapped {
		t.Error("Reins of Power did not untap the opponent's creature")
	}
	effectstest.AssertKeywords(t, g, mine, "haste")
	effectstest.AssertKeywords(t, g, theirs, "haste")
}

// The swap reverts at cleanup, same as every other until-end-of-turn
// control change.
func TestReinsOfPowerGivesCreaturesBackAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Reins of Power", "Instant", reinsOfPowerOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, mine); got == me.ID {
		t.Fatalf("setup: the swap did not happen")
	}

	effectstest.AdvancePastCleanupForTest(t, g)

	if got := effectstest.ControllerOf(t, g, mine); got != me.ID {
		t.Errorf("after cleanup: my creature's controller %s, want me %s", got, me.ID)
	}
	if got := effectstest.ControllerOf(t, g, theirs); got != opp.ID {
		t.Errorf("after cleanup: their creature's controller %s, want them %s", got, opp.ID)
	}
}
