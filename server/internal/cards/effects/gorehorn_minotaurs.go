package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "be8ad29c-2232-426d-bda9-a289dcc6c265",
		Name:         "Gorehorn Minotaurs",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Bloodthirst(2, "Gorehorn Minotaurs")},
	})
}
