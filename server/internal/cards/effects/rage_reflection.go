package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rage Reflection — Enchantment {4}{R}{R} (Shadowmoor, #5591):
//
//	"Creatures you control have double strike."
//
// Brave the Sands' team grant (b16GrantKeywords over the creatures you
// control), a layer 6 effect (CR 613.1f) that reaches a creature the
// moment it enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03fb26b1-6d9c-455b-9b2f-a052a9222db7",
		Name:         "Rage Reflection",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "double strike"),
		},
	})
}
