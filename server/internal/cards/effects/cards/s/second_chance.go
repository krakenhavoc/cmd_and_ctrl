package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Second Chance — Enchantment {2}{U} (Urza's Legacy, S82 OLD b716, #5623):
//
//	"At the beginning of your upkeep, if you have 5 or less life, sacrifice this enchantment and take an extra turn after this one."
//
// An intervening "if" (CR 603.4): asked as the upkeep begins and again as the
// ability resolves. The extra turn is taken whether or not the sacrifice
// worked (it is a second instruction of the same sentence, CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa19bbf0-7de3-4ec3-a98b-2ac7533c6b0b",
		Name:         "Second Chance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ExtraTurns: 1},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you have 5 or less life",
				func(g *game.Game, you uuid.UUID) bool {
					p := g.PlayerByIDForEffect(you)
					return p != nil && p.Life <= 5
				},
				effects.AtYourUpkeep("Second Chance — sacrifice this enchantment and take an extra turn after this one",
					func(g *game.Game, item *game.StackItem) error {
						if err := effects.SacrificeThisIfStillOnBattlefield(g, item); err != nil {
							return err
						}
						return effects.TakeExtraTurn{}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}
