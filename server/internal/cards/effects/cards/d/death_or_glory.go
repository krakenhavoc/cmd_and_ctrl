package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death or Glory — Sorcery {4}{W} (Invasion, S82 older sets batch 662, #5616):
//
//	"Separate all creature cards in your graveyard into two piles. Exile the
//	 pile of an opponent's choice and return the other to the battlefield."
//
// Split the Spoils' pile split: you separate, the opponent you name takes one
// pile. The taken pile is exiled and the other returns together under your
// control. Either pile may be empty; with no creature cards nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2ae0968-28ed-4ebb-a4e8-69ff6e6a6f1d",
		Name:         "Death or Glory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB662DeathOrGlory(ctx.Game, item)
		},
	})
}
