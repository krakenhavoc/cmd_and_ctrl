package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Noggle Bandit — Creature — Noggle Rogue {1}{U/R}{U/R}, 2/2 (Eventide, #5590):
//
//	"This creature can't be blocked except by creatures with defender."
//
// Silhana Ledgewalker's block rule with defender as the allowed set: every
// blocker without defender, read through the layers, is refused.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e7707df-54e3-4307-9695-2de4e3edcf89",
		Name:         "Noggle Bandit",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedExceptBy(OnSelf(), HasKeyword("defender"), "creatures with defender"),
		},
	})
}
