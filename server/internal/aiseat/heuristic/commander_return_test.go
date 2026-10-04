package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// commander_return_test.go is ADR 0115 owner decision 2: a bot sends
// its commander to the command zone (CR 903.9a) unless it could cast
// it from the graveyard or exile it went to.

const commanderReturnChoiceID = "00000000-0000-4000-8000-000000002115"

func commanderReturnDecision(t *testing.T, playable bool) string {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withTurn(3, 1, "precombat_main"),
		withChoice(protocol.PendingChoiceView{
			ID:               commanderReturnChoiceID,
			Kind:             "commander_return",
			Chooser:          seatID(0).String(),
			FromPlayer:       seatID(0).String(),
			Count:            1,
			Source:           cardID(9),
			Reason:           "Test Commander — put it into the command zone?",
			PlayableFromZone: playable,
		}))
	in := input(0, v, []legal.Move{
		choiceMove(t, 0, commanderReturnChoiceID, "yes", map[string]any{"apply": true}),
		choiceMove(t, 0, commanderReturnChoiceID, "no", map[string]any{"apply": false}),
	}...)
	return chose(t, in, decide(t, heuristic.New(), in))
}

func TestCommanderReturnSendsTheCommanderHome(t *testing.T) {
	if got := commanderReturnDecision(t, false); got != "yes" {
		t.Errorf("a commander it cannot cast where it is: chose %q, want yes", got)
	}
}

func TestCommanderReturnKeepsACommanderCastableWhereItIs(t *testing.T) {
	if got := commanderReturnDecision(t, true); got != "no" {
		t.Errorf("a commander castable from its zone: chose %q, want no", got)
	}
}
