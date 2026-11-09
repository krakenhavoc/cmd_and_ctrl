package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b6_test.go — Planechase plane cards, batch 6 (ADR 0216 slice 4,
// #4870): each plane is played face up on a Planechase table
// (planes_harness_test.go) and its static, triggered and chaos abilities are
// driven by the engine.

var (
	planeGreatAerie = planeFixture{"00f00001-bd84-4dbf-a707-f4b1549100d4", "The Great Aerie", "Plane — Tarkir"}
)

func TestGreatAerieBolstersOnEntryAndEachUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, planeGreatAerie)
	g, _ := newPlaneGame(t, planeGreatAerie)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	small := pushVanillaCreature(g, me.ID, "Small", 1, 1)
	big := pushVanillaCreature(g, me.ID, "Big", 4, 4)
	planeswalkToPlane(t, g, planeGreatAerie.Oracle)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, small, game.CounterPlusOne); got != 3 {
		t.Errorf("least-toughness creature has %d counters, want 3", got)
	}
	if got := countersOn(g, big, game.CounterPlusOne); got != 0 {
		t.Errorf("the big creature got %d counters", got)
	}
}

func TestGreatAerieChaosDealsToughnessBothWays(t *testing.T) {
	g, _ := newPlaneGame(t, planeGreatAerie)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mine := pushVanillaCreature(g, me.ID, "Wall", 0, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Ogre", 5, 2)
	rollBlank(t, g, me.ID)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, mine)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	// Wall deals 3 to Ogre (toughness 2: dies); Ogre deals 2 to Wall (survives).
	g.WithWriteLock(func() {
		if _, ok := g.LookupCardForEffect(theirs); ok && g.FindCardZoneForEffect(theirs).Kind == game.ZoneBattlefield {
			t.Error("the opposing creature survived 3 damage")
		}
		if z := g.FindCardZoneForEffect(mine); z == nil || z.Kind != game.ZoneBattlefield {
			t.Error("the wall died to 2 damage")
		}
	})
}
