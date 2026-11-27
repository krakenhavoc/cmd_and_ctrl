package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zephyr Net — Enchantment — Aura {1}{U} (Lorwyn, S82 OLD, #5593):
//
//	"Enchant creature
//	 Enchanted creature has defender and flying."
//
// Sky Tether's defender grant beside Angelic Gift's flying grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd510cf2-df54-4024-919b-e4377cdb2be1",
		Name:         "Zephyr Net",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.GrantToAttached("defender", "flying")},
	})
}
