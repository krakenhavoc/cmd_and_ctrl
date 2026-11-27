package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gift of the Deity — Enchantment — Aura {4}{B/G}:
//
//	Enchant creature
//	As long as enchanted creature is black, it gets +1/+1 and has deathtouch.
//	As long as enchanted creature is green, it gets +1/+1 and all creatures able to block it do so.
//
// Shield of the Oversoul's shape: each clause is a conditional static read live off the enchanted creature's post-layer colours (layer 5 runs first), so a creature of both colours gets both halves and a colourless one gets neither (CR 613.1).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3548489b-053e-4f81-a0be-07422cbe357d",
		Name:         "Gift of the Deity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.B43AttachedHasColor("B"), 1, 1),
			effects.B43GrantToAttachedWhen(effects.B43AttachedHasColor("B"), "deathtouch"),
			effects.B43PumpAttachedWhen(effects.B43AttachedHasColor("G"), 1, 1),
			effects.BlockRequirementWhere(game.BlockRequirementLure, effects.B43AttachedHasColor("G")),
		},
	})
}
