package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fa5a9083-e6f7-4f61-9327-a6b96db220c7",
		Name:         "Crackdown",
		Completeness: CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{
			doesntUntapDuringTheirControllersUntapSteps(And(Creature(), PowerGE(3), Not(OfColor("W")))),
		},
	})
}
