package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weight of the Underworld — Enchantment — Aura {3}{B} (Born of the Gods,
// #5559):
//
//	"Enchant creature
//	 Enchanted creature gets -3/-2."
//
// Dead Weight at -3/-2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b960d17b-3679-4dc5-9a5d-c8e7fd81fa9a",
		Name:         "Weight of the Underworld",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpEnchantedCreature(-3, -2)},
	})
}
