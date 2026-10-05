package heuristic

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// gang_test.go pins #1548's combat math: one attacker against every
// creature blocking it.

func named(c *protocol.CardView, id string) *protocol.CardView { c.InstanceID = id; return c }

// TestGangKills reads the gang rule off groups small enough to check by
// hand, the timing rules first: a first striker picks off the blockers
// it can before they deal regular damage, and a group that only adds up
// on paper does not kill it.
func TestGangKills(t *testing.T) {
	wurm := func() *protocol.CardView { return body(7, 7, "trample") }
	bear := func() *protocol.CardView { return body(2, 2) }
	ogre := func() *protocol.CardView { return body(4, 4) }
	group := func(cs ...*protocol.CardView) []*protocol.CardView { return cs }
	for _, tc := range []struct {
		name     string
		attacker *protocol.CardView
		blockers []*protocol.CardView
		want     bool
	}{
		{"one Ogre chumps a Wurm", wurm(), group(ogre()), false},
		{"two Ogres kill a Wurm: 8 ≥ 7", wurm(), group(ogre(), ogre()), true},
		{"a Bear and an Ogre do not: 6 < 7", wurm(), group(bear(), ogre()), false},
		{"three Bears do not: 6 < 7", wurm(), group(bear(), bear(), bear()), false},
		{"four Bears do: 8 ≥ 7", wurm(), group(bear(), bear(), bear(), bear()), true},
		{"no blockers", wurm(), nil, false},
		{"a 0-power wall deals nothing", body(1, 1), group(body(0, 8)), false},

		// The issue's first-strike case: the first striker kills one
		// Bear before it deals damage, and the other Bear's 2 do not
		// finish a 2/3.
		{"a first striker kills one of two Bears first", body(2, 3, "first strike"), group(bear(), bear()), false},
		{"without first strike the two Bears kill it", body(2, 3), group(bear(), bear()), true},
		// It spends its first-strike damage where it saves it: the 3/3
		// that would finish it, not the Bear.
		{"a first striker kills the blocker that matters", body(3, 3, "first strike"), group(bear(), body(3, 3)), false},
		// …and cannot when that blocker is too big to kill first.
		{"a first striker cannot kill an Ogre first", body(3, 3, "first strike"), group(bear(), ogre()), true},
		// Every deathtoucher has to go first, or it is lethal.
		{"a first striker kills the deathtoucher first", body(3, 3, "first strike"), group(body(1, 1, "deathtouch"), bear()), false},
		{"a first striker cannot kill two deathtouchers", body(1, 3, "first strike"), group(body(1, 1, "deathtouch"), body(1, 1, "deathtouch")), true},
		{"deathtouch kills without first strike", body(7, 7), group(body(1, 1, "deathtouch")), true},
		// Blockers that strike first deal their damage whatever the
		// attacker does.
		{"two first-strike Bears kill a 4/4 first", body(4, 4, "first strike"), group(body(2, 2, "first strike"), body(2, 2, "first strike")), true},
		{"a first-strike deathtoucher is lethal first", body(4, 4, "first strike"), group(body(1, 1, "first strike", "deathtouch")), true},
		// A double striker hits in both steps, unless the first striker
		// kills it in the first.
		{"a double-strike Bear deals 4", body(3, 4), group(body(2, 2, "double strike")), true},
		{"a first striker kills a double striker after one hit", body(3, 3, "first strike"), group(body(2, 2, "double strike")), false},
		{"a deathtouch first striker needs 1 point per blocker", body(2, 4, "first strike", "deathtouch"), group(ogre(), ogre(), bear()), false},
		{"indestructible never dies", body(7, 7, "indestructible"), group(ogre(), ogre()), false},
		{"protection prevents the damage", func() *protocol.CardView {
			w := wurm()
			w.Protection = []protocol.ProtectionView{{Printed: "red", Kind: "color", Value: "R"}}
			return w
		}(), group(func() *protocol.CardView { o := ogre(); o.Colors = []string{"R"}; return o }(), ogre()), false},
		{"marked damage counts", func() *protocol.CardView { w := wurm(); w.DamageMarked = 4; return w }(), group(bear(), body(1, 1)), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := gangKills(tc.attacker, tc.blockers); got != tc.want {
				t.Fatalf("gangKills = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestLossesKillTheMostValue: the attacker divides its damage as badly
// for the defender as it can. Cheapest-first would let a second blocker
// "save" the valuable one, which is how a gang-aware block planner
// would learn to throw an Ogre beside the Wurm blocking a Wurm.
func TestLossesKillTheMostValue(t *testing.T) {
	st := &state{w: DefaultWeights()}
	wurm := named(body(7, 7, "trample"), "w")
	theirWurm := named(body(7, 7, "trample"), "tw")
	o1, o2 := named(body(4, 4), "o1"), named(body(4, 4), "o2")
	bear := named(body(2, 2), "b")
	ids := func(cs []*protocol.CardView) []string {
		var out []string
		for _, c := range cs {
			out = append(out, c.InstanceID)
		}
		return out
	}
	for _, tc := range []struct {
		name     string
		attacker *protocol.CardView
		blockers []*protocol.CardView
		want     []string
	}{
		{"a Wurm kills the Wurm, not the Ogre beside it", wurm, []*protocol.CardView{theirWurm, o1}, []string{"tw"}},
		{"a Wurm kills one of two Ogres, the first on a tie", wurm, []*protocol.CardView{o1, o2}, []string{"o1"}},
		{"a Wurm kills a Bear and an Ogre", wurm, []*protocol.CardView{bear, o1}, []string{"b", "o1"}},
		{"a 1-power deathtoucher kills one", named(body(1, 3, "deathtouch"), "dt"), []*protocol.CardView{bear, o1}, []string{"o1"}},
		{"a double striker divides twice its power", named(body(2, 2, "double strike"), "ds"), []*protocol.CardView{o1}, []string{"o1"}},
		{"killed by first strike first, it kills nothing", named(body(4, 2), "a"), []*protocol.CardView{named(body(2, 2, "first strike"), "fs"), o1}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ids(st.losses(tc.attacker, tc.blockers))
			if len(got) != len(tc.want) {
				t.Fatalf("losses = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("losses = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

// TestBlockToSurviveGangsUpToKill is the defender model's half of
// #1548: once the defender would live, a block that does not kill its
// attacker becomes one when spare blockers can join it for little
// enough, and the joiners soak up a trampler's overflow — NOW falls by
// what they absorb (the issue's blockToSurvive/trample interaction).
func TestBlockToSurviveGangsUpToKill(t *testing.T) {
	st := &state{view: &protocol.GameView{}, w: DefaultWeights()}
	p := New()

	t.Run("the spare Ogre joins the chumped Wurm", func(t *testing.T) {
		w := named(body(7, 7, "trample"), "w")
		o1, o2 := named(body(4, 4), "o1"), named(body(4, 4), "o2")
		d := p.blockToSurvive(st, &SeatEval{ID: "def", Life: 6}, []*protocol.CardView{w}, []*protocol.CardView{o1, o2})
		if d.through != 0 {
			t.Errorf("NOW = %d, want 0: two Ogres absorb all 7", d.through)
		}
		if len(d.blockedBy["w"]) != 2 || len(d.spare) != 0 {
			t.Errorf("blocked by %d, %d spare; want both Ogres on the Wurm", len(d.blockedBy["w"]), len(d.spare))
		}
		// The bot assigns down the declared order: the chump dies, the
		// joiner lives.
		if !d.deadDef["o1"] || d.deadDef["o2"] {
			t.Errorf("dead %v; want the first Ogre only", d.deadDef)
		}
		if d.survives(st, "def", w) {
			t.Error("the Wurm survives two Ogres")
		}
	})

	t.Run("a join that cannot kill stays home", func(t *testing.T) {
		w := named(body(7, 7, "trample"), "w")
		b1, b2 := named(body(2, 2), "b1"), named(body(2, 2), "b2")
		d := p.blockToSurvive(st, &SeatEval{ID: "def", Life: 6}, []*protocol.CardView{w}, []*protocol.CardView{b1, b2})
		if d.through != 5 || len(d.spare) != 1 {
			t.Errorf("NOW = %d with %d spare; want 5 over one Bear and the other kept", d.through, len(d.spare))
		}
	})

	t.Run("a join costing more than the kill is not made", func(t *testing.T) {
		// A 2/3 deathtoucher chumped by a Bear kills an Ogre that
		// joins as well: the Ogre is worth more than the deathtoucher,
		// so the defender keeps it.
		a := named(body(2, 3, "deathtouch"), "a")
		b, o := named(body(2, 2), "b"), named(body(4, 4), "o")
		if cost := st.lossValue(a, []*protocol.CardView{b, o}) - st.lossValue(a, []*protocol.CardView{b}); cost <= st.w.CombatValue(a) {
			t.Fatalf("the join costs %.2f, not more than the %.2f kill; the case does not ask the question", cost, st.w.CombatValue(a))
		}
		if _, ok := st.gangJoin(a, []*protocol.CardView{b}, []*protocol.CardView{o}); ok {
			t.Error("gangJoin made a join that costs more than the kill")
		}
	})

	t.Run("the spare goes to one Wurm, and the other lives", func(t *testing.T) {
		w1, w2 := named(body(7, 7, "trample"), "w1"), named(body(7, 7, "trample"), "w2")
		o1, o2, o3 := named(body(4, 4), "o1"), named(body(4, 4), "o2"), named(body(4, 4), "o3")
		d := p.blockToSurvive(st, &SeatEval{ID: "def", Life: 8}, []*protocol.CardView{w1, w2}, []*protocol.CardView{o1, o2, o3})
		if d.through != 3 {
			t.Errorf("NOW = %d, want 3: one Wurm held by two Ogres, the other chumped with 3 over", d.through)
		}
		dead := 0
		for _, w := range []*protocol.CardView{w1, w2} {
			if !d.survives(st, "def", w) {
				dead++
			}
		}
		if dead != 1 {
			t.Errorf("%d Wurms counted dead, want 1: the one spare Ogre can join only one block", dead)
		}
	})
}

// TestOrderedKills is what the bot's own canonical damage assignment
// kills: lethal down the declared order while the damage lasts.
func TestOrderedKills(t *testing.T) {
	w := body(7, 7, "trample")
	o1, o2 := named(body(4, 4), "o1"), named(body(4, 4), "o2")
	b := named(body(2, 2), "b")
	wall := named(body(0, 3, "indestructible"), "wall")
	for _, tc := range []struct {
		name     string
		blockers []*protocol.CardView
		want     int
	}{
		{"chump first, joiner lives", []*protocol.CardView{o1, o2}, 1},
		{"Bear then Ogre: both", []*protocol.CardView{b, o1}, 2},
		{"an indestructible wall soaks its 3 and lives", []*protocol.CardView{wall, o1}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(orderedKills(w, tc.blockers)); got != tc.want {
				t.Fatalf("%d killed, want %d", got, tc.want)
			}
		})
	}
}
