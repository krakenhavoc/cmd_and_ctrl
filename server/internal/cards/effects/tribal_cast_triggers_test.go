package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tribal_cast_triggers_test.go — #4177: the cast triggers of Door of
// Destinies ("a spell of the chosen type") and Vanquisher's Banner ("a
// creature spell of the chosen type"), which the older tribal tests left
// untried.

func chargeCounters(t *testing.T, g *game.Game, id interface{ String() string }) int {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID.String() == id.String() {
			return c.Counters["charge"]
		}
	}
	t.Fatalf("permanent %s is not on the battlefield", id)
	return 0
}

func castFree(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) {
	t.Helper()
	castFromHandForTest(t, g, p, name, typeLine, "", "", game.CastSpellParams{FromZone: "hand"})
	passPriorityAroundTable(t, g)
}

func TestDoorOfDestiniesGrowsOnYourSpellsOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	door := pushNamedTribePermanent(t, g, me.ID, "Door of Destinies", "Artifact", doorOfDestiniesOracle, "Sliver")

	castFree(t, g, me, "Metallic Sliver", "Artifact Creature — Sliver")
	if got := chargeCounters(t, g, door); got != 1 {
		t.Fatalf("after your Sliver spell: %d charge counters, want 1", got)
	}
	// "A spell of the chosen type", not "a creature spell": a Kindred
	// Instant naming the type grows it too.
	castFree(t, g, me, "Sliver Surge", "Kindred Instant — Sliver")
	if got := chargeCounters(t, g, door); got != 2 {
		t.Errorf("after a Kindred Sliver instant: %d charge counters, want 2", got)
	}
	castFree(t, g, me, "Goblin Test", "Creature — Goblin")
	if got := chargeCounters(t, g, door); got != 2 {
		t.Errorf("a Goblin spell grew the Door: %d charge counters", got)
	}
	castFree(t, g, opp, "Their Sliver Surge", "Kindred Instant — Sliver")
	if got := chargeCounters(t, g, door); got != 2 {
		t.Errorf("an opponent's Sliver spell grew the Door: %d charge counters", got)
	}
}

func TestVanquishersBannerDrawsOnYourCreatureSpellsOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushNamedTribePermanent(t, g, me.ID, "Vanquisher's Banner", "Artifact", vanquishersBannerOracle, "Elf")

	before := me.Hand.Size()
	castFree(t, g, me, "Llanowar Elves", "Creature — Elf Druid")
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("an Elf creature spell drew %d, want 1", got)
	}
	// "A CREATURE spell": a Kindred Instant naming Elf draws nothing.
	before = me.Hand.Size()
	castFree(t, g, me, "Elf Surge", "Kindred Instant — Elf")
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("a Kindred Elf instant drew %d, want 0", got)
	}
	castFree(t, g, me, "Goblin Test", "Creature — Goblin")
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("a Goblin spell drew %d, want 0", got)
	}
	oppBefore, myBefore := opp.Hand.Size(), me.Hand.Size()
	castFree(t, g, opp, "Their Elf Surge", "Kindred Instant — Elf")
	if opp.Hand.Size()-oppBefore != 0 || me.Hand.Size()-myBefore != 0 {
		t.Error("an opponent's Elf spell drew a card")
	}
}
