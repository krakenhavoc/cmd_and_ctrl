package heuristic

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// race_soonest_test.go pins #2400's three refinements to the race's
// model of the defender, each behind its own Config switch, and the
// order in which Config.RaceEndsSoonest ranks the swings that qualify.

// TestSwingKeyRanksTheKillThatComesSoonest: fewer turns first, then
// more damage the defender cannot avoid, and a tie is no better, so the
// smaller swing tried first keeps it.
func TestSwingKeyRanksTheKillThatComesSoonest(t *testing.T) {
	for _, tc := range []struct {
		name string
		k, o swingKey
		want bool
	}{
		{"a kill a turn sooner beats more damage now", swingKey{turns: 1, sure: 0}, swingKey{turns: 2, sure: 9}, true},
		{"a kill a turn later loses to less damage now", swingKey{turns: 3, sure: 9}, swingKey{turns: 2, sure: 0}, false},
		{"more sure damage on the same turn", swingKey{turns: 2, sure: 5}, swingKey{turns: 2, sure: 3}, true},
		{"less sure damage on the same turn", swingKey{turns: 2, sure: 3}, swingKey{turns: 2, sure: 5}, false},
		{"a tie keeps the smaller swing", swingKey{turns: 2, sure: 4}, swingKey{turns: 2, sure: 4}, false},
	} {
		if got := tc.k.beats(tc.o); got != tc.want {
			t.Errorf("%s: %+v beats %+v = %v, want %v", tc.name, tc.k, tc.o, got, tc.want)
		}
	}
}

// TestFreeAnswerRefinements reads the free answer with each #2400 rule
// on and off.
//
//   - trample (Config.FreeTrampleOverflow): a Wurm the defender's
//     commander chumps for free puts its 7 less the commander's 3 over
//     (CR 702.19b). Off, the chump holds it in full, as #2310 counted.
//   - tax (Config.FreeCommanderNeedsTax): a commander cast twice costs
//     {1}{R} plus {4} to cast again (CR 903.8). With five lands the
//     defender cannot pay, so blocking with it loses it for good and is
//     not free; with six it can. Off, it is always free.
func TestFreeAnswerRefinements(t *testing.T) {
	const def = "def"
	view := func(lands int) *protocol.GameView {
		v := &protocol.GameView{Seats: []protocol.PlayerView{{ID: def, CommanderCasts: map[string]int{"cmdr": 2}}}}
		for i := 0; i < lands; i++ {
			v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{
				InstanceID: string(rune('a' + i)), Controller: def, TypeLine: "Basic Land — Mountain",
			})
		}
		return v
	}
	cmdr := func() *protocol.CardView {
		c := chief(3, 3)
		c.ManaCost = "{1}{R}"
		return c
	}
	wurm := func() *protocol.CardView { return named(body(7, 7, "trample"), "wurm") }
	ogre := func() *protocol.CardView { return named(body(3, 3), "mine") }
	for _, tc := range []struct {
		name    string
		lands   int
		rules   freeRules
		swing   *protocol.CardView
		through int
		recast  bool
	}{
		{"a chumped Wurm is held in full", 0, freeRules{}, wurm(), 0, true},
		{"a chumped Wurm tramples over", 0, freeRules{trample: true}, wurm(), 4, true},
		{"the tax not counted: the commander is free", 0, freeRules{}, ogre(), 0, true},
		{"five lands cannot pay the tax: not free", 5, freeRules{tax: true}, ogre(), 3, false},
		{"six lands pay the tax: free", 6, freeRules{tax: true}, ogre(), 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &state{view: view(tc.lands), w: DefaultWeights()}
			ans := st.freeAnswer(def, []*protocol.CardView{tc.swing}, []*protocol.CardView{cmdr()}, tc.rules)
			if ans.through != tc.through {
				t.Errorf("through = %d, want %d", ans.through, tc.through)
			}
			if ans.recast["cmdr"] != tc.recast {
				t.Errorf("commander recast = %v, want %v", ans.recast["cmdr"], tc.recast)
			}
		})
	}
}

// TestRecastCommandersSplitsTheDead is Config.RecastChumpForNext's
// split of blockToSurvive's dead: an ordinary creature stays dead, a
// commander is cast again before my next attack (CR 903.9a) and so is
// back to block but too new to swing back, and a commander with haste
// is simply back.
func TestRecastCommandersSplitsTheDead(t *testing.T) {
	bear := named(body(2, 2), "bear")
	cmdr := chief(3, 3)
	hasty := named(body(3, 3, "haste"), "hasty")
	hasty.IsCommander = true
	st := &state{bf: map[string]*protocol.CardView{"bear": bear, "cmdr": cmdr, "hasty": hasty}}
	dead, recast := recastCommanders(st, map[string]bool{"bear": true, "cmdr": true, "hasty": true})
	if !dead["bear"] || dead["cmdr"] || dead["hasty"] || len(dead) != 1 {
		t.Errorf("dead = %v, want the Bear alone", dead)
	}
	if !recast["cmdr"] || recast["hasty"] || len(recast) != 1 {
		t.Errorf("recast = %v, want the commander without haste alone", recast)
	}
}
