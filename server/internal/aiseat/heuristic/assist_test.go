package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// assist_test.go — ADR 0150 §6: a bot asked to help pay for another
// player's spell with assist answers 0, however much it could pay.
func TestABotHelperAnswersZero(t *testing.T) {
	pol := heuristic.New()
	const id = "44444444-4444-4444-4444-444444444444"
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withChoice(protocol.PendingChoiceView{
		ID: id, Kind: "pay_amount", Chooser: seatID(1).String(),
		PayAmount: &protocol.PayAmountView{Min: 0, Max: 3, Unit: "assist", Resource: "mana"},
		Assist:    &protocol.AssistPromptView{Caster: seatID(0).String(), SpellName: "Fan Favorite", Generic: 3},
	}))
	in := input(1, v,
		choiceMove(t, 1, id, "none", map[string]any{"amount": 0}),
		choiceMove(t, 1, id, "all", map[string]any{"amount": 3}))
	none, all := rankValue(t, pol, in, "none"), rankValue(t, pol, in, "all")
	if none <= all {
		t.Errorf("help nothing %.3f, help {3} %.3f; a bot helper must answer 0", none, all)
	}
}
