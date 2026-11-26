package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guard Duty — Enchantment — Aura {W} (Rise of the Eldrazi, S82 OLD batch
// 484, #5581):
//
//	"Enchant creature
//	 Enchanted creature has defender."
//
// Sky Tether's grant, without its other half: a layer 6 keyword grant over the
// attached creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b20e187-c200-40a0-8db2-481f0bf90168",
		Name:         "Guard Duty",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.GrantToAttached("defender")},
	})
}
