package aiseat

import (
	"encoding/json"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// OpeningRollIndex is a bot's answer to the opening roll (ADR 0121 §4),
// shared by Layer A (aiseat/rules) and the heuristic so the two can
// never disagree about it. It returns the move to make and a reason,
// or -1 when the window is not the opening roll's.
//
// A seat that owes a die rolls it: the window offers nothing else. A
// seat that won chooses to take the first turn itself. Going first is
// the usual choice in Commander, and nothing a bot can see before the
// deal (no hand exists yet) could argue for handing it away. Neither
// is a judgement about a position, which is why Layer A answers both
// and a model tier never spends a call on either.
//
// The chooser's own seat is read off the view (in.Seat's PlayerView).
// Without a view — a test, a harness — it falls back to the choice the
// enumerator marked AlwaysLegal, which is the chooser's own (legal's
// openingRollMoves).
func OpeningRollIndex(in Input) (int, string) {
	if len(in.Moves) == 0 {
		return -1, ""
	}
	roll := -1
	for i := range in.Moves {
		if in.Moves[i].Kind != legal.KindOpeningRoll {
			return -1, ""
		}
		if in.Moves[i].Type == legal.TypeRollOpening && roll < 0 {
			roll = i
		}
	}
	if roll >= 0 {
		return roll, "roll for the first turn"
	}
	self := -1
	for _, s := range in.View.Seats {
		if s.ID == in.Seat.String() {
			self = s.Seat
			break
		}
	}
	fallback := -1
	for i := range in.Moves {
		m := &in.Moves[i]
		if m.Type != legal.TypeChooseStartingPlayer {
			continue
		}
		if self >= 0 {
			var p struct {
				Seat *int `json:"seat"`
			}
			if json.Unmarshal(m.Params, &p) == nil && p.Seat != nil && *p.Seat == self {
				return i, "take the first turn"
			}
		}
		if m.AlwaysLegal && fallback < 0 {
			fallback = i
		}
	}
	if fallback >= 0 {
		return fallback, "take the first turn"
	}
	return -1, ""
}
