package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vow of Lightning — Enchantment — Aura {2}{R} (Commander wave 2, #5368):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has first strike, and can't attack you or
//	 planeswalkers you control."
//
// Three statics on the Aura, all read off its attachment so they end when
// it leaves: the layer 7c pump, the keyword grant, and the CR 508.1c
// attack restriction naming the Aura's controller (cmdr2B19AttachedCantAttackYou).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfed67f3-70d2-49c2-8f25-16e4d10fa5be",
		Name:         "Vow of Lightning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("first strike"),
			effects.Cmdr2B19AttachedCantAttackYou(),
		},
	})
}
