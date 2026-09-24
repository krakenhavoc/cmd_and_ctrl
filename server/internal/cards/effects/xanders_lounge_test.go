package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const xandersLoungeOracle = "8291543f-d086-48aa-b2b7-5481ca8c9198"

func TestXandersLoungeEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Xander's Lounge", xandersLoungeOracle)
	top100AssertEnteredTapped(t, g, id, "Xander's Lounge")
}

func TestXandersLoungeTapsForBlueBlackOrRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Xander's Lounge", "Land", xandersLoungeOracle)

	for _, want := range []string{"U", "B", "R"} {
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := riderLatestManaPick(g, me.ID)
		if pick == nil || len(pick.ColorOptions) != 3 {
			t.Fatalf("colour options %+v, want U, B and R", pick)
		}
		riderAnswerManaPicks(t, g, me.ID, want)
		if got := batch01PoolColors(me); len(got) != 1 || got[0] != want {
			t.Errorf("pool %v, want [%s]", got, want)
		}
		me.ManaPool.EmptyPool()
		b08Untap(g, land)
	}
}

func TestXandersLoungeCyclesForACard(t *testing.T) {
	g := newCatalogGame(t)
	id, me := cycleFromHand(t, g, "Xander's Lounge", "Land — Island Swamp Mountain", xandersLoungeOracle, "{C}{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Triome is not in the graveyard")
	}
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, got)
	}
}
