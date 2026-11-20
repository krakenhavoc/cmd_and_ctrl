package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raven's Run Dragoon — Creature — Elf Knight {2}{G/W}{G/W}, 3/3
// (Shadowmoor, #5591):
//
//	"This creature can't be blocked by black creatures."
//
// A block rule, Shield Mare's: a blocker of that colour, as it is now, is
// refused (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e405800a-27e6-468a-938b-484346ee20cd",
		Name:         "Raven's Run Dragoon",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), And(Creature(), OfColor("B")), "black creatures"),
		},
	})
}
