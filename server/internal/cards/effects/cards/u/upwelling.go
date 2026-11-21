package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Upwelling — Enchantment {3}{G}:
//
//	"Players don't lose unspent mana as steps and phases end."
//
// A static over EVERY player's pool (Spec.ManaPool, #2166), derived from
// the battlefield by the step-boundary sweep: it leaving ends it at once,
// and the next boundary empties pools normally. Colourless is kept too.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "420ff0f4-6056-4b40-9a28-fd4c23d6b81c",
		Name:         "Upwelling",
		Completeness: effects.CompletenessFull,
		ManaPool: []game.ManaPoolStatic{
			{Players: game.ManaPoolEachPlayer, Kind: game.ManaPoolKeep},
		},
	})
}
