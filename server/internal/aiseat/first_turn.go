package aiseat

import (
	"encoding/json"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// first_turn.go is ADR 0125 §5.2: a practice bot that wins the opening
// roll hands the first turn to the player.
//
// The practice table rolls for the first turn like any other table
// (ADR 0121 §1), so the tutorial's player sees the dice, the banner and
// the choice they will meet at their first real table. The tutorial's
// steps follow the player's own first turn, so when the bot wins, the
// bot chooses the player. CR 103.1 lets the winner choose any player
// ("the players determine which one of them will choose who takes the
// first turn"), so this is a legal choice made through the ordinary
// choose_starting_player move, not a rigged roll: the dice are the
// engine's, and the engine knows nothing of it.
//
// It is ONE window of one table type. Config.FirstTurnTo is set only
// from SeatSpec.FirstTurnTo, which only the practice route sets. Every
// other window, rolling included, goes to the seat's policy unchanged,
// and a bot at any other table chooses as ADR 0121 §4 has it. It is
// not a tutorial tier, which ADR 0076 §2.2 rejected.

// RuleFirstTurnTo names the runner's answer in a decision trace (Layer
// "A": a fixed answer, not a judgement about a position).
const RuleFirstTurnTo = "first-turn-to"

// FirstTurnIndex returns the index of the move that hands the first
// turn to seat, or -1. It answers only a window whose moves are ALL
// choose_starting_player, which is the winner's choice of ADR 0121 §2
// and nothing else; a die to roll, a mulligan or any later window is
// -1. A window that does not offer seat (an eliminated seat is not
// offered) is -1 too, and the seat's policy answers it.
func FirstTurnIndex(moves []legal.Move, seat int) int {
	if len(moves) == 0 {
		return -1
	}
	found := -1
	for i := range moves {
		m := &moves[i]
		if m.Type != legal.TypeChooseStartingPlayer {
			return -1
		}
		if found >= 0 {
			continue
		}
		var p struct {
			Seat *int `json:"seat"`
		}
		if json.Unmarshal(m.Params, &p) == nil && p.Seat != nil && *p.Seat == seat {
			found = i
		}
	}
	return found
}
