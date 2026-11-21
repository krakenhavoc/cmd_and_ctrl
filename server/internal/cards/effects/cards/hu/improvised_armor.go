package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Improvised Armor — Enchantment — Aura {3}{W} (Dominaria Remastered, #5948):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+5.
//	 Cycling {3}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aca7c0a7-b365-421e-aeb0-49d3a9873e4f",
		Name:         "Improvised Armor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 5)},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
