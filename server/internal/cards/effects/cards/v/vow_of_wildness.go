package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vow of Wildness — Enchantment — Aura {2}{G} (Commander Legends):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3, has trample, and can't attack you or
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
		OracleID:     "569a65fb-e35f-4388-ab49-f57be46b4889",
		Name:         "Vow of Wildness",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("trample"),
			effects.Cmdr2B19AttachedCantAttackYou(),
		},
	})
}
