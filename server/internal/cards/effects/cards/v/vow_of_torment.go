package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vow of Torment — Enchantment — Aura {2}{B} (Commander Legends):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has menace, and can't attack you or
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
		OracleID:     "237b87f4-783b-40ac-9fcf-4ae76a33229d",
		Name:         "Vow of Torment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("menace"),
			effects.Cmdr2B19AttachedCantAttackYou(),
		},
	})
}
