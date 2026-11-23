package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Back to Basics —
//
// "Nonbasic lands don't untap during their controllers' untap steps."
func init() {
	effects.Register(effects.Spec{
		OracleID:              "05c2dec2-d2f7-4036-b91f-4fccba10a8bb",
		Name:                  "Back to Basics",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.B751NonbasicLand())},
	})
}
