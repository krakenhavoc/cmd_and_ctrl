package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cec95b9f-9d1e-4988-befe-d0c9e20e428d",
		Name:         "Lost in Space",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ownerTucksFirstTargetTopOrBottom(ctx, "Lost in Space — put it on your choice of the top or bottom of your library")
			}),
			Surveil{N: 1},
		),
	})
}
