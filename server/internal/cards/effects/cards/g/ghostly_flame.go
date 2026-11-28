package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghostly Flame — Enchantment {B}{R} (Ice Age, S82 OLD batch 848, #5636):
//
//	"Black and/or red permanents and spells are colorless sources of damage."
//
// The #6230 world rule (ColorlessDamageSources "BR"), carried by the Flame
// itself: as a source of damage, a black or red permanent or spell has no color
// (CR 105.2c), for protection, prevention and "a red source of your choice";
// it keeps its colors for everything else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a461ca3-7923-4df4-9f15-4072545da117",
		Name:         "Ghostly Flame",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.ColorlessDamageSources("BR", "Black and/or red permanents and spells are colorless sources of damage."),
		},
	})
}
