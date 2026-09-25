package heuristic

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// trample_test.go pins #1504: the attacker's damage estimates count a
// blocked trampler's overflow, and never count more than the engine
// would send over whatever the defender does.

// body is a creature for these boards; kws are its keywords.
func body(p, t int, kws ...string) *protocol.CardView {
	return &protocol.CardView{TypeLine: "Creature", Power: p, Toughness: t, Abilities: kws}
}

// TestUnblockedPowerCountsTrampleOverflow reads the estimate off boards
// small enough to check by hand. Every number is the defender's true
// best except the first striker's, which is 0 on purpose: whether it
// kills the Wurm first is a combat step the estimate does not play.
func TestUnblockedPowerCountsTrampleOverflow(t *testing.T) {
	st := &state{view: &protocol.GameView{}}
	wurm := func() *protocol.CardView { return body(7, 7, "trample") }
	bear := func() *protocol.CardView { return body(2, 2) }
	ogre := func() *protocol.CardView { return body(4, 4) }
	damaged := ogre()
	damaged.DamageMarked = 3
	for _, tc := range []struct {
		name      string
		attackers []*protocol.CardView
		blockers  []*protocol.CardView
		want      int
	}{
		{"a Bear chumps a Wurm: 5 over", []*protocol.CardView{wurm()}, []*protocol.CardView{bear()}, 5},
		{"two Bears on one Wurm absorb 4", []*protocol.CardView{wurm()}, []*protocol.CardView{bear(), bear()}, 3},
		{"a Wurm blocks a Wurm: nothing over", []*protocol.CardView{wurm()}, []*protocol.CardView{wurm()}, 0},
		{"one Bear against a Wurm and an Ogre: it blocks the Ogre", []*protocol.CardView{wurm(), ogre()}, []*protocol.CardView{bear()}, 7},
		{"two Wurms, three Bears: 2+1 on them or 1+1", []*protocol.CardView{wurm(), wurm()}, []*protocol.CardView{bear(), bear(), bear()}, 8},
		{"deathtouch assigns 1 per blocker", []*protocol.CardView{body(7, 7, "trample", "deathtouch")}, []*protocol.CardView{ogre(), ogre(), ogre()}, 4},
		{"damage already marked is part of lethal", []*protocol.CardView{wurm()}, []*protocol.CardView{damaged}, 6},
		{"an indestructible blocker still only absorbs lethal", []*protocol.CardView{wurm()}, []*protocol.CardView{body(2, 2, "indestructible")}, 5},
		{"a first striker may kill it first: nothing counted", []*protocol.CardView{wurm()}, []*protocol.CardView{body(2, 2, "first strike")}, 0},
		{"a first-striking trampler is not held by one", []*protocol.CardView{body(7, 7, "trample", "first strike")}, []*protocol.CardView{body(2, 2, "first strike")}, 5},
		{"a flying trampler over Bears is unblocked", []*protocol.CardView{body(7, 7, "trample", "flying")}, []*protocol.CardView{bear(), bear()}, 7},
		// The two that pin trampleBound's relaxation: a second blocker
		// on a trampler is worth only what the first left over (the
		// slot values), and a blocker behind one that already holds
		// the trampler saves nothing at all — the relaxation credits
		// it, and unblockedPower's max with its own matching is what
		// takes the credit back.
		{"an Ogre second in line absorbs only what is left", []*protocol.CardView{wurm(), wurm(), bear()}, []*protocol.CardView{wurm(), ogre(), ogre()}, 2},
		{"a spare Ogre behind a held Wurm saves nothing", []*protocol.CardView{wurm(), body(3, 3, "flying")}, []*protocol.CardView{wurm(), ogre()}, 3},
		{"no trample: the Bear stops the Ogre", []*protocol.CardView{ogre()}, []*protocol.CardView{bear()}, 0},
		{"no trample: one of two gets through", []*protocol.CardView{ogre(), bear()}, []*protocol.CardView{bear()}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := unblockedPower(st, "", tc.attackers, tc.blockers); got != tc.want {
				t.Fatalf("unblockedPower = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestUnblockedPowerNeverOverestimates is the conservative half as a
// property: on random small boards the estimate is never more than
// what the defender's best blocks actually let through, found by trying
// every set of blocks and assigning damage the way the engine does. On
// boards without a trampler it is exactly that — the #1261 matching,
// unchanged.
//
// First strike stays out of the random boards: whether a group of
// first strikers kills a trampler before it deals damage is a second
// combat step this brute force does not play, and trampleAbsorb counts
// a first striker as holding it in full, which is the safe side.
func TestUnblockedPowerNeverOverestimates(t *testing.T) {
	st := &state{view: &protocol.GameView{}}
	rng := rand.New(rand.NewPCG(1504, 1))
	random := func(attacker bool) *protocol.CardView {
		c := body(rng.IntN(8), 1+rng.IntN(7))
		if attacker && rng.IntN(2) == 0 {
			c.Abilities = append(c.Abilities, "trample")
		}
		if attacker && rng.IntN(6) == 0 {
			c.Abilities = append(c.Abilities, "deathtouch")
		}
		if rng.IntN(5) == 0 {
			c.Abilities = append(c.Abilities, "flying")
		}
		if !attacker && rng.IntN(6) == 0 {
			c.Abilities = append(c.Abilities, "reach")
		}
		if !attacker && rng.IntN(4) == 0 {
			c.DamageMarked = rng.IntN(c.Toughness)
		}
		return c
	}
	exact := 0
	for n := 0; n < 3000; n++ {
		attackers := make([]*protocol.CardView, 1+rng.IntN(4))
		for i := range attackers {
			attackers[i] = random(true)
		}
		blockers := make([]*protocol.CardView, rng.IntN(5))
		for i := range blockers {
			blockers[i] = random(false)
		}
		got := unblockedPower(st, "", attackers, blockers)
		best := bruteBestBlocks(st, attackers, blockers)
		if got > best {
			t.Fatalf("board %d: estimate %d > the defender's best %d\n%s", n, got, best, describe(attackers, blockers))
		}
		trample := false
		for _, a := range attackers {
			trample = trample || hasKeyword(a, "trample")
		}
		if !trample && got != best {
			t.Fatalf("board %d, no trampler: estimate %d != the defender's best %d\n%s", n, got, best, describe(attackers, blockers))
		}
		if got == best {
			exact++
		}
	}
	t.Logf("exact on %d of 3000 boards", exact)
}

// bruteBestBlocks tries every assignment of blockers to attackers
// (each blocker blocks one attacker it can block, or nothing) and
// returns the least damage any of them lets through, with damage
// assigned as the engine assigns it: a blocked attacker without
// trample deals the player nothing, and a trampler assigns each of its
// blockers lethal damage — its toughness less marked damage, 1 from a
// deathtouch source — and the rest to the player.
func bruteBestBlocks(st *state, attackers, blockers []*protocol.CardView) int {
	pick := make([]int, len(blockers))
	best := -1
	var rec func(i int)
	rec = func(i int) {
		if i == len(blockers) {
			through := 0
			for ai, a := range attackers {
				if a.Power <= 0 {
					continue
				}
				blocked, lethal := false, 0
				for bi, b := range blockers {
					if pick[bi] != ai+1 {
						continue
					}
					blocked = true
					need := max(b.Toughness-b.DamageMarked, 1)
					if hasKeyword(a, "deathtouch") {
						need = 1
					}
					lethal += need
				}
				switch {
				case !blocked:
					through += a.Power
				case hasKeyword(a, "trample"):
					through += max(a.Power-lethal, 0)
				}
			}
			if best < 0 || through < best {
				best = through
			}
			return
		}
		for ai := 0; ai <= len(attackers); ai++ {
			if ai > 0 && !couldBlock(st, "", attackers[ai-1], blockers[i]) {
				continue
			}
			pick[i] = ai
			rec(i + 1)
		}
	}
	rec(0)
	return best
}

// TestBlockToSurviveCountsTrampleOverAChump is the race's NOW number:
// a defender that has to block a trampler to live chumps it, and the
// chump saves only its own toughness. Before #1504 a chumped Wurm
// connected for nothing, so a swing that leaves the defender on 1 read
// as a swing that did nothing at all.
func TestBlockToSurviveCountsTrampleOverAChump(t *testing.T) {
	p := New()
	st := &state{view: &protocol.GameView{}}
	id := func(c *protocol.CardView, s string) *protocol.CardView { c.InstanceID = s; return c }
	wurm := id(body(7, 7, "trample"), "wurm")

	t.Run("one chump", func(t *testing.T) {
		bear := id(body(2, 2), "bear")
		now, dead, blocked := p.blockToSurvive(st, &SeatEval{ID: "def", Life: 6},
			[]*protocol.CardView{wurm}, []*protocol.CardView{bear}).parts()
		if now != 5 {
			t.Errorf("NOW = %d, want 5: the Bear absorbs 2 of the Wurm's 7", now)
		}
		if !dead["bear"] || !blocked["wurm"] {
			t.Errorf("dead %v, blocked %v; want the Bear dead and the Wurm blocked", dead, blocked)
		}
	})

	t.Run("a second chump while the overflow still kills", func(t *testing.T) {
		b1, b2 := id(body(2, 2), "b1"), id(body(2, 2), "b2")
		now, dead, _ := p.blockToSurvive(st, &SeatEval{ID: "def", Life: 4},
			[]*protocol.CardView{wurm}, []*protocol.CardView{b1, b2}).parts()
		if now != 3 || !dead["b1"] || !dead["b2"] {
			t.Errorf("NOW = %d, dead %v; want 3 with both Bears dead", now, dead)
		}
	})

	t.Run("an indestructible blocker survives and 5 still go over", func(t *testing.T) {
		wall := id(body(2, 2, "indestructible"), "wall")
		now, dead, _ := p.blockToSurvive(st, &SeatEval{ID: "def", Life: 3},
			[]*protocol.CardView{wurm}, []*protocol.CardView{wall}).parts()
		if now != 5 || dead["wall"] {
			t.Errorf("NOW = %d, dead %v; want 5 through and the wall alive", now, dead)
		}
	})

	t.Run("without trample the chump holds it", func(t *testing.T) {
		ogre := id(body(7, 7), "ogre")
		bear := id(body(2, 2), "bear")
		now, _, _ := p.blockToSurvive(st, &SeatEval{ID: "def", Life: 6},
			[]*protocol.CardView{ogre}, []*protocol.CardView{bear}).parts()
		if now != 0 {
			t.Errorf("NOW = %d, want 0", now)
		}
	})
}

func describe(attackers, blockers []*protocol.CardView) string {
	s := ""
	for _, a := range attackers {
		s += fmt.Sprintf("  attacker %d/%d %v\n", a.Power, a.Toughness, a.Abilities)
	}
	for _, b := range blockers {
		s += fmt.Sprintf("  blocker  %d/%d (damage %d) %v\n", b.Power, b.Toughness, b.DamageMarked, b.Abilities)
	}
	return s
}
