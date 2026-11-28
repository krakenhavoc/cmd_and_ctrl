package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Conquer — Enchantment — Aura {3}{R}{R} (Ice Age, S82 OLD #5636):
//
//	"Enchant land
//	 You control enchanted land."
//
// Control Magic's layer 2 control change, on a land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be15e131-0753-4ceb-a36b-4f9172241080",
		Name:         "Conquer",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
	})
}
