package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Challenger Troll — Creature — Troll {4}{G}, 6/5:
//
//	"Each creature you control with power 4 or greater can't be blocked
//	 by more than one creature."
//
// A block rule scoped to the creatures its controller controls whose
// current power is 4 or greater, read live as each block is declared
// (CR 509.1b). It binds the Troll itself while its power stays 4 or more,
// and it does not release a creature that is already blocked by two or
// more (nothing is re-checked after the declaration).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78cb1f05-edad-4430-9095-f3eabd80bf41",
		Name:         "Challenger Troll",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			MaxBlockers(OnMatching(And(Creature(), YouControl(), PowerGE(4))), 1),
		},
	})
}
