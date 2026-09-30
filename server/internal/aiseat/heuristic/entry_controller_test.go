package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// entry_controller_test.go is ADR 0102 owner decision 6: "this enters
// under the control of an opponent of your choice" gives a HARMFUL
// permanent (Captive Audience) to the strongest opponent and a HELPFUL
// one (Pendant of Prosperity) to the weakest. Same board, two purposes,
// two answers.

const entryControllerChoiceID = "00000000-0000-4000-8000-00000000e102"

func entryControllerMoves(t *testing.T) []legal.Move {
	t.Helper()
	return []legal.Move{
		choiceMove(t, 0, entryControllerChoiceID, "give to seat 1", map[string]any{"option_index": 0}),
		choiceMove(t, 0, entryControllerChoiceID, "give to seat 2", map[string]any{"option_index": 1}),
	}
}

func withEntryControllerChoice(purpose string) viewOpt {
	return withChoice(protocol.PendingChoiceView{
		ID:         entryControllerChoiceID,
		Kind:       "entry_controller",
		Chooser:    seatID(0).String(),
		FromPlayer: seatID(0).String(),
		Count:      1,
		Reason:     "choose an opponent to control it",
		PickOptions: []protocol.PickOptionView{
			{Label: "seat 1", Player: seatID(1).String()},
			{Label: "seat 2", Player: seatID(2).String()},
		},
		ControlPurpose: purpose,
	})
}

// entryControllerBoard: seat 2 is far ahead of seat 1 — a board of
// fat creatures against nothing.
func entryControllerBoard() viewOpt {
	return withBattlefield(
		creature(cardID(1), 2, "Dragon", 5, 5),
		creature(cardID(2), 2, "Ogre", 4, 4),
		creature(cardID(3), 2, "Troll", 3, 3),
	)
}

func TestAHarmfulPermanentGoesToTheStrongestOpponent(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2)},
		entryControllerBoard(), withEntryControllerChoice("harm"))
	in := input(0, v, entryControllerMoves(t)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "give to seat 2" {
		t.Errorf("a Captive Audience went to %q — it must go to the strongest opponent", got)
	}
}

func TestAHelpfulPermanentGoesToTheWeakestOpponent(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2)},
		entryControllerBoard(), withEntryControllerChoice("benefit"))
	in := input(0, v, entryControllerMoves(t)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "give to seat 1" {
		t.Errorf("a Pendant of Prosperity went to %q — it must go to the weakest opponent", got)
	}
}
