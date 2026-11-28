package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marble Titan — Creature — Giant {3}{W}, 3/3 (Tempest, S82 OLD, #5628):
//
//	"Creatures with power 3 or greater don't untap during their controllers'
//	 untap steps."
//
// An untap-step restriction over every creature at the table whose effective
// power is 3 or more as the untap step asks (CR 502.3), the Titan itself and
// its controller's creatures included. Power is read live, so a creature
// shrunk below 3 untaps and one pumped to 3 does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62674cfd-f400-44d9-bfff-15bdfbd243fe",
		Name:         "Marble Titan",
		Completeness: effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{
			effects.DoesntUntapDuringTheirControllersUntapSteps(effects.And(effects.Creature(), effects.PowerGE(3))),
		},
	})
}
