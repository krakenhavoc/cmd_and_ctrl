package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// attack_direction_test.go — #4211, ADR 0206: the bot answering "choose
// left or right" (Pramikon, Sky Rampart; Teyo, Geometric Tactician).
// Left is the next player in turn order, right the previous one, and
// every player may then attack only the nearest opponent that way. So
// under "left" the bot is attacked by the player before it, and under
// "right" by the player after it: it takes the direction whose attacker
// has the weaker board.

// directionInput is a four-seat table with the bot in seat 0 and the
// given creature power on the other seats, answering a left-or-right
// prompt whose options come in the given order.
func directionInput(t *testing.T, power map[int]int, eliminated []int, words ...string) aiseat.Input {
	t.Helper()
	const choiceID = "choice-direction"
	seats := make([]protocol.PlayerView, 4)
	var cards []protocol.CardView
	for i := range seats {
		var opts []seatOpt
		for _, e := range eliminated {
			if e == i {
				opts = append(opts, withEliminated())
			}
		}
		seats[i] = newSeat(i, opts...)
		if p := power[i]; p > 0 {
			cards = append(cards, creature("c"+string(rune('a'+i)), i, "Bruiser", p, p))
		}
	}
	options := make([]protocol.PickOptionView, len(words))
	for i, w := range words {
		options[i] = protocol.PickOptionView{Label: w}
	}
	v := newView(seats, withBattlefield(cards...), withChoice(protocol.PendingChoiceView{
		ID:          choiceID,
		Kind:        "option_pick",
		Chooser:     seatID(0).String(),
		Reason:      "Pramikon, Sky Rampart — choose left or right.",
		PickOptions: options,
	}))
	var moves []legal.Move
	for i, w := range words {
		moves = append(moves, choiceMove(t, 0, choiceID, w, map[string]any{"option_index": i}))
	}
	return input(0, v, moves...)
}

func TestTheBotChoosesTheDirectionWhoseAttackerIsWeaker(t *testing.T) {
	cases := []struct {
		name       string
		power      map[int]int
		eliminated []int
		words      []string
		want       string
	}{
		// Under left, seat 3 (the player before the bot) attacks it.
		{"strong player before it: right", map[int]int{3: 8, 1: 1}, nil, []string{"Left", "Right"}, "Right"},
		// Under right, seat 1 (the player after it) attacks it.
		{"strong player after it: left", map[int]int{1: 8, 3: 1}, nil, []string{"Left", "Right"}, "Left"},
		{"the answer does not depend on the order offered", map[int]int{3: 8, 1: 1}, nil, []string{"Right", "Left"}, "Right"},
		// A player who has left the game is skipped: with seat 3 gone the
		// player before the bot is seat 2.
		{"skips an eliminated player", map[int]int{3: 9, 2: 6, 1: 1}, []int{3}, []string{"Left", "Right"}, "Right"},
		{"skips an eliminated player, other way", map[int]int{3: 9, 2: 1, 1: 6}, []int{3}, []string{"Left", "Right"}, "Left"},
		// Nothing separates them: the first answer offered.
		{"an even table keeps the first answer", map[int]int{1: 3, 3: 3}, nil, []string{"Left", "Right"}, "Left"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := directionInput(t, tc.power, tc.eliminated, tc.words...)
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
				t.Errorf("the bot chose %q, want %q", got, tc.want)
			}
		})
	}
}

// With the same two attackers, the bot prefers the direction that leaves
// it the weaker opponent to attack.
func TestTheBotBreaksATieTowardTheWeakerOpponentItMayAttack(t *testing.T) {
	// Seats 1 and 3 have the same board, so the attacker is a wash. Seat 3
	// is on 5 life, and under "right" it is the bot's target.
	in := directionInput(t, map[int]int{1: 3, 3: 3}, nil, "Left", "Right")
	in.View.Seats[3].Life = 5
	// Right makes seat 3 the bot's target.
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Right" {
		t.Errorf("the bot chose %q, want Right so it may attack the player on 5 life", got)
	}
}
