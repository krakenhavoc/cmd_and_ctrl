package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ichTekikOracle = "fcbada94-54fe-4330-8303-c46162420228"

func ichTekikEnter(t *testing.T, g *game.Game, owner *game.Player) uuid.UUID {
	t.Helper()
	c := game.NewCard("Ich-Tekik, Salvage Splicer", owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = "Legendary Creature — Phyrexian Human Artificer", ichTekikOracle, 1, 1
	owner.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter Ich-Tekik: %v", err)
	}
	return c.InstanceID
}

func TestIchTekikMakesAGolemWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ichTekikEnter(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Phyrexian Golem"); got != 1 {
		t.Fatalf("Golems = %d, want 1", got)
	}
}

func TestIchTekikGrowsWithEachGolemWhenAnArtifactIsPutIntoAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ich := ichTekikEnter(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	mine := effectstest.B21Push(g, me.ID, "My Golem", "Artifact Creature — Golem", "", 2, 2)
	theirs := effectstest.B21Push(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "", 2, 2)
	bear := effectstest.B21Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	rock := effectstest.PushArtifactForTest(g, opp.ID, "Their Rock")

	effectstest.KillCreature(t, g, opp.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)

	want := map[string]int{"ich": 1, "mine": 1, "theirs": 0, "bear": 0}
	for name, id := range map[string]uuid.UUID{"ich": ich, "mine": mine, "theirs": theirs, "bear": bear} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != want[name] {
			t.Errorf("%s counters = %d, want %d", name, got, want[name])
		}
	}
	golem := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Phyrexian Golem" && c.Controller == me.ID {
				golem = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if golem != 1 {
		t.Errorf("the Golem token has %d counters, want 1", golem)
	}
}

func TestIchTekikIgnoresANonArtifactDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ich := ichTekikEnter(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B21Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ich, game.CounterPlusOne); got != 0 {
		t.Errorf("Ich-Tekik counters = %d, want 0", got)
	}
}
