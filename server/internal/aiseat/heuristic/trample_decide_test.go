package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// trample_decide_test.go pins #1504 at the decision: the attack
// planner's damage estimates count what a blocked trampler puts over
// its blockers. Before it, a Wurm with a Bear in front of it read as a
// Wurm that did nothing, so a one-Wurm edge was never cashed where a
// one-Drake edge was.

// trampleBoard is seat 0 (the bot, to attack) with the given creatures
// against seat 1 at theirLife; the bot is at 20, out of reach.
func trampleBoard(theirLife int, cards ...protocol.CardView) (protocol.GameView, []string) {
	var mine []string
	for _, c := range cards {
		if c.Controller == seatID(0).String() {
			mine = append(mine, c.InstanceID)
		}
	}
	return newView(
		[]protocol.PlayerView{newSeat(0, withLife(20)), newSeat(1, withLife(theirLife))},
		withBattlefield(cards...),
		withTurn(20, 0, "declare_attackers"),
	), mine
}

// TestLethalPushCountsTrampleOverAChump: one 7/7 trampler against a
// defender at 5 whose only creature is a Bear. The Bear in front of the
// Wurm absorbs 2 and the other 5 go over: the all-in is lethal.
func TestLethalPushCountsTrampleOverAChump(t *testing.T) {
	v, mine := trampleBoard(5,
		creature(cardID(10), 0, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(20), 1, "Bear", 2, 2),
	)
	in := input(0, v, attackAll(t, mine)...)
	d := decide(t, heuristic.New(), in)
	if !strings.Contains(d.Reason, "all-in for the kill") {
		t.Fatalf("decided %q (%s); a Bear in front of a 7/7 trampler lets 5 through to a player on 5", chose(t, in, d), d.Reason)
	}
}

// TestLethalPushTrampleStopsWhereTheBlockersDo is the other side of the
// same count: the defender is at 6, one point more than a chump lets
// through, and at 5 it has a first striker instead — which may kill the
// Wurm before it assigns any damage, so nothing is counted over it. A
// lethal check that over-counts is an alpha strike into the crack-back.
func TestLethalPushTrampleStopsWhereTheBlockersDo(t *testing.T) {
	for _, tc := range []struct {
		name    string
		life    int
		blocker protocol.CardView
	}{
		{"one point short", 6, creature(cardID(20), 1, "Bear", 2, 2)},
		{"a first striker", 5, creature(cardID(20), 1, "Knight", 2, 2, keywords("first strike"))},
		{"no trample", 5, creature(cardID(20), 1, "Bear", 2, 2)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kw := keywords("trample")
			if tc.name == "no trample" {
				kw = keywords()
			}
			v, mine := trampleBoard(tc.life, creature(cardID(10), 0, "Wurm", 7, 7, kw), tc.blocker)
			in := input(0, v, attackAll(t, mine)...)
			if d := decide(t, heuristic.New(), in); strings.Contains(d.Reason, "all-in for the kill") {
				t.Fatalf("pushed for the kill (%s) where the defender can block its way out", d.Reason)
			}
		})
	}
}

// TestRaceCashesTheWurmThatIsLeftOver is #1504's board at its smallest:
// two Wurms against one Wurm and a Bear, the defender at 8. All-in is
// not lethal — Wurm blocks Wurm, the Bear chumps the other, 5 go over.
// But one Wurm now connects for 7 (they need not block it), and next
// turn the same two into the same two connect for 5 more: 12 ≥ 8, with
// a Wurm kept home that holds anything they can send back. Counting a
// chumped trampler as stopped, NEXT was 0 and the bot passed.
func TestRaceCashesTheWurmThatIsLeftOver(t *testing.T) {
	v, mine := trampleBoard(8,
		creature(cardID(10), 0, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(11), 0, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(20), 1, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(21), 1, "Bear", 2, 2),
	)
	in := input(0, v, attackAll(t, mine)...)
	d := decide(t, heuristic.New(), in)
	if !strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("decided %q (%s); want the two-turn race on the Wurm edge", chose(t, in, d), d.Reason)
	}
	if !strings.Contains(d.Reason, "7 now + 5 next turn") {
		t.Errorf("race numbers %q; want 7 now + 5 next turn", d.Reason)
	}
}
