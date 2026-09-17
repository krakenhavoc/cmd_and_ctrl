package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestAttacksSeeLandwalk — ADR 0045 addendum Decision 17: the attack
// planner's blocker estimate reads landwalk from the defender's lands
// in the public view. The same 2/2 that stays home against an untapped
// 5/5 swings when it has islandwalk and the defender controls an
// Island, because nothing they have can block it — and stays home
// again when the defender's only land is not an Island, or when the
// landwalk is nonbasic and every land is basic.
func TestAttacksSeeLandwalk(t *testing.T) {
	island := func(c *protocol.CardView) { c.Name, c.TypeLine = "Island", "Basic Land — Island" }
	tropical := func(c *protocol.CardView) { c.Name, c.TypeLine = "Tropical Island", "Land — Island Forest" }
	for _, tc := range []struct {
		name    string
		keyword string
		land    cardOpt
		attacks bool
	}{
		{"islandwalk against an Island", "islandwalk", island, true},
		{"islandwalk against a Mountain", "islandwalk", func(*protocol.CardView) {}, false},
		{"nonbasic landwalk against a basic Island", "nonbasic landwalk", island, false},
		{"nonbasic landwalk against a Tropical Island", "nonbasic landwalk", tropical, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := newView(
				[]protocol.PlayerView{newSeat(0), newSeat(1)},
				withBattlefield(
					creature(cardID(10), 0, "Fish", 2, 2, keywords(tc.keyword)),
					creature(cardID(20), 1, "Giant", 5, 5),
					land(cardID(21), 1, tc.land),
				),
				withTurn(5, 0, "declare_attackers"),
			)
			in := input(0, v, passMove(0), attackMove(t, 0, cardID(10), 1))
			got := chose(t, in, decide(t, heuristic.New(), in))
			if attacked := got != "Pass priority"; attacked != tc.attacks {
				t.Errorf("chose %q, want attack = %v", got, tc.attacks)
			}
		})
	}
}
