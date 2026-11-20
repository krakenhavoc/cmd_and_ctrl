package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ajani's Presence — Instant {W} (Journey into Nyx, #5558):
//
//	"Strive — This spell costs {2}{W} more to cast for each target
//	 beyond the first.
//	 Any number of target creatures each get +1/+1 and gain
//	 indestructible until end of turn. (Damage and effects that say
//	 "destroy" don't destroy them.)"
//
// Strive is Call the Coppercoats' coloured per-target surcharge (CR
// 207.2c). Zero targets is legal; each target still legal as the spell
// resolves gets the effect, and one that left does not (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "925a0b9b-a1fa-4aad-89d9-e2698ac3582f",
		Name:         "Ajani's Presence",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "indestructible")),
		Targets:      TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			CostsMorePerTargetBeyondFirst("{2}{W}", "Strive — This spell costs {2}{W} more to cast for each target beyond the first."),
		},
		OnResolve: jouStrivePump(1, 1, "Ajani's Presence — +1/+1 and indestructible", "indestructible"),
	})
}
