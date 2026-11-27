package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Edge of the Divinity — Enchantment — Aura {W/B}:
//
//	Enchant creature
//	As long as enchanted creature is white, it gets +1/+2.
//	As long as enchanted creature is black, it gets +2/+1.
//
// Shield of the Oversoul's shape: each clause is a conditional static read live off the enchanted creature's post-layer colours (layer 5 runs first), so a creature of both colours gets both halves and a colourless one gets neither (CR 613.1).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16ef24ea-4d84-43e5-bd4f-b6efadc87616",
		Name:         "Edge of the Divinity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.B43AttachedHasColor("W"), 1, 2),
			effects.B43PumpAttachedWhen(effects.B43AttachedHasColor("B"), 2, 1),
		},
	})
}
