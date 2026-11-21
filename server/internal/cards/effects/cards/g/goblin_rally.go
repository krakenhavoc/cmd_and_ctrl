package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Rally — Sorcery {3}{R}{R} (Return to Ravnica, #5566):
//
//	"Create four 1/1 red Goblin creature tokens."
//
// Dragon Fodder at four: the table's 1/1 red Goblin row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ccb6ce8-ccb6-4809-a1cc-cebd38744815",
		Name:         "Goblin Rally",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4, CreatureTokens: 4},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.RedGoblinToken(), N: 4}),
	})
}
