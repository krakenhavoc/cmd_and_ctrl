package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Queen's Commission — Sorcery {2}{W} (Ixalan, #5532):
//
//	"Create two 1/1 white Vampire creature tokens with lifelink."
//
// Two 1/1 white Vampire tokens with lifelink from the token table
// (the one Elenda's Hierophant makes), under the caster's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e4c02c2-2e54-4e8d-8d94-6b2454027e7e",
		Name:         "Queen's Commission",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("1/1 white Vampire with lifelink"), N: 2}),
	})
}
