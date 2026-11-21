package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stark Industries Executive — {R} Creature — Human Advisor 1/2 (Marvel
// Super Heroes, #4455):
//
//	"{2}, {T}: Create a Treasure token."
//
// A plain activated token maker (Warren Soultrader's Treasure without the
// sacrifice); the Treasure is the catalog's own TreasureToken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c07970d-063a-44a1-9761-b80aafa90dac",
		Name:         "Stark Industries Executive",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
