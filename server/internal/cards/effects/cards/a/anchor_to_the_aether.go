package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anchor to the Aether — Sorcery {2}{U} (Jumpstart 2022, #5948):
//
//	"Put target creature on top of its owner's library. Scry 1."
//
// Banishment Decree's tuck, then the controller scries 1. A target that
// has become illegal fizzles the whole spell, so no scry follows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e00326d-2bf4-4731-a6cd-1c015f94f553",
		Name:         "Anchor to the Aether",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.PutTargetOnTopOfOwnersLibrary(item, ctx); err != nil {
				return err
			}
			return effects.Scry{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
