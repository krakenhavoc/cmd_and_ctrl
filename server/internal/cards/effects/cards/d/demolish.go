package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demolish — Sorcery {3}{R} (Ravnica Remastered, #5948):
//
//	"Destroy target artifact or land."
//
// One target clause and the shared "Destroy target [permanent]." body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08c6d2fd-d9ae-4f95-bb9d-7e16b1039814",
		Name:         "Demolish",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
