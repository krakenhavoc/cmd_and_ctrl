package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nagging Thoughts — Sorcery {1}{U} (Shadows over Innistrad, S82 OLD):
//
//	"Look at the top two cards of your library. Put one of them into your hand
//	 and the other into your graveyard.
//	 Madness {1}{U}"
//
// Waker of Waves' look-at-two (private to the looker), with madness declared
// once (Spec.Madness), which grows both halves of the keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e74c70be-2b70-4b04-9909-161d6f41a156",
		Name:         "Nagging Thoughts",
		Completeness: effects.CompletenessFull,
		Madness:      "{1}{U}",
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 2),
				Max:    1,
				Label:  "Nagging Thoughts — put one of them into your hand and the other into your graveyard",
				Key:    effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
