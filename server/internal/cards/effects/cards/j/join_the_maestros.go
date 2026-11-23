package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "849a7ae8-2b4e-4fa5-9607-2bfc961fee40",
		Name:          "Join the Maestros",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(2)},
		Purpose:       game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("4/3 black Ogre Warrior"), N: 1}.Apply(ctx)
		},
	})
}
