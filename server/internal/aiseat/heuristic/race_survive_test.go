package heuristic

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// race_survive_test.go pins #1527: the race's NEXT keeps a blocked
// attacker of mine that nothing the defender could put in front of it
// kills.

// parts is blockToSurvive's answer as the three numbers #1504's tests
// were written against.
func (d defence) parts() (int, map[string]bool, map[string]bool) {
	return d.through, d.deadDef, d.blockedMine
}

// TestBlockedAttackerSurvives reads the survival rule off boards small
// enough to check by hand: my one attacker, the defender's blockers,
// the defender's life (which decides how many of them it has to use),
// and whether the attacker lives to swing again next turn.
func TestBlockedAttackerSurvives(t *testing.T) {
	st := &state{view: &protocol.GameView{}}
	named := func(c *protocol.CardView, id string) *protocol.CardView { c.InstanceID = id; return c }
	wurm := func() *protocol.CardView { return body(7, 7, "trample") }
	bear := func(id string) *protocol.CardView { return named(body(2, 2), id) }
	red := func(c *protocol.CardView) *protocol.CardView { c.Colors = []string{"R"}; return c }
	proRed := func(c *protocol.CardView) *protocol.CardView {
		c.Protection = []protocol.ProtectionView{{Printed: "red", Kind: "color", Value: "R"}}
		return c
	}
	for _, tc := range []struct {
		name     string
		attacker *protocol.CardView
		blockers []*protocol.CardView
		life     int
		want     bool
	}{
		// The issue's board: a Wurm with only a Bear to stop it.
		{"a Bear chumps a Wurm: it lives", wurm(), []*protocol.CardView{bear("b1")}, 6, true},
		{"two Bears on a Wurm: 4 < 7, it lives", wurm(), []*protocol.CardView{bear("b1"), bear("b2")}, 4, true},
		{"a Wurm blocks a Wurm: it dies", wurm(), []*protocol.CardView{named(wurm(), "w")}, 6, false},
		{"two Ogres on a Wurm: 8 ≥ 7, it dies", wurm(),
			[]*protocol.CardView{named(body(4, 4), "o1"), named(body(4, 4), "o2")}, 2, false},
		// Blocks to kill when it can: the model chumps with the Bear,
		// but a spare creature that kills the Wurm alone would be put
		// there instead.
		{"a spare deathtoucher could block instead: it dies", wurm(),
			[]*protocol.CardView{bear("b1"), named(body(1, 1, "deathtouch"), "dt")}, 6, false},
		{"a spare double striker could block instead: it dies", wurm(),
			[]*protocol.CardView{bear("b1"), named(body(4, 1, "double strike"), "ds")}, 6, false},
		// A spare that could only kill it together with the chump is a
		// gang block, which the model does not play (see survives).
		{"a spare Ogre behind the chumping Ogre: it lives", wurm(),
			[]*protocol.CardView{named(body(4, 4), "o1"), named(body(4, 4), "o2")}, 6, true},
		// First strike and deathtouch, one blocker at a time.
		{"a double striker blocks: 8 ≥ 7, it dies", wurm(), []*protocol.CardView{named(body(4, 4, "double strike"), "ds")}, 4, false},
		{"a deathtouch chump: it dies", wurm(), []*protocol.CardView{named(body(1, 1, "deathtouch"), "dt")}, 7, false},
		{"my first striker kills the blocker first: it lives", body(3, 3, "first strike"),
			[]*protocol.CardView{named(body(3, 3), "k")}, 1, true},
		{"my first striker into a first striker: it dies", body(3, 3, "first strike"),
			[]*protocol.CardView{named(body(3, 3, "first strike"), "k")}, 1, false},
		{"my first striker cannot kill a deathtoucher's 4 toughness: it dies", body(3, 3, "first strike"),
			[]*protocol.CardView{named(body(1, 4, "deathtouch"), "dt")}, 1, false},
		{"indestructible: it lives", body(7, 7, "trample", "indestructible"), []*protocol.CardView{named(wurm(), "w")}, 6, true},
		{"protected from the blocker: it lives", proRed(wurm()), []*protocol.CardView{red(named(wurm(), "w"))}, 6, true},
		{"damage already marked counts", func() *protocol.CardView { w := wurm(); w.DamageMarked = 5; return w }(),
			[]*protocol.CardView{bear("b1")}, 6, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.attacker.InstanceID = "atk"
			d := New().blockToSurvive(st, &SeatEval{ID: "def", Life: tc.life}, []*protocol.CardView{tc.attacker}, tc.blockers)
			if !d.blockedMine["atk"] {
				t.Fatalf("the model left the attacker unblocked (through %d); the board does not ask the question", d.through)
			}
			if got := d.survives(st, "def", tc.attacker); got != tc.want {
				t.Fatalf("survives = %v, want %v (blocked by %d, %d spare)", got, tc.want, len(d.blockedBy["atk"]), len(d.spare))
			}
		})
	}
}

// TestGuessedBlocksKillEverything: where the greedy cannot find the
// defender's blocks, blockToSurvive assumes every attacker blocked by
// something it cannot name, and none of them is counted alive.
func TestGuessedBlocksKillEverything(t *testing.T) {
	d := defence{guessed: true, blockedMine: map[string]bool{"atk": true}}
	a := body(7, 7, "trample")
	a.InstanceID = "atk"
	if d.survives(&state{view: &protocol.GameView{}}, "def", a) {
		t.Fatal("a guessed block counted the attacker alive")
	}
}
