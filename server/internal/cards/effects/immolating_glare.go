package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Immolating Glare — Instant {1}{W} (Oath of the Gatewatch, #5545):
//
//	"Destroy target attacking creature."
//
// Kill Shot's target clause (an attacking creature, judged at announce
// and at resolution) and Bedevil's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "046b6989-5b56-4b52-a655-109b1eb26a93",
		Name:         "Immolating Glare",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
