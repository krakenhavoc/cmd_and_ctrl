package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// costed_options_test.go — #2854: an option_pick whose options cost
// mana. The bot pays, and pays the least that buys a way out.

func costedPickInput(t *testing.T, options ...protocol.PickOptionView) aiseat.Input {
	t.Helper()
	const choiceID = "choice-chill"
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withChoice(protocol.PendingChoiceView{
			ID:          choiceID,
			Kind:        "option_pick",
			Chooser:     seatID(0).String(),
			Reason:      "Winter's Chill — pay {1} or {2} for Bear?",
			PickOptions: options,
		}))
	var moves []legal.Move
	for i, o := range options {
		m := choiceMove(t, 0, choiceID, o.Label, map[string]any{"option_index": i})
		if o.ManaCost != "" {
			m.Cost = &legal.MoveCost{Mana: o.ManaCost}
		}
		moves = append(moves, m)
	}
	return input(0, v, moves...)
}

func TestHeuristicPaysTheCheapestCostedOption(t *testing.T) {
	in := costedPickInput(t,
		protocol.PickOptionView{Label: "Pay nothing"},
		protocol.PickOptionView{Label: "Pay {1}", ManaCost: "{1}"},
		protocol.PickOptionView{Label: "Pay {2}", ManaCost: "{2}"},
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pay {1}" {
		t.Errorf("the bot chose %q, want the cheapest payment", got)
	}
}

func TestHeuristicPaysRatherThanTakingThePenalty(t *testing.T) {
	in := costedPickInput(t,
		protocol.PickOptionView{Label: "Pay nothing: take 1 damage"},
		protocol.PickOptionView{Label: "Pay {3}", ManaCost: "{3}"},
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pay {3}" {
		t.Errorf("the bot chose %q, want to pay", got)
	}
}
