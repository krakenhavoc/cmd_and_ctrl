package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phalanx Formation — Instant {2}{W} (Journey into Nyx, #5558):
//
//	"Strive — This spell costs {1}{W} more to cast for each target
//	 beyond the first.
//	 Any number of target creatures each gain double strike until end
//	 of turn. (They deal both first-strike and regular combat damage.)"
//
// Strive is Call the Coppercoats' coloured per-target surcharge (CR
// 207.2c). Zero targets is legal; each target still legal as the spell
// resolves gets the effect, and one that left does not (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "799e7bb4-b03c-4568-b321-7810868bdcd7",
		Name:         "Phalanx Formation",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "double strike")),
		Targets:      TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			CostsMorePerTargetBeyondFirst("{1}{W}", "Strive — This spell costs {1}{W} more to cast for each target beyond the first."),
		},
		OnResolve: jouStrivePump(0, 0, "Phalanx Formation — double strike", "double strike"),
	})
}
