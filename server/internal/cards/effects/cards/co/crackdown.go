package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crackdown — Enchantment {2}{W}:
//
//	"Nonwhite creatures with power 3 or greater don't untap during their
//	 controllers' untap steps."
//
// Meekstone's untap-step restriction (CR 502.3) with a colour clause: it
// binds every player's creatures, the controller's included, and reads the
// creature's layered power and colours at the untap step, so a creature that
// shrinks below 3 or becomes white untaps again.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa5a9083-e6f7-4f61-9327-a6b96db220c7",
		Name:         "Crackdown",
		Completeness: effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{
			effects.DoesntUntapDuringTheirControllersUntapSteps(effects.And(effects.Creature(), effects.PowerGE(3), effects.Not(effects.OfColor("W")))),
		},
	})
}
