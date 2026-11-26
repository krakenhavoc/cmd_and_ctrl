package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Setessan Training — Enchantment — Aura {1}{G} (Core Set 2021, #5511):
//
//	"Enchant creature you control
//	 When this Aura enters, draw a card.
//	 Enchanted creature gets +1/+0 and has trample."
//
// The enchant clause admits only a creature you control, and the draw is the
// Aura's own enters trigger, so it happens even if the creature is gone by
// the time it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ff3b9c4-80ec-44d9-99ec-6fea3a86468b",
		Name:         "Setessan Training",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Setessan Training — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
