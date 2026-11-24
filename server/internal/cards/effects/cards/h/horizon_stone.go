package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horizon Stone — Artifact {5} (Commander Legends):
//
//	"If you would lose unspent mana, that mana becomes colorless instead."
//
// Kruphix's mana-pool static (game.ManaPoolBecomesColorless, #2166): at every
// step and phase boundary the controller's unspent mana becomes colorless and
// stays, keeping its restrictions and riders.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "add46292-470e-47fa-a261-4884d02e65fe",
		Name:         "Horizon Stone",
		Completeness: effects.CompletenessFull,
		ManaPool:     []game.ManaPoolStatic{{Kind: game.ManaPoolBecomesColorless}},
	})
}
