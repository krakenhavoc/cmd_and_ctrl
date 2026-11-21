package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight's Pledge — Enchantment — Aura {1}{W} (Core Set 2019, #5527):
//
//	"Enchant creature
//	Enchanted creature gets +2/+2."
//
// Spectral Steel's Aura half: enchant creature and +2/+2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63604320-68de-415e-b1f9-dd0f85c5d8a3",
		Name:         "Knight's Pledge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
	})
}
