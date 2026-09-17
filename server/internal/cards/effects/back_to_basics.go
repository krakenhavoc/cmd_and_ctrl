package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Back to Basics —
//
// "Nonbasic lands don't untap during their controllers' untap steps."
func init() {
	Register(Spec{
		OracleID:              "05c2dec2-d2f7-4036-b91f-4fccba10a8bb",
		Name:                  "Back to Basics",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringTheirControllersUntapSteps(b751NonbasicLand())},
	})
}
