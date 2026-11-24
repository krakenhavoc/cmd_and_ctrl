package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arachnoform — Enchantment — Aura {1}{G} (Kaldheim, #5505):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has reach, and is every creature type."
//
// A 7c pump, a layer-6 keyword and the layer-4 "every creature type"
// fact Maskwood Nexus uses (ApplyAllCreatureTypes).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e41ae94-db31-4060-881b-a2f6d834c1ee",
		Name:         "Arachnoform",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("reach"),
			{
				Label:     "Enchanted creature is every creature type.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply:     effects.ApplyAllCreatureTypes,
			},
		},
	})
}
