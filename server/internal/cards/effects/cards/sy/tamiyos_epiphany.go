package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tamiyo's Epiphany — Sorcery {3}{U} (War of the Spark, #5520):
//
//	"Scry 4, then draw two cards."
//
// The draw waits for the scry answer (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48db03c4-c847-4901-8ecb-f480ede94a37",
		Name:         "Tamiyo's Epiphany",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 4, Draws: 2},
		OnResolve:    effects.ResolveSteps(effects.Scry{N: 4}, effects.DrawCards{N: 2}),
	})
}
