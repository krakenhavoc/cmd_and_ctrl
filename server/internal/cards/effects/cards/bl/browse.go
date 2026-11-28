package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Browse — Enchantment {2}{U}{U} (Alliances, #5634):
//
//	"{2}{U}{U}: Look at the top five cards of your library, put one of them
//	 into your hand, and exile the rest."
//
// A look (private to you) at the top five and a mandatory pick of one for the
// hand; the rest are exiled face up (the exile is public). Ashiok's keyed take
// is the same sentence: one into hand, the others exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49491601-e272-4603-9348-185a2953bd4b",
		Name:         "Browse",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{U}: Look at the top five cards of your library, put one of them into your hand, and exile the rest.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{U}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, 5),
					Max:    1,
					Label:  "Browse — put one of them into your hand and exile the rest",
					Key:    effects.AshiokTakeRestIntoExile,
				}.Apply(ctx)
			},
		}},
	})
}
