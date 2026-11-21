package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eternal Isolation — Sorcery {1}{W} (Core Set 2020, #5518):
//
//	"Put target creature with power 4 or greater on the bottom of its owner's
//	 library."
//
// The power is the creature's current power, judged at cast and again as
// it resolves (CR 608.2b); the tuck is the shared bottom-of-library body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f69a103-d557-4d28-847e-aa94a440e908",
		Name:         "Eternal Isolation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82One13TuckTargetToLibraryBottom(ctx.Game, item)
		},
	})
}
