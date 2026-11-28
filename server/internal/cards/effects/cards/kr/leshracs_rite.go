package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leshrac's Rite — Enchantment — Aura {B} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Enchanted creature has swampwalk."
//
// A layer 6 keyword grant on the attached creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ced4538a-4919-4ef8-8bf2-2de9e0388587",
		Name:         "Leshrac's Rite",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("swampwalk")},
	})
}
