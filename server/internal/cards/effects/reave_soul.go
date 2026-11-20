package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reave Soul — Sorcery {1}{B} (Magic Origins, #5549):
//
//	"Destroy target creature with power 3 or less."
//
// Vanquish the Weak's shape: the power bound is checked at announce and
// at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8eb328e6-332e-435c-9733-4c5336f17824",
		Name:         "Reave Soul",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with power 3 or less", PowerLE(3)),
		OnResolve:    destroyEachLegalTarget,
	})
}
