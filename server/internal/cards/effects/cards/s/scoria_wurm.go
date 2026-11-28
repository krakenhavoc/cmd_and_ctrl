package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scoria Wurm — Creature — Wurm {4}{R}, 7/7 (Urza's Saga, S82 OLD batch 731,
// #5624):
//
//	"At the beginning of your upkeep, flip a coin. If you lose the flip,
//	 return this creature to its owner's hand."
//
// The controller calls the flip (CR 705.2). On a lost flip the Wurm returns,
// but only the object that triggered: one that has left and come back since is
// a new object (CR 400.7) and stays.
//
// No simplification.
func init() {
	scoriaWurmFlipped = effects.CoinFlipThen("coin-flip/scoria-wurm-upkeep", effects.ReturnThisToHandUnlessYouWonTheFlip)
	effects.Register(effects.Spec{
		OracleID:     "ee45c80c-699c-4e37-97e6-1859c6601c99",
		Name:         "Scoria Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Scoria Wurm — flip a coin; if you lose the flip, return this creature to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Scoria Wurm — call the coin flip",
						ThenKey:  scoriaWurmFlipped,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// scoriaWurmFlipped is "if you lose the flip, return this creature to its
// owner's hand", keyed. Set in init, with the card. An on-disk key.
var scoriaWurmFlipped game.CoinFlipThen
