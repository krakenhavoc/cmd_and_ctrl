package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Power Boost — Enchantment {2}{R} (Marvel Super Heroes Commander,
// #4455): "Creatures you control get +1/+0." A plain +1/+0 anthem over
// the controller's creatures, built the way Domri's static is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fed58ccc-db56-4094-8b56-bd29b00e9327",
		Name:         "Power Boost",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.B41CreatureYouControlStatic, 1, 0)},
	})
}
