package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clout of the Dominus — Enchantment — Aura {U/R}:
//
//	Enchant creature
//	As long as enchanted creature is blue, it gets +1/+1 and has shroud.
//	As long as enchanted creature is red, it gets +1/+1 and has haste.
//
// Shield of the Oversoul's shape: each clause is a conditional static read live off the enchanted creature's post-layer colours (layer 5 runs first), so a creature of both colours gets both halves and a colourless one gets neither (CR 613.1).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e412163-a93a-4738-a8f0-a4d27c960360",
		Name:         "Clout of the Dominus",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.B43AttachedHasColor("U"), 1, 1),
			effects.B43GrantToAttachedWhen(effects.B43AttachedHasColor("U"), "shroud"),
			effects.B43PumpAttachedWhen(effects.B43AttachedHasColor("R"), 1, 1),
			effects.B43GrantToAttachedWhen(effects.B43AttachedHasColor("R"), "haste"),
		},
	})
}
