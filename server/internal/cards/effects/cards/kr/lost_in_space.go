package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lost in Space — Instant {3}{U} (Edge of Eternities, #4461):
//
//	"Target artifact or creature's owner puts it on their choice of the
//	 top or bottom of their library. Surveil 1."
//
// Vanish from Sight's two steps over an artifact-or-creature target: the
// owner chooses the end of their library (ownerTucksFirstTargetTopOrBottom),
// then the caster surveils 1. ResolveSteps runs the surveil only once the
// owner has answered (CR 608.2c). A target gone by resolution leaves the
// spell without a legal target, so it does nothing, surveil included
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cec95b9f-9d1e-4988-befe-d0c9e20e428d",
		Name:         "Lost in Space",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.OwnerTucksFirstTargetTopOrBottom(ctx, "Lost in Space — put it on your choice of the top or bottom of your library")
			}),
			effects.Surveil{N: 1},
		),
	})
}
