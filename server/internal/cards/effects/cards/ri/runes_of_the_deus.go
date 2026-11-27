package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Runes of the Deus — Enchantment — Aura {4}{R/G} (Shadowmoor, S82 older sets):
//
//	"Enchant creature
//	 As long as enchanted creature is red, it gets +1/+1 and has double strike.
//	 As long as enchanted creature is green, it gets +1/+1 and has trample."
//
// Shield of the Oversoul's shape: each clause is a +1/+1 static and a granted
// keyword, both conditional on the enchanted creature's colour as the layers
// leave it (colour is layer 5, before abilities and P/T), so a creature of both
// colours gets both clauses and a colourless one neither.
//
// No simplification.
func init() {
	one := effects.B43AttachedHasColor("R")
	two := effects.B43AttachedHasColor("G")
	effects.Register(effects.Spec{
		OracleID:     "674b9469-3213-4cb0-be7e-db209bc09f92",
		Name:         "Runes of the Deus",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(one, 1, 1),
			effects.B43GrantToAttachedWhen(one, "double strike"),
			effects.B43PumpAttachedWhen(two, 1, 1),
			effects.B43GrantToAttachedWhen(two, "trample"),
		},
	})
}
