package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Scouts — Sorcery {3}{R}{R} (Mirage, S82 OLD b799, #5633):
//
//	"Create three 1/1 red Goblin Scout creature tokens with mountainwalk.
//	 (They can't be blocked as long as defending player controls a Mountain.)"
//
// Three tokens made together, each a 1/1 red Goblin Scout with mountainwalk.
// The card never had a printed token card (tokens start in 2007), so the
// template follows the oracle text and is a declared unmatched token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87b55d99-badc-45e3-8364-173a400c4d7e",
		Name:         "Goblin Scouts",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin Scout with mountainwalk"), N: 3}),
	})
}
