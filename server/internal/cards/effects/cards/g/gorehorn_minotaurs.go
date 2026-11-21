package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorehorn Minotaurs — Creature — Minotaur Warrior {2}{R}{R}, 3/3 (Magic
// 2012, #5573):
//
//	"Bloodthirst 2 (If an opponent was dealt damage this turn, this
//	 creature enters with two +1/+1 counters on it.)"
//
// Bloodthirst is the shared entry replacement (bloodthirst.go), the entry
// replacement and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be8ad29c-2232-426d-bda9-a289dcc6c265",
		Name:         "Gorehorn Minotaurs",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Bloodthirst(2, "Gorehorn Minotaurs")},
	})
}
