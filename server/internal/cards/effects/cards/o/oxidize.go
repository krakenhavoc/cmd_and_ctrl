package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oxidize — Instant {G} (Darksteel, #5605):
//
//	"Destroy target artifact. It can't be regenerated."
//
// Terror's no-regeneration destroy on an artifact target (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ff78353-a26e-4b5a-948d-1c3b41d2dd1f",
		Name:         "Oxidize",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}
