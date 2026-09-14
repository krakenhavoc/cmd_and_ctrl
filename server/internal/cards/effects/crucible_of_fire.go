package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crucible of Fire — Enchantment {3}{R} (EDHREC rank 3792):
//
//	"Dragon creatures you control get +3/+3."
//
// The Dragon anthem: a layer 7c lord static over the controller's
// Dragons — every one of them, no "other" — through the shared
// TribalAnthem builder, so a changeling counts and an animated
// Dragon Egg does not until it is a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c572396-4e53-4c86-9b04-f41341dcea05",
		Name:         "Crucible of Fire",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Dragon"}, YoursOnly: true}, 3, 3),
		},
	})
}
