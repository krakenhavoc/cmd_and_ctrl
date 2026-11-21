package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sultai Soothsayer — Creature — Snake Shaman {2}{B}{G}{U}, 2/5:
//
//	"When this creature enters, look at the top four cards of your
//	 library. Put one of them into your hand and the rest into your
//	 graveyard."
//
// Confounding Riddle's look: a private look at the top four, a mandatory
// pick of one for the hand, and the rest put into the graveyard (not
// milled). A shorter library offers what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fe127e0-ee38-4f80-969b-13e84712a26f",
		Name:         "Sultai Soothsayer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sultai Soothsayer — look at the top four cards, one to your hand and the rest to your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 4),
						Max:    1,
						Label:  "Sultai Soothsayer — put one of them into your hand and the rest into your graveyard",
						Key:    effects.TakeRestGraveyard,
					}.Apply(ctx)
				}),
		},
	})
}
