package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verdigris — Instant {2}{G} (Tempest, S82 OLD, #5628):
//
//	"Destroy target artifact."
//
// One target clause and the shared "Destroy target [permanent]." body, as
// Shatter has it. Artifact() reads the post-layer type set.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb41ea1d-3491-4193-a390-8eb202fbae12",
		Name:         "Verdigris",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
