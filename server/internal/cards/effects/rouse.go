package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rouse — Instant {1}{B} (Mercadian Masques, #5619):
//
//	"If you control a Swamp, you may pay 2 life rather than pay this
//	 spell's mana cost.
//	 Target creature gets +2/+0 until end of turn."
//
// Snuff Out's alternative cost at 2 life (PayLifeInstead with
// ControlsA("Swamp"), checked as the spell is cast) on the shared
// boostTheTargetUntilEOT body. The life is a cost, paid with the spell
// on the stack (CR 601.2h), so it is not given back if Rouse is
// countered. The boost ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "475099d9-4743-420a-ada8-3932f5ef9e82",
		Name:         "Rouse",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0)),
		Targets:      TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{
			PayLifeInstead("Pay 2 life (you control a Swamp)", 2, ControlsA("Swamp")),
		},
		OnResolve: boostTheTargetUntilEOT(2, 0, "Rouse — +2/+0"),
	})
}
