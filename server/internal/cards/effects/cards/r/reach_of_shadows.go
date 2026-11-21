package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reach of Shadows — Instant {4}{B} (Fate Reforged,
// #5552):
//
//	"Destroy target creature that's one or more colors."
//
// Doom Blade's destroy with the clause "one or more colors": a colorless
// creature (an artifact creature, a face-down 2/2) is not a legal target,
// read as it is targeted and again as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e01382e0-a685-41c4-9ab6-727ffdbf2e0a",
		Name:         "Reach of Shadows",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature that's one or more colors", effects.Not(effects.Colorless())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
