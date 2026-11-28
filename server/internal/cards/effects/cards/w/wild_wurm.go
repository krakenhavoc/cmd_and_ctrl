package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Wurm — Creature — Wurm {3}{R}, 5/4 (Tempest, S82 OLD batch 768, #5628):
//
//	"When this creature enters, flip a coin. If you lose the flip, return this
//	 creature to its owner's hand."
//
// Scoria Wurm's flip on an enters trigger: the controller calls it, and the
// Wurm goes back to hand only when the call is lost and it is still the object
// that triggered (CR 400.7).
//
// No simplification.
func init() {
	wildWurmFlipped = effects.CoinFlipThen("coin-flip/wild-wurm-enters", effects.ReturnThisToHandUnlessYouWonTheFlip)
	effects.Register(effects.Spec{
		OracleID:     "ba14d185-9432-4a67-8bfb-544c2cebe71d",
		Name:         "Wild Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wild Wurm — flip a coin; if you lose the flip, return this creature to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Wild Wurm — call the coin flip",
						ThenKey:  wildWurmFlipped,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

var wildWurmFlipped game.CoinFlipThen
