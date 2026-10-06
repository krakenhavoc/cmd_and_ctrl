package heuristic

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// race_free_test.go pins #2310's model: the defender's free answer to a
// swing — block with everything whose block costs it nothing, a
// blocker that lives or its commander, and take the rest.

// chief is the defender's commander, "cmdr".
func chief(p, t int, kws ...string) *protocol.CardView {
	c := named(body(p, t, kws...), "cmdr")
	c.IsCommander = true
	return c
}

// TestFreeAnswer reads the free answer off boards small enough to check
// by hand: what connects past the free blocks, which of my attackers
// they kill, and whether the defender's commander dies and is cast
// again.
func TestFreeAnswer(t *testing.T) {
	st := &state{view: &protocol.GameView{}, w: DefaultWeights()}
	bear := func(id string) *protocol.CardView { return named(body(2, 2), id) }
	for _, tc := range []struct {
		name     string
		swing    []*protocol.CardView
		blockers []*protocol.CardView
		through  int
		dead     []string // my attackers the free blocks kill
		recast   bool     // the commander dies blocking, is cast again, and cannot swing back
	}{
		// The issue's loop: one commander into the other. Both die,
		// theirs comes back, and nothing connects.
		{"their commander blocks mine",
			[]*protocol.CardView{named(body(3, 3), "mine")}, []*protocol.CardView{chief(3, 3)},
			0, []string{"mine"}, true},
		// Cast again it is too new to attack — unless it has haste.
		{"a commander with haste can swing back",
			[]*protocol.CardView{named(body(3, 3), "mine")}, []*protocol.CardView{chief(3, 3, "haste")},
			0, []string{"mine"}, false},
		{"a commander that lives is not cast again",
			[]*protocol.CardView{bear("mine")}, []*protocol.CardView{chief(3, 3)},
			0, []string{"mine"}, false},
		// A blocked trampler counts as held in full, as #1548's
		// freeThrough did: an undercount, the direction the race errs.
		{"a commander chumping a Wurm holds it in full",
			[]*protocol.CardView{named(body(7, 7, "trample"), "wurm")}, []*protocol.CardView{chief(3, 3)},
			0, nil, true},
		{"a blocker that lives is free, and kills",
			[]*protocol.CardView{bear("mine")}, []*protocol.CardView{named(body(4, 4), "ogre")},
			0, []string{"mine"}, false},
		{"a blocker that dies is not free",
			[]*protocol.CardView{named(body(4, 4), "mine")}, []*protocol.CardView{bear("b")},
			4, nil, false},
		// A spare free blocker joins the block to kill, as
		// defence.survives assumes for the other answer.
		{"two walls that live kill a 5/5 between them",
			[]*protocol.CardView{named(body(5, 5), "mine")},
			[]*protocol.CardView{named(body(3, 6), "w1"), named(body(3, 6), "w2")},
			0, []string{"mine"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ans := st.freeAnswer("def", tc.swing, tc.blockers)
			if ans.through != tc.through {
				t.Errorf("through = %d, want %d", ans.through, tc.through)
			}
			want := map[string]bool{}
			for _, id := range tc.dead {
				want[id] = true
			}
			for _, a := range tc.swing {
				if ans.mineDead[a.InstanceID] != want[a.InstanceID] {
					t.Errorf("%s dead = %v, want %v", a.InstanceID, ans.mineDead[a.InstanceID], want[a.InstanceID])
				}
			}
			if ans.recast["cmdr"] != tc.recast {
				t.Errorf("commander recast = %v, want %v", ans.recast["cmdr"], tc.recast)
			}
			if len(ans.dead) != 0 {
				t.Errorf("dead %v; nothing of the defender's dies for good in the free answer", ans.dead)
			}
		})
	}
}
