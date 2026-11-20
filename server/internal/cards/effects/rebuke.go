package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cf35ec1e-11e2-4028-89df-9755fd6772bf",
		Name:         "Rebuke",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
