package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morningtide — Sorcery {1}{W} (Torment, #5612):
//
//	"Exile all graveyards."
//
// Scavenger Grounds' ability as a sorcery: every seat's graveyard, the
// caster's included, through the shared b02ExileAllGraveyards.
// Morningtide is still on the stack as it resolves, so it goes to its
// owner's graveyard afterwards and is not exiled with the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8432372-672b-4bf7-9000-17e43356f149",
		Name:         "Morningtide",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerGraveyard},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B02ExileAllGraveyards(ctx.Game, item)
		},
	})
}
