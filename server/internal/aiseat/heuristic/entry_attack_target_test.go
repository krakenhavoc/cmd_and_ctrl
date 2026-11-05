package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// entry_attack_target_test.go — CR 508.4's "what does it attack?" for a
// creature entering tapped and attacking (Kira Nerys, Firebrand). The bot
// sends it at the opponent with the least life, and with nothing to tell
// the defenders apart keeps it with the attack it already made (the first
// option).

const entryAttackChoiceID = "00000000-0000-4000-8000-00000000e508"

func entryAttackMoves(t *testing.T) []legal.Move {
	t.Helper()
	return []legal.Move{
		choiceMove(t, 0, entryAttackChoiceID, "attack seat 1", map[string]any{"option_index": 0}),
		choiceMove(t, 0, entryAttackChoiceID, "attack seat 2", map[string]any{"option_index": 1}),
	}
}

func withEntryAttackChoice() viewOpt {
	return withChoice(protocol.PendingChoiceView{
		ID:         entryAttackChoiceID,
		Kind:       "entry_attack_target",
		Chooser:    seatID(0).String(),
		FromPlayer: seatID(0).String(),
		Count:      1,
		Reason:     "Kira Nerys, Firebrand — what does it attack?",
		PickOptions: []protocol.PickOptionView{
			{Label: "Attack Seat1", Player: seatID(1).String()},
			{Label: "Attack Seat2", Player: seatID(2).String()},
		},
	})
}

func TestAnEnteringAttackerGoesAtTheOpponentWithTheLeastLife(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2, withLife(6))}, withEntryAttackChoice())
	in := input(0, v, entryAttackMoves(t)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "attack seat 2" {
		t.Errorf("the entering attacker went to %q — it must go at the opponent at 6 life", got)
	}
}

func TestAnEnteringAttackerKeepsToTheAttackOnATie(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2)}, withEntryAttackChoice())
	in := input(0, v, entryAttackMoves(t)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "attack seat 1" {
		t.Errorf("the entering attacker went to %q — on a tie it keeps to the first offered defender", got)
	}
}
