package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerial Formation — Instant {U} (Journey into Nyx, #5558):
//
//	"Strive — This spell costs {2}{U} more to cast for each target
//	 beyond the first.
//	 Any number of target creatures each get +1/+1 and gain flying
//	 until end of turn."
//
// Strive is Call the Coppercoats' coloured per-target surcharge (CR
// 207.2c). Zero targets is legal; each target still legal as the spell
// resolves gets the effect, and one that left does not (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0168b7e-a8b8-4ce3-af12-a43f3818b6a1",
		Name:         "Aerial Formation",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "flying")),
		Targets:      TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			CostsMorePerTargetBeyondFirst("{2}{U}", "Strive — This spell costs {2}{U} more to cast for each target beyond the first."),
		},
		OnResolve: jouStrivePump(1, 1, "Aerial Formation — +1/+1 and flying", "flying"),
	})
}
