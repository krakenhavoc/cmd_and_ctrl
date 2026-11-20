package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Rally — Sorcery {3}{R}{R} (Return to Ravnica, #5566):
//
//	"Create four 1/1 red Goblin creature tokens."
//
// Dragon Fodder at four: the table's 1/1 red Goblin row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ccb6ce8-ccb6-4809-a1cc-cebd38744815",
		Name:         "Goblin Rally",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4, CreatureTokens: 4},
		OnResolve:    ResolveSteps(CreateToken{Template: RedGoblinToken(), N: 4}),
	})
}
