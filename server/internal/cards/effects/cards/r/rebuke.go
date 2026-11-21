package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rebuke — Instant {2}{W} (Innistrad, #5572):
//
//	"Destroy target attacking creature."
//
// Kill Shot's body: the target must be attacking at announce and at
// resolution, so one removed from combat in response is an illegal
// target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf35ec1e-11e2-4028-89df-9755fd6772bf",
		Name:         "Rebuke",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
