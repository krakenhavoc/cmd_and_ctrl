package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vow of Flight — Enchantment — Aura {2}{U} (Commander Legends):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has flying, and can't attack you or
//	 planeswalkers you control."
//
// Vow of Lightning's three statics: the layer 7c pump, the keyword grant, and
// the CR 508.1c attack restriction naming the Aura's controller
// (Cmdr2B19AttachedCantAttackYou), all read off the attachment so they end
// when the Aura leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81fe554a-5257-446c-ad53-ee45bf5edb40",
		Name:         "Vow of Flight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
			effects.Cmdr2B19AttachedCantAttackYou(),
		},
	})
}
