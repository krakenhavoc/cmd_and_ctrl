package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fists of the Demigod — Enchantment — Aura {1}{B/R} (Shadowmoor, S82 older sets):
//
//	"Enchant creature
//	 As long as enchanted creature is black, it gets +1/+1 and has wither.
//	 As long as enchanted creature is red, it gets +1/+1 and has first strike."
//
// Shield of the Oversoul's shape: each clause is a +1/+1 static and a granted
// keyword, both conditional on the enchanted creature's colour as the layers
// leave it (colour is layer 5, before abilities and P/T), so a creature of both
// colours gets both clauses and a colourless one neither.
//
// No simplification.
func init() {
	one := effects.B43AttachedHasColor("B")
	two := effects.B43AttachedHasColor("R")
	effects.Register(effects.Spec{
		OracleID:     "0e6f9fc3-5cc7-42cc-ae83-9be875a1579f",
		Name:         "Fists of the Demigod",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(one, 1, 1),
			effects.B43GrantToAttachedWhen(one, "wither"),
			effects.B43PumpAttachedWhen(two, 1, 1),
			effects.B43GrantToAttachedWhen(two, "first strike"),
		},
	})
}
