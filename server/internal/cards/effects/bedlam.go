package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bedlam — Enchantment {2}{R}{R}:
//
//	"Creatures can't block."
//
// One block rule (CR 509.1b) whose blockers are every creature on the
// battlefield, whoever controls it, against every attacker: the engine
// refuses any declared block while Bedlam is on the battlefield. Both
// sides are read from effective characteristics at the moment the block
// is declared, so a permanent that becomes a creature is covered and
// Bedlam leaving lets creatures block again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c28d4caf-9b7c-40b1-9cbe-275ca9dae06a",
		Name:         "Bedlam",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnMatching(Creature()), OnMatching(Creature()), "Bedlam: creatures can't block"),
		},
	})
}
