package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Ward — Instant {W} (Ice Age, S82 OLD, #5636):
//
//	"Regenerate target creature."
//
// Creates a regeneration shield on the target as the spell resolves (CR 701.19a); it lasts until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0544b707-ec67-43e3-a25d-fd7005ab673d",
		Name:         "Death Ward",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.S82OldB844DeathWard,
	})
}
