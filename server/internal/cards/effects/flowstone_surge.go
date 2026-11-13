package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flowstone Surge — Enchantment {1}{R}
// (Nemesis, #5618):
//
//	"Creatures you control get +1/-1."
//
// A layer 7c anthem (CR 613.4c) over the creatures its controller
// controls, at +1/-1. A creature with 1 toughness dies to the
// state-based check (CR 704.5f) while the Surge is on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb1755a0-3334-419b-8cb5-5a3ac7fa5b13",
		Name:         "Flowstone Surge",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{YoursOnly: true}, 1, -1),
		},
	})
}
