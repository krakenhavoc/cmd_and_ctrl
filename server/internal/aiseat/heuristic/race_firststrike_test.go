package heuristic

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// race_firststrike_test.go pins #1549: a first-strike blocker that kills
// its attacker in the first combat damage step takes no damage back
// (CR 510.4), so blockToSurvive must not count it dead.

func TestBlockerDies(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b *protocol.CardView
		want bool
	}{
		{"a 4/4 first striker blocks a 4/4: it lives", body(4, 4), body(4, 4, "first strike"), false},
		{"a first striker too small to kill first dies", body(4, 4), body(3, 3, "first strike"), true},
		{"vs a first striker, both strike at once: it dies", body(4, 4, "first strike"), body(4, 4, "first strike"), true},
		{"vs a double striker: it dies", body(4, 4, "double strike"), body(4, 4, "first strike"), true},
		{"a double striker too small to kill in its first hit dies", body(4, 4), body(2, 4, "double strike"), true},
		{"a double striker that kills in its first hit lives", body(4, 4), body(4, 4, "double strike"), false},
		{"a deathtouch first striker kills any attacker first", body(6, 6), body(1, 3, "first strike", "deathtouch"), false},
		{"a deathtouch attacker kills a plain blocker", body(1, 1, "deathtouch"), body(4, 4), true},
		{"a deathtouch attacker loses to a first striker that kills it first", body(2, 2, "deathtouch"), body(2, 2, "first strike"), false},
		{"an indestructible attacker is not killed first", body(4, 4, "indestructible"), body(4, 4, "first strike"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := blockerDies(tc.a, tc.b); got != tc.want {
				t.Fatalf("blockerDies = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBlockToSurviveKeepsAFirstStrikerThatKillsFirst(t *testing.T) {
	st := &state{view: &protocol.GameView{}}
	for _, tc := range []struct {
		name     string
		attacker *protocol.CardView
		blocker  *protocol.CardView
		dead     bool
	}{
		{"kills first", body(4, 4), body(4, 4, "first strike"), false},
		{"attacker has first strike", body(4, 4, "first strike"), body(4, 4, "first strike"), true},
		{"attacker has double strike", body(4, 4, "double strike"), body(4, 4, "first strike"), true},
		{"not enough power", body(4, 4), body(3, 4, "first strike"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.attacker.InstanceID, tc.blocker.InstanceID = "atk", "blk"
			d := New().blockToSurvive(st, &SeatEval{ID: "def", Life: 4},
				[]*protocol.CardView{tc.attacker}, []*protocol.CardView{tc.blocker})
			if !d.blockedMine["atk"] {
				t.Fatalf("attacker left unblocked (through %d)", d.through)
			}
			if d.deadDef["blk"] != tc.dead {
				t.Errorf("blocker dead = %v, want %v", d.deadDef["blk"], tc.dead)
			}
		})
	}
}
