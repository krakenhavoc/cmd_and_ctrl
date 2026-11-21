package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uncharted Voyage — Instant {3}{U} (Foundations, #4465):
//
//	"Target creature's owner puts it on their choice of the top or bottom
//	 of their library.
//	 Surveil 1."
//
// Lost in Space's two steps over a creature target: the owner chooses the
// end of their library (ownerTucksFirstTargetTopOrBottom), then the caster
// surveils 1. ResolveSteps runs the surveil only once the owner has
// answered (CR 608.2c). A target gone by resolution leaves the spell
// without a legal target, so it does nothing, surveil included
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d6796ce-9316-430a-8126-8b0414cf181c",
		Name:         "Uncharted Voyage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.OwnerTucksFirstTargetTopOrBottom(ctx, "Uncharted Voyage — put it on your choice of the top or bottom of your library")
			}),
			effects.Surveil{N: 1},
		),
	})
}
