package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sicken — Enchantment — Aura {B} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 Enchanted creature gets -1/-1.
//	 Cycling {2}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5208a5f2-eebe-4adc-8f29-543b60116817",
		Name:         "Sicken",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-1, -1)},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
