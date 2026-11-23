package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Champion — Creature — Goblin Warrior {R}, 0/1 (Modern Horizons, #5519):
//
//	"Haste
//	 Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)"
//
// Haste and the engine's exalted keyword (CR 702.83), Aven Squire's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "414446e7-44fa-455c-8cf4-8afa170b6586",
		Name:            "Goblin Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", game.KeywordExalted},
	})
}
