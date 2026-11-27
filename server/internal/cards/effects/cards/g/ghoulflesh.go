package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghoulflesh — {B} Aura: "Enchant creature. Enchanted creature gets -1/-1
// and is a black Zombie in addition to its other colors and types."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "640c0424-b862-438f-a527-ff606151040a",
		Name:         "Ghoulflesh",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpAttached(-1, -1),
			effects.AddAttachedSubtypes("Zombie"),
			effects.AddAttachedColors("B"),
		},
	})
}
