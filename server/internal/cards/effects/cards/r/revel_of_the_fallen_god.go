package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revel of the Fallen God — Sorcery {3}{R}{R}{G}{G} (Journey into Nyx, S82
// OLD, #5558):
//
//	"Create four 2/2 red and green Satyr creature tokens with haste."
//
// Xenagos, the Reveler's Satyr token, four of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d46b8dd-2a4a-4d9f-b4c6-2b45f05081c2",
		Name:         "Revel of the Fallen God",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4, CreatureTokens: 4},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("2/2 red and green Satyr with haste"), N: 4}.Apply(ctx)
		},
	})
}
