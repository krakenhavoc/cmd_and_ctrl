package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampaging Ceratops — Creature — Dinosaur {4}{R}, 5/4:
//
//	"This creature can't be blocked except by three or more creatures."
//
// The reference card for a MINIMUM count rule (#750, ADR 0045
// addendum Decision 12): MinBlockers(OnSelf(), 3). Menace is the same
// shape with two, built into the engine; this is the first minimum a
// card declares. One or two blockers are refused whole at declaration
// with too_few_blockers, and the legal-move enumerator offers a
// three-creature group instead.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91d01119-9b3e-4629-8cf0-bbca2c19a7cc",
		Name:         "Rampaging Ceratops",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			MinBlockers(OnSelf(), 3),
		},
	})
}
