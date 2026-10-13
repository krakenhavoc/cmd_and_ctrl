package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// villainous_choice_test.go — #3140: the bot facing a villainous choice
// (CR 701.55) takes the cheaper of two bad options.

func villainousInput(t *testing.T, life int, options ...protocol.PickOptionView) aiseat.Input {
	t.Helper()
	const choiceID = "choice-villain"
	v := newView([]protocol.PlayerView{newSeat(0, withLife(life)), newSeat(1)},
		withChoice(protocol.PendingChoiceView{
			ID:          choiceID,
			Kind:        "option_pick",
			Chooser:     seatID(0).String(),
			Reason:      "Damocles Base — sacrifice a nontoken creature, or lose 2 life?",
			PickOptions: options,
		}))
	var moves []legal.Move
	for i, o := range options {
		m := choiceMove(t, 0, choiceID, o.Label, map[string]any{"option_index": i})
		if o.LifeCost > 0 {
			m.Cost = &legal.MoveCost{Life: o.LifeCost}
		}
		moves = append(moves, m)
	}
	return input(0, v, moves...)
}

func TestTheBotTakesTheCheaperOfTwoBadOptions(t *testing.T) {
	in := villainousInput(t, 40,
		protocol.PickOptionView{Label: "Sacrifice a nontoken creature", Harm: 2},
		protocol.PickOptionView{Label: "Lose 2 life", LifeCost: 2, Harm: 1},
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Lose 2 life" {
		t.Errorf("the bot chose %q, want the cheaper loss", got)
	}
}

// Life is cheap at 40 and precious at 5: the same prompt flips.
func TestTheBotPricesLifeByHowCloseToDeathItIs(t *testing.T) {
	in := villainousInput(t, 4,
		protocol.PickOptionView{Label: "Sacrifice a nontoken creature", Harm: 2},
		protocol.PickOptionView{Label: "Lose 3 life", LifeCost: 3, Harm: 1},
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Sacrifice a nontoken creature" {
		t.Errorf("the bot chose %q at 4 life, want to keep it", got)
	}
}

func TestTheBotNeverTakesItsLastLife(t *testing.T) {
	in := villainousInput(t, 5,
		protocol.PickOptionView{Label: "Lose 5 life", LifeCost: 5},
		protocol.PickOptionView{Label: "Shuffle another creature into your library", Harm: 2},
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Shuffle another creature into your library" {
		t.Errorf("the bot chose %q, which kills it", got)
	}
}

// An option that would do nothing arrives priced at zero, and is taken.
func TestTheBotTakesAnOptionThatWouldDoNothingToIt(t *testing.T) {
	in := villainousInput(t, 40,
		protocol.PickOptionView{Label: "Lose 2 life", LifeCost: 2, Harm: 1},
		protocol.PickOptionView{Label: "Sacrifice a nontoken creature"}, // no creature: Harm zeroed
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Sacrifice a nontoken creature" {
		t.Errorf("the bot chose %q, want the free one", got)
	}
}

// A prompt with no hint at all keeps the enumerator's first answer, as
// any other option pick does.
func TestAnUnhintedOptionPickIsUnchanged(t *testing.T) {
	in := villainousInput(t, 40,
		protocol.PickOptionView{Label: "First"},
		protocol.PickOptionView{Label: "Second"},
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "First" {
		t.Errorf("the bot chose %q, want the first", got)
	}
}
