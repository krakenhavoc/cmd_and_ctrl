package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Juggernaut — Artifact Creature — Juggernaut {4}, 5/3 (Foundations,
// #4465):
//
//	"This creature attacks each combat if able.
//	 This creature can't be blocked by Walls."
//
// Zurgo Helmsmasher's AttacksEachCombat requirement (CR 508.1d) and
// Stromkirk Noble's CantBeBlockedBy block rule with a creature-type test
// (Subtype "Wall", read through the layers), checked as blockers are
// declared (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ac9116f-36bc-4d71-b696-d6ee064e1d58",
		Name:         "Juggernaut",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), Subtype("Wall"), "Walls"),
		},
	})
}
