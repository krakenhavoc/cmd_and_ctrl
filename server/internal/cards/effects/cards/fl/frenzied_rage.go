package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frenzied Rage — Enchantment — Aura {1}{R} (Dominaria, #5530):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1 and has menace. (It can't be blocked except by two or more creatures.)"
//
// Untamed Hunger's Aura: +2/+1 and menace on the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a04c31c-f5e4-43ff-a5b5-207426b8dca6",
		Name:         "Frenzied Rage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1), effects.GrantToAttached("menace")},
	})
}
