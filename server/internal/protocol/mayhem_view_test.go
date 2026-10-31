package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mayhem_view_test.go — #4628, CR 702.187. The client's cast button on a
// graveyard card comes off `castable_here` and the card's
// `alternative_costs`, which the view derives from the same predicates
// the engine and the enumerator read. A mayhem card shows the button and
// the mayhem price exactly while it was discarded this turn.

const viewSadisticSlashOracle = "0e471b6d-4ba4-46aa-a802-81e401cbfa56"

func TestCastableHereAndTheMayhemPriceFollowTheDiscard(t *testing.T) {
	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me := g.Seats[0]

	milled := graveyardCard(me, "Sadistic Slash", viewSadisticSlashOracle)

	// The one that really goes through a discard.
	c := game.NewCard("Sadistic Slash", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{3}{B}"
	c.OracleID = viewSadisticSlashOracle
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Hand.PushTop(c)
		if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
			t.Errorf("discard: %v", err)
		}
	})
	if !me.Graveyard.Contains(c.InstanceID) {
		t.Fatal("fixture: the Slash was not discarded")
	}

	v := ViewOfGameFor(g, me.ID.String())
	got := graveyardCardView(t, v, 0, c.InstanceID)
	if !got.CastableHere {
		t.Error("no castable_here on a card discarded this turn")
	}
	if len(got.AlternativeCosts) != 1 || got.AlternativeCosts[0].Key != "mayhem" || got.AlternativeCosts[0].ManaCost != "{1}{B}" {
		t.Errorf("alternative_costs = %+v, want mayhem {1}{B}", got.AlternativeCosts)
	}
	if !got.AlternativeCostRequired {
		t.Error("the printed cost is offered from the graveyard")
	}

	never := graveyardCardView(t, v, 0, milled)
	if never.CastableHere || len(never.AlternativeCosts) != 0 {
		t.Errorf("a card that was never discarded: castable_here=%v alternative_costs=%+v", never.CastableHere, never.AlternativeCosts)
	}
}
