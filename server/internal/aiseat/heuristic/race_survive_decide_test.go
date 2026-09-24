package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// race_survive_decide_test.go pins #1527 at the decision: the two-turn
// race keeps a blocked attacker that nothing in front of it can kill.

// TestRaceCountsTheChumpedWurmThatLives: two Wurms against one Wurm and
// a Bear, the defender at 6. All-in is not lethal — Wurm blocks Wurm,
// the Bear chumps the other, 5 go over — and one Wurm alone is traded
// off. But the Wurm the Bear chumped is still a 7/7 next turn, into a
// board with nothing left on it: 5 now + 7 next ≥ 6, with nothing of
// theirs left to swing back. Counting every blocked attacker dead, NEXT
// was 0 and the bot passed.
func TestRaceCountsTheChumpedWurmThatLives(t *testing.T) {
	v, mine := trampleBoard(6,
		creature(cardID(10), 0, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(11), 0, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(20), 1, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(21), 1, "Bear", 2, 2),
	)
	in := input(0, v, attackAll(t, mine)...)
	d := decide(t, heuristic.New(), in)
	if !strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("decided %q (%s); want the two-turn race on the chumped Wurm", chose(t, in, d), d.Reason)
	}
	if !strings.Contains(d.Reason, "5 now + 7 next turn") {
		t.Errorf("race numbers %q; want 5 now + 7 next turn", d.Reason)
	}
}

// TestRaceStillCountsAKillableAttackerDead is the same board with one
// more creature on the defender's side that could have blocked the
// second Wurm and killed it. The model still chumps with the Bear (it
// saves the most), but a defender that can kill the Wurm is assumed to,
// so NEXT is 0 again and nothing races — exactly as before #1527.
func TestRaceStillCountsAKillableAttackerDead(t *testing.T) {
	for _, tc := range []struct {
		name   string
		killer protocol.CardView
	}{
		{"a deathtoucher", creature(cardID(22), 1, "Snake", 1, 1, keywords("deathtouch"))},
		{"a double striker", creature(cardID(22), 1, "Duelist", 4, 1, keywords("double strike"))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, mine := trampleBoard(6,
				creature(cardID(10), 0, "Wurm", 7, 7, keywords("trample")),
				creature(cardID(11), 0, "Wurm", 7, 7, keywords("trample")),
				creature(cardID(20), 1, "Wurm", 7, 7, keywords("trample")),
				creature(cardID(21), 1, "Bear", 2, 2),
				tc.killer,
			)
			in := input(0, v, attackAll(t, mine)...)
			if d := decide(t, heuristic.New(), in); strings.Contains(d.Reason, raceLabel) {
				t.Fatalf("raced (%s) on a Wurm %s could have killed", d.Reason, tc.name)
			}
		})
	}
}

// TestRaceSurvivorIsNotHome keeps the crack-back check as pessimistic
// as it was: a blocked attacker that lives is still tapped on the turn
// in between, so it holds nothing. The chumped-Wurm board again, with a
// third seat's Ogre that can swing at me on 4 life: counted home, the
// surviving Wurm would hold the Ogre and the race would go. It does
// not, because the Wurm is tapped and the Ogre is lethal.
func TestRaceSurvivorIsNotHome(t *testing.T) {
	cards := []protocol.CardView{
		creature(cardID(10), 0, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(11), 0, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(20), 1, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(21), 1, "Bear", 2, 2),
		creature(cardID(30), 2, "Ogre", 4, 4),
	}
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(4)), newSeat(1, withLife(6)), newSeat(2, withLife(100))},
		withBattlefield(cards...),
		withTurn(20, 0, "declare_attackers"),
	)
	in := input(0, v, attackAll(t, []string{cards[0].InstanceID, cards[1].InstanceID})...)
	if d := decide(t, heuristic.New(), in); strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("raced (%s) with nothing untapped to hold the Ogre that kills me", d.Reason)
	}
}
