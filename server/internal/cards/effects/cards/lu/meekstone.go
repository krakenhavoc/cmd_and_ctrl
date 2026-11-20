package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meekstone —
//
// "Creatures with power 3 or greater don't untap during their controllers'
// untap steps."
func init() {
	effects.Register(effects.Spec{
		OracleID:              "5ba73182-30a7-4bad-9cb6-c0feecc2db33",
		Name:                  "Meekstone",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.And(effects.Creature(), effects.PowerGE(3)))},
	})
}
