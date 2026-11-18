package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "849a7ae8-2b4e-4fa5-9607-2bfc961fee40",
		Name:          "Join the Maestros",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(2)},
		Purpose:       game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("4/3 black Ogre Warrior"), N: 1}.Apply(ctx)
		},
	})
}
