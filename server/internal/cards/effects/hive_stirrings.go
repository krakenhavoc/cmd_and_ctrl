package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hive Stirrings — Sorcery {2}{W} (Magic 2014, #5563):
//
//	"Create two 1/1 colorless Sliver creature tokens."
//
// Thrumming Hivepool's two Sliver tokens as a sorcery.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7dd0614-e11e-4b6b-a2cb-9904e457148f",
		Name:         "Hive Stirrings",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("1/1 colorless Sliver"), N: 2}),
	})
}
