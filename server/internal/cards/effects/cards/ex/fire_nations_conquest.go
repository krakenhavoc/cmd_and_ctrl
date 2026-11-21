package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Nation's Conquest — Enchantment {2}{R} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Creatures you control get +1/+0."
//
// Power Boost's anthem: a layer-7c +1/+0 to every creature you control,
// read continuously (CR 611.3a), so a creature that arrives later gets
// it too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8893af49-2f90-4aed-bcaf-905bb9e8ed8f",
		Name:         "Fire Nation's Conquest",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.B41CreatureYouControlStatic, 1, 0)},
	})
}
