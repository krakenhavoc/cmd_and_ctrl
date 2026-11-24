package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// On Serra's Wings — Legendary Enchantment — Aura {3}{W} (Commander Legends):
//
//	"Enchant creature
//	 Enchanted creature is legendary, gets +1/+1, and has flying, vigilance,
//	 and lifelink."
//
// Three statics read off the attachment: a layer 4 Legendary supertype (CR
// 613.1d, so the legend rule and every "legendary creature" reader see it
// through the effective supertypes), the layer 7c +1/+1, and the layer 6
// keywords. The Aura is itself a legendary enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f43584de-4f60-46e6-9e6b-e704bc3ecd59",
		Name:         "On Serra's Wings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			{
				Label:     "Enchanted creature is legendary.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply:     effects.AddLegendarySupertype,
			},
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("flying", "vigilance", "lifelink"),
		},
	})
}
