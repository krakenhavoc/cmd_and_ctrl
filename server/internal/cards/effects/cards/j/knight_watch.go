package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight Watch — Sorcery {4}{W} (Gatecrash, S82 OLD #5565):
//
//	"Create two 2/2 white Knight creature tokens with vigilance."
//
// The shared "2/2 white Knight with vigilance" token row, twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4965f5fe-a5fe-4976-822c-5cf40fea8c34",
		Name:         "Knight Watch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 2}),
	})
}
