package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cursed Flesh — Enchantment — Aura {B} (Invasion, S82 older sets batch 661, #5616):
//
//	"Enchant creature
//	 Enchanted creature gets -1/-1 and has fear. (It can't be blocked except
//	 by artifact creatures and/or black creatures.)"
//
// A layer 7c -1/-1 and a layer 6 fear grant, both on the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96e244ef-21b9-4795-a881-c3db23b60cc9",
		Name:         "Cursed Flesh",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpAttached(-1, -1),
			effects.GrantToAttached("fear"),
		},
	})
}
