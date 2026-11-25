package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// In Bolas's Clutches — Legendary Enchantment — Aura {4}{U}{U} (Dominaria, S82
// OLD, #5530):
//
//	"Enchant permanent
//	 You control enchanted permanent.
//	 Enchanted permanent is legendary."
//
// Confiscate's layer-2 control change on the widest enchant clause, with On
// Serra's Wings' layer-4 Legendary supertype, so the legend rule sees the
// permanent as legendary for as long as the Aura is there. The Aura is itself
// a legendary enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "201a4455-1b6b-4e13-a0d5-4cf2116e1673",
		Name:         "In Bolas's Clutches",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(),
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
			{
				Label:     "Enchanted permanent is legendary.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply:     effects.AddLegendarySupertype,
			},
		},
	})
}
