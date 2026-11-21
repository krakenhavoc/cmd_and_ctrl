package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hive Stirrings — Sorcery {2}{W} (Magic 2014, #5563):
//
//	"Create two 1/1 colorless Sliver creature tokens."
//
// Thrumming Hivepool's two Sliver tokens as a sorcery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7dd0614-e11e-4b6b-a2cb-9904e457148f",
		Name:         "Hive Stirrings",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Sliver"), N: 2}),
	})
}
