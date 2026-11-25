package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Braid of Fire — Enchantment {1}{R} (Secret Lair Drop):
//
//	"Cumulative upkeep—Add {R}."
//
// The upkeep cost is adding the mana, once per age counter (CR 702.24a):
// "yes" puts that much red mana in your pool during the upkeep and "no"
// sacrifices the enchantment. OldB008AddManaCumulativeUpkeep is the shared
// payment shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "796009b3-73b5-42c0-a490-66f12be5a42e",
		Name:         "Braid of Fire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB008AddManaCumulativeUpkeep("Braid of Fire — cumulative upkeep: add {R}", "R"),
		},
	})
}
