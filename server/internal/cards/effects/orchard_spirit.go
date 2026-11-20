package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orchard Spirit — Creature — Spirit {2}{G}, 2/2 (Innistrad, #5572):
//
//	"This creature can't be blocked except by creatures with flying or
//	 reach."
//
// Canopy Cover's block rule on the creature itself: a blocker must have
// flying or reach (CR 509.1b), judged when blockers are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6eb56a7-4dfe-4d51-a1df-55f892c1b014",
		Name:         "Orchard Spirit",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedExceptBy(OnSelf(), Or(HasKeyword("flying"), HasKeyword("reach")),
				"creatures with flying or reach"),
		},
	})
}
