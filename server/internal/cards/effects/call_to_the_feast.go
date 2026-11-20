package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Call to the Feast — Sorcery {2}{W}{B} (Ixalan, #5532):
//
//	"Create three 1/1 white Vampire creature tokens with lifelink."
//
// Three 1/1 white Vampire tokens with lifelink from the token table
// (the one Elenda's Hierophant makes), under the caster's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30536d1f-8b1a-474f-a508-d3426480a532",
		Name:         "Call to the Feast",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("1/1 white Vampire with lifelink"), N: 3}),
	})
}
