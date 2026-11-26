package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Messenger's Speed — Enchantment — Aura {R} (Theros, S82 OLD b390, #5561):
//
//	"Enchant creature
//	 Enchanted creature has trample and haste."
//
// A layer 6 keyword grant to the host (Prophetic Ravings' haste, with trample).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa21315d-4da9-4bbd-9604-c406b83c0db2",
		Name:         "Messenger's Speed",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("trample", "haste")},
	})
}
