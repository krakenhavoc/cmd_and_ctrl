package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0b768f8f-2213-45b9-bced-8fb1bbb441c3",
		Name:         "Tibalt, Rakish Instigator",
		Completeness: CompletenessFull,
		CantGainLife: OpponentsCantGainLife(),
		Activated: []ActivatedAbility{{
			Label: "−2: Create a 1/1 red Devil creature token with \"When this token dies, it deals 1 damage to any target.\"",
			Cost:  LoyaltyCost(-2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: devilToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
