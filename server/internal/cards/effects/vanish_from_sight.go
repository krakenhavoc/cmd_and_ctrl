package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanish from Sight — Instant {3}{U} (Duskmourn: House of Horror,
// #4466):
//
//	"Target nonland permanent's owner puts it on their choice of the top
//	 or bottom of their library. Surveil 1."
//
// Trip Up's placement: the choice is the OWNER's, not the caster's
// (ownerTucksFirstTargetTopOrBottom). The surveil is the caster's and is
// its own step, so it waits for the owner's answer (CR 608.2c): a
// permanent its owner put on top of the caster's own library is the card
// the caster then looks at. A target that is gone fizzles the spell, the
// surveil included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1056b27c-125c-4daa-a0c8-533efc9ccc32",
		Name:         "Vanish from Sight",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ownerTucksFirstTargetTopOrBottom(ctx, "Vanish from Sight — put it on your choice of the top or bottom of your library")
			}),
			Surveil{N: 1},
		),
	})
}
