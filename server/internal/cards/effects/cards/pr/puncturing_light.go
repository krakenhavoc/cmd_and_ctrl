package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Puncturing Light — Instant {1}{W} (Shadows over Innistrad, S82 OLD b293):
//
//	"Destroy target attacking or blocking creature with power 3 or less."
//
// Protective Response's attacking-or-blocking target (CR 506.4) with
// Vanquish the Weak's power bound, both judged at announce and again as the
// spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15e71226-feb0-4870-a3ea-d8604ca5b3dc",
		Name:         "Puncturing Light",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking or blocking creature with power 3 or less", effects.And(effects.AttackingOrBlocking(), effects.PowerLE(3))),
		OnResolve:    effects.DestroyEachLegalTarget,
	})
}
