package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// attrition_commander_decide_test.go pins #2474 at the decision: the
// attrition horizon trades my commander for theirs when the board
// after it, both commanders cast again, kills
// (Config.HorizonRecastsCommander and Config.HorizonPlaysFreeAnswer).

// lastBoard is the board the full seed-1409 Wurm mirror stopped on:
// my Bear and 3/3 commander, me on 6, against their 3/3 commander on 3.
// Each side has lands to cast its commander again. myLands sets mine.
func lastBoard(t *testing.T, myLands int) (protocol.GameView, []legal.Move) {
	t.Helper()
	v, moves := smallBoard(t, 6, 3,
		creature(cardID(1), 0, "Chief", 3, 3, commander()),
		creature(cardID(2), 0, "Bear", 2, 2),
		creature(cardID(999), 1, "Chief", 3, 3, commander()),
	)
	v = withRecastMana(v)
	for i := 0; i < myLands; i++ {
		v.Battlefield.Cards = append(v.Battlefield.Cards, land(cardID(800+i), 0))
	}
	v.Battlefield.Count = len(v.Battlefield.Cards)
	return v, moves
}

// TestAttritionTradesCommandersForTheKill is the issue's line. Swing
// both: the defender, on 3, has to block my commander with its own,
// both die, and the Bear's 2 put it on 1. Both commanders are cast
// again, mine summoning sick, so I wait a turn; then the Bear and the
// commander swing into its one blocker and one gets through.
//
// The bot declares the commander and the Bear one at a time and then
// passes. With either switch off it sends nothing: blockToSurvive's
// model only guesses at this board and counts both attackers dead, and
// a commander of mine that dies counts as lost.
func TestAttritionTradesCommandersForTheKill(t *testing.T) {
	v, moves := lastBoard(t, 2)
	for i, step := range []struct {
		declare string
		want    string
	}{
		{"", cardID(1)},
		{cardID(1), cardID(2)},
		{cardID(2), "Pass"},
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
		if step.want != "Pass" && !strings.Contains(d.Reason, attritionLabel) {
			t.Fatalf("step %d: reason %q; want an attrition swing", i, d.Reason)
		}
	}

	for _, tc := range []struct {
		name  string
		tweak func(*heuristic.Config)
	}{
		{"my commander counted lost", func(c *heuristic.Config) { c.HorizonRecastsCommander = false }},
		{"the guessed answer", func(c *heuristic.Config) { c.HorizonPlaysFreeAnswer = false }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, moves := lastBoard(t, 2)
			in := input(0, v, moves...)
			if d := decideWith(t, in, tc.tweak); strings.Contains(d.Reason, attritionLabel) {
				t.Fatalf("decided %q (%s); want no attrition swing", chose(t, in, d), d.Reason)
			}
		})
	}
}

// TestAttritionKeepsACommanderItCannotRecast: the same board with no
// lands on my side. My commander that dies in the trade stays in the
// command zone, so the exchange costs me a creature for none of theirs
// and the bot does not make it.
func TestAttritionKeepsACommanderItCannotRecast(t *testing.T) {
	v, moves := lastBoard(t, 0)
	in := input(0, v, moves...)
	if d := decideWith(t, in, nil); strings.Contains(d.Reason, attritionLabel) {
		t.Fatalf("decided %q (%s); my commander cannot be cast again", chose(t, in, d), d.Reason)
	}
}

// TestAttritionRecastsInTheSameTurn is #3080. Swinging both on the
// issue's board, the bot's commander dies and it casts it again in the
// second main phase, from the two lands it left untapped, as it does in
// the full Wurm mirror (#3079). On its next turn the commander can
// attack, so the Bear and the commander are lethal into one blocker:
// the line is one exchange, with no turn spent waiting for it.
//
// With the lands tapped before combat, or with
// Config.HorizonRecastsSameTurn off, the commander is cast on the next
// turn, summoning sick, and the line waits a turn for it: two
// exchanges.
func TestAttritionRecastsInTheSameTurn(t *testing.T) {
	for _, tc := range []struct {
		name   string
		tapped bool
		tweak  func(*heuristic.Config)
		want   string
	}{
		{"cast again after combat", false, nil, "within 1 exchange(s)"},
		{"my lands tapped before combat", true, nil, "within 2 exchange(s)"},
		{"the switch off", false, func(c *heuristic.Config) { c.HorizonRecastsSameTurn = false }, "within 2 exchange(s)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, moves := lastBoard(t, 2)
			if tc.tapped {
				for i := range v.Battlefield.Cards {
					if c := &v.Battlefield.Cards[i]; c.Controller == seatID(0).String() && c.TypeLine == "Basic Land — Mountain" {
						c.Tapped = true
					}
				}
			}
			in := input(0, v, moves...)
			d := decideWith(t, in, tc.tweak)
			if !strings.Contains(d.Reason, attritionLabel) || !strings.Contains(d.Reason, tc.want) {
				t.Fatalf("decided %q (%s); want an attrition swing %s", chose(t, in, d), d.Reason, tc.want)
			}
		})
	}
}

// TestAttritionWaitsForTheCommanderItCastAgain is the line's middle
// turn, when the defender blocks commander with commander and my mana
// did not pay to cast mine again after combat: the defender is on 1,
// its commander is back, and mine was cast again in this turn's first
// main phase and cannot attack. The Bear alone dies to a block that
// costs the defender nothing, so the bot keeps it home. A turn later
// both swing into the one blocker, and one gets through. Where the
// mana did pay, mine was cast in the second main phase of the trade's
// turn (TestAttritionRecastsInTheSameTurn), and the reachable board is
// the second one below, where both attack.
func TestAttritionWaitsForTheCommanderItCastAgain(t *testing.T) {
	board := func(opts ...cardOpt) (protocol.GameView, []legal.Move) {
		v, moves := smallBoard(t, 3, 1,
			creature(cardID(1), 0, "Chief", 3, 3, append([]cardOpt{commander()}, opts...)...),
			creature(cardID(2), 0, "Bear", 2, 2),
			creature(cardID(999), 1, "Chief", 3, 3, commander()),
		)
		return withRecastMana(v), moves
	}

	v, moves := board(sick())
	// The commander is summoning sick: no attack for it. moves is a
	// pass, then an attack by each of mine in board order.
	moves = append(moves[:1], moves[2:]...)
	in := input(0, v, moves...)
	if d := decideWith(t, in, nil); strings.Contains(chose(t, in, d), "Attack") {
		t.Fatalf("decided %q (%s); the Bear alone dies to their commander for nothing", chose(t, in, d), d.Reason)
	}

	v, moves = board()
	in = input(0, v, moves...)
	if d := decideWith(t, in, nil); !strings.Contains(chose(t, in, d), "Attack") {
		t.Fatalf("decided %q (%s); the Bear and the commander are lethal into one blocker", chose(t, in, d), d.Reason)
	}
}
