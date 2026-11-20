package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arlinn's Wolf — Creature — Wolf {2}{G}, 3/2 (War of the Spark, #5520):
//
//	"This creature can't be blocked by creatures with power 2 or less."
//
// Gate Colossus's evasion clause: a block rule on itself that refuses
// any blocker with power 2 or less.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe4c0f02-c5fd-4cd5-9f80-0936e3875f1c",
		Name:         "Arlinn's Wolf",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
