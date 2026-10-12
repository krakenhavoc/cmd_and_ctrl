package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// race_soonest_decide_test.go pins #2400 at the decision: among the
// swings that race, the bot sends the one whose kill comes soonest
// (Config.RaceEndsSoonest), and the defender's commander is a free
// blocker only while its controller can pay to cast it again
// (Config.FreeCommanderNeedsTax).

func smallestFirst(c *heuristic.Config) { c.RaceEndsSoonest = false }

// withRecastMana gives seat 1 the two lands it needs to cast a {1}{R}
// commander again from the command zone, so the commander is a free
// blocker under Config.FreeCommanderNeedsTax as #2310's boards assume.
func withRecastMana(v protocol.GameView) protocol.GameView {
	cards := append([]protocol.CardView(nil), v.Battlefield.Cards...)
	cards = append(cards, land(cardID(700), 1), land(cardID(701), 1))
	v.Battlefield.Cards = cards
	v.Battlefield.Count = len(cards)
	return v
}

// TestRaceSendsTheSwingThatEndsSoonest is loopBoard with their 3/3 an
// ordinary creature. Three swings race:
//
//   - my commander alone, 3 now + 4 next turn, which is the smallest and
//     what the race used to send;
//   - my commander and two Bears, 5 now + 2 next turn;
//   - everything, 4 now + 4 next turn, since the defender has to chump.
//
// All three kill in two turns, so the damage the defender cannot avoid
// decides it: 5, against 3 and 4. The bot declares the commander and
// two Bears one at a time, and once they are in it passes, because the
// rest is the reserve the crack-back counted on.
func TestRaceSendsTheSwingThatEndsSoonest(t *testing.T) {
	v, moves := loopBoard(t, false)
	for i, step := range []struct {
		declare string // the attacker declared before this decision
		want    string // what the bot does: an attacker's ID, or Pass
	}{
		{"", cardID(1)},
		{cardID(1), cardID(10)},
		{cardID(10), cardID(11)},
		{cardID(11), "Pass"},
	} {
		if step.declare != "" {
			v, moves = declared(v, moves, step.declare)
		}
		in := input(0, v, moves...)
		d := decideWith(t, in, nil)
		got := chose(t, in, d)
		if !strings.Contains(got, step.want) {
			t.Fatalf("step %d: decided %q (%s); want %s", i, got, d.Reason, step.want)
		}
		if step.want != "Pass" && !strings.Contains(d.Reason, "5 now + 2 next turn") {
			t.Fatalf("step %d: reason %q; want the race on my commander and two Bears, 5 now + 2 next turn", i, d.Reason)
		}
	}

	// The old objective sends the commander alone.
	v, moves = loopBoard(t, false)
	in := input(0, v, moves...)
	if d := decideWith(t, in, smallestFirst); !strings.Contains(d.Reason, "3 now + 4 next turn") {
		t.Fatalf("smallest first: decided %q (%s); want the lone commander's race", chose(t, in, d), d.Reason)
	}
}

// TestRaceCountsACommanderItsControllerCannotRecast is the issue's
// loopBoard, their 3/3 their commander, cast twice already: the next
// cast costs its {1}{R} plus {4} (CR 903.8). With no mana sources the
// defender cannot cast it again, so blocking with it is not free. The
// race is then the same as against an ordinary 3/3, 5 now + 2 next
// turn. With six lands it can, and the race is the one #2310 pinned, 4
// now + 2 next turn.
func TestRaceCountsACommanderItsControllerCannotRecast(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lands int
		tweak func(*heuristic.Config)
		want  string
	}{
		{"priced out", 0, nil, "5 now + 2 next turn"},
		{"priced out, the tax not counted", 0, func(c *heuristic.Config) { c.FreeCommanderNeedsTax = false }, "4 now + 2 next turn"},
		{"six lands pay for it", 6, nil, "4 now + 2 next turn"},
		{"five lands do not", 5, nil, "5 now + 2 next turn"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, moves := loopBoard(t, true)
			v.Seats[1] = newSeat(1, withLife(6), withCommanderCasts(2))
			for i := 0; i < tc.lands; i++ {
				v.Battlefield.Cards = append(v.Battlefield.Cards, land(cardID(500+i), 1))
			}
			v.Battlefield.Count = len(v.Battlefield.Cards)
			in := input(0, v, moves...)
			d := decideWith(t, in, tc.tweak)
			if !strings.Contains(d.Reason, raceLabel) || !strings.Contains(d.Reason, tc.want) {
				t.Fatalf("decided %q (%s); want the race %s", chose(t, in, d), d.Reason, tc.want)
			}
		})
	}
}
