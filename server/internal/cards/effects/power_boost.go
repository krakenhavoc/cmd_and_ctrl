package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Power Boost — Enchantment {2}{R} (Marvel Super Heroes Commander,
// #4455): "Creatures you control get +1/+0." A plain +1/+0 anthem over
// the controller's creatures, built the way Domri's static is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fed58ccc-db56-4094-8b56-bd29b00e9327",
		Name:         "Power Boost",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b16Anthem(b41CreatureYouControlStatic, 1, 0)},
	})
}
