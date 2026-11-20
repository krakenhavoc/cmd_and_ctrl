package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silverstrike — Instant {3}{W} (Shadows over Innistrad, #5543):
//
//	"Destroy target attacking creature. You gain 3 life."
//
// Radiant Strike's destroy-then-gain-life (destroyTargetThenGainLife) on
// an attacking creature (attackingCreature, CR 506.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8652fa06-33dc-4fe5-90cb-83cbfe5c681b",
		Name:         "Silverstrike",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking creature", attackingCreature()),
		OnResolve:    destroyTargetThenGainLife(3),
	})
}
