package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// banding_test.go — ADR 0224 §6 (#5723): the bot blocks a creature in a
// band only to survive, since blocking one member blocks the band
// (CR 702.22h) and the attacking player divides the blocker's damage
// (CR 702.22k).

func banded(c protocol.CardView, band int) protocol.CardView {
	c.AttackBand = band
	return c
}

// bandView: seat 1 attacks seat 0 with a band of a 1/1 and a 5/5; seat 0
// has a 2/2 and the given life.
func bandView(life int) protocol.GameView {
	return newView(
		[]protocol.PlayerView{newSeat(0, withLife(life)), newSeat(1)},
		withBattlefield(
			banded(creature(cardID(20), 1, "Hero", 1, 1, keywords("banding"), attacking(0)), 1),
			banded(creature(cardID(21), 1, "Giant", 5, 5, attacking(0)), 1),
			creature(cardID(10), 0, "Bear", 2, 2),
		),
		withTurn(5, 1, "declare_blockers"),
	)
}

func TestBotDoesNotBlockABandAtAHealthyLife(t *testing.T) {
	v := bandView(40)
	in := input(0, v, blockMove(t, 0, cardID(10), cardID(20)), blockMove(t, 0, cardID(10), cardID(21)))
	if d := decide(t, heuristic.New(), in); d.Index != aiseat.Decline {
		t.Fatalf("chose %q (%s); at 40 life the bot blocks no band", chose(t, in, d), d.Reason)
	}
}

// At 6 life the band's 6 is lethal, and a block on its 1/1 stops all of
// it: the bot takes it, though the 1/1 alone is not worth a chump.
func TestBotBlocksABandToSurvive(t *testing.T) {
	v := bandView(6)
	in := input(0, v, blockMove(t, 0, cardID(10), cardID(20)), blockMove(t, 0, cardID(10), cardID(21)))
	d := decide(t, heuristic.New(), in)
	if d.Index == aiseat.Decline {
		t.Fatal("declined a block that stops a lethal band")
	}
}
