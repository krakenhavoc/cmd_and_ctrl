package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimmerburst — Instant {3}{U} (Duskmourn: House of Horror, #4466):
//
//	"Draw two cards. Create a 1/1 white Glimmer enchantment creature
//	 token."
//
// Two sentences in printed order (CR 608.2c): the draws, then Grand
// Entryway's Glimmer from the token table. ResolveSteps holds the token
// back while a draw waits on a prompt (a replacement offer).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a26820b9-3bec-41b2-9812-99e4a2211297",
		Name:         "Glimmerburst",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Tokens: 1, CreatureTokens: 1},
		OnResolve: ResolveSteps(
			DrawCards{N: 2},
			CreateToken{Template: TokenCard("1/1 white Glimmer enchantment"), N: 1},
		),
	})
}
