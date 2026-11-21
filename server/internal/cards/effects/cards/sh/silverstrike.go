package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silverstrike — Instant {3}{W} (Shadows over Innistrad, #5543):
//
//	"Destroy target attacking creature. You gain 3 life."
//
// Radiant Strike's destroy-then-gain-life (destroyTargetThenGainLife) on
// an attacking creature (attackingCreature, CR 506.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8652fa06-33dc-4fe5-90cb-83cbfe5c681b",
		Name:         "Silverstrike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking creature", effects.IsAttackingCreature()),
		OnResolve:    effects.DestroyTargetThenGainLife(3),
	})
}
