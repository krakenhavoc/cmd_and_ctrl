package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flight — Enchantment — Aura {U} (Magic 2012, S82 older sets batch 451,
// #5573):
//
//	"Enchant creature
//	 Enchanted creature has flying."
//
// Pollenbright Wings' grant, without the trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a4068b0-fb4f-429c-a94e-47849f3eb7ef",
		Name:         "Flight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
	})
}
