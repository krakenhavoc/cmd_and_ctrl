package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meekstone —
//
// "Creatures with power 3 or greater don't untap during their controllers'
// untap steps."
func init() {
	Register(Spec{
		OracleID:              "5ba73182-30a7-4bad-9cb6-c0feecc2db33",
		Name:                  "Meekstone",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringTheirControllersUntapSteps(And(Creature(), PowerGE(3)))},
	})
}
