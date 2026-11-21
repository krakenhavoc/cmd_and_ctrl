package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adventurous Impulse — Sorcery {G} (Foundations Jumpstart, #5929):
//
//	"Look at the top three cards of your library. You may reveal a creature
//	 or land card from among them and put it into your hand. Put the rest on
//	 the bottom of your library in any order."
//
// Commune with Nature's dig (TakeFromLibraryToHand with
// takeRevealRestBottomAny) over three cards and creature or land cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bf49f39-b3d5-4148-8dc2-b5b5011769d2",
		Name:         "Adventurous Impulse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Match:    effects.Or(effects.Creature(), effects.Land()),
				Max:      1,
				Optional: true,
				Label:    "Adventurous Impulse — reveal a creature or land card and put it into your hand",
				Key:      effects.TakeRevealRestBottomAny,
			}.Apply(ctx)
		},
	})
}
