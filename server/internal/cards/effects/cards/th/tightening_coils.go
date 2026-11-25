package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tightening Coils — Enchantment — Aura {1}{U} (Battle for Zendikar, #5547):
//
//	"Enchant creature
//	 Enchanted creature gets -6/-0 and loses flying."
//
// Both halves are statics on the Aura, so they end the moment it leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71af7e54-e22d-4354-a35c-80e4e5f4031c",
		Name:         "Tightening Coils",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-6, 0), effects.RemoveFromAttached("flying")},
	})
}
