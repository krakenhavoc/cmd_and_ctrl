package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tibalt, Rakish Instigator — Legendary Planeswalker — Tibalt {2}{R},
// loyalty 5:
//
//	"Your opponents can't gain life.
//	 −2: Create a 1/1 red Devil creature token with "When this token
//	 dies, it deals 1 damage to any target.""
//
// "Your opponents can't gain life" is ADR 0107 §5's battlefield static
// (CR 119.7, #1880). The −2 makes the catalog's Devil (Spiked
// Corridor's, devilToken()), whose dies trigger is the token's own.
// Starting loyalty comes from the deck import (ADR 0032).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b768f8f-2213-45b9-bced-8fb1bbb441c3",
		Name:         "Tibalt, Rakish Instigator",
		Completeness: effects.CompletenessFull,
		CantGainLife: effects.OpponentsCantGainLife(),
		Activated: []effects.ActivatedAbility{{
			Label: "−2: Create a 1/1 red Devil creature token with \"When this token dies, it deals 1 damage to any target.\"",
			Cost:  effects.LoyaltyCost(-2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.DevilToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
