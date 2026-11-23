package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Migratory Route — Sorcery {3}{W}{U}:
//
//	"Create four 1/1 white Bird creature tokens with flying.
//	 Basic landcycling {2}"
//
// Mass Production's four tokens, and Apex Witchstalker's landcycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6a74582-62b6-48d1-89d6-e26059229840",
		Name:         "Migratory Route",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4, CreatureTokens: 4},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: 4}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
