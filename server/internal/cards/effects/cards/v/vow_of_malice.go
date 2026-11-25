package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vow of Malice — Enchantment — Aura {2}{B} (Commander 2015, S82 OLD #5546):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has intimidate, and can't attack you or
//	 planeswalkers you control."
//
// Vow of Duty's three statics with intimidate for vigilance, all read off the
// Aura's attachment so they end when it leaves: the layer 7c pump, the keyword
// grant, and the CR 508.1c attack restriction naming the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "894bc28a-6627-43dd-9650-86300b6fe5e6",
		Name:         "Vow of Malice",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("intimidate"),
			effects.Cmdr2B19AttachedCantAttackYou(),
		},
	})
}
