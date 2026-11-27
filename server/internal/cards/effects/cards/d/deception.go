package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deception — Sorcery {2}{B}
// (Portal Three Kingdoms, #5622):
//
//	"Target opponent discards two cards."
//
// The discarding player chooses the cards (CR 701.9a); the body is Mind
// Rot's, shared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae1b44df-1d8c-453a-9efb-6d1c93e52418",
		Name:         "Deception",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.MindRotOnResolve,
	})
}
