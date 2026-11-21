package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Garruk's Gorehorn — Creature — Beast {4}{G}, 7/3 (Core Set 2021, #5511):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (Silverback Ape's
// shape). The cost, types and body ride game.Card from Scryfall.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8384392-1914-4423-9f49-92f5643d15dc",
		Name:         "Garruk's Gorehorn",
		Completeness: effects.CompletenessFull,
	})
}
