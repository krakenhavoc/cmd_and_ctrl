package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sworn Companions — Sorcery {2}{W} (Guilds of Ravnica, #5525):
//
//	"Create two 1/1 white Soldier creature tokens with lifelink."
//
// The token table's 1/1 white lifelink Soldier, twice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1a9fd14-fecb-4d2a-983c-70143238275b",
		Name:         "Sworn Companions",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("1/1 white Soldier with lifelink"), N: 2}),
	})
}
