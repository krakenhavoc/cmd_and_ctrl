package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const huddleUpOracle = "81776a16-bf38-450b-995b-beb29121cab2"

// TestHuddleUpMakesEachTargetedPlayerDrawOne: both picks draw one card,
// and a third player who was not picked draws nothing.
func TestHuddleUpMakesEachTargetedPlayerDrawOne(t *testing.T) {
	assistCardSpec(t, huddleUpOracle, "Huddle Up")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]

	castCatalogSpell(t, g, "Huddle Up", "Sorcery", huddleUpOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID},
		{Kind: game.TargetPlayer, ID: opp.ID},
	})
	meBefore, oppBefore, byBefore := me.Hand.Size(), opp.Hand.Size(), bystander.Hand.Size()
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != meBefore+1 || opp.Hand.Size() != oppBefore+1 {
		t.Errorf("hands went %d->%d and %d->%d, want each up by one", meBefore, me.Hand.Size(), oppBefore, opp.Hand.Size())
	}
	if bystander.Hand.Size() != byBefore {
		t.Errorf("an unpicked player's hand went %d->%d, want no draw", byBefore, bystander.Hand.Size())
	}
}
