package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Favor of the Overbeing — Enchantment — Aura {1}{G/U}:
//
//	Enchant creature
//	As long as enchanted creature is green, it gets +1/+1 and has vigilance.
//	As long as enchanted creature is blue, it gets +1/+1 and has flying.
//
// Shield of the Oversoul's shape: each clause is a conditional static read live off the enchanted creature's post-layer colours (layer 5 runs first), so a creature of both colours gets both halves and a colourless one gets neither (CR 613.1).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56568fc4-7133-40e2-833c-29ecc1f953d6",
		Name:         "Favor of the Overbeing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.B43AttachedHasColor("G"), 1, 1),
			effects.B43GrantToAttachedWhen(effects.B43AttachedHasColor("G"), "vigilance"),
			effects.B43PumpAttachedWhen(effects.B43AttachedHasColor("U"), 1, 1),
			effects.B43GrantToAttachedWhen(effects.B43AttachedHasColor("U"), "flying"),
		},
	})
}
