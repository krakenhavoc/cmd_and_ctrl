package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desperate Stand — Sorcery {R}{W} (Journey into Nyx, #5558):
//
//	"Strive — This spell costs {R}{W} more to cast for each target
//	 beyond the first.
//	 Any number of target creatures each get +2/+0 and gain first
//	 strike and vigilance until end of turn."
//
// Strive is Call the Coppercoats' coloured per-target surcharge (CR
// 207.2c). Zero targets is legal; each target still legal as the spell
// resolves gets the effect, and one that left does not (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8252564-3e61-4ed5-bef1-6ec22882854b",
		Name:         "Desperate Stand",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "first strike", "vigilance")),
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{R}{W}", "Strive — This spell costs {R}{W} more to cast for each target beyond the first."),
		},
		OnResolve: effects.JouStrivePump(2, 0, "Desperate Stand — +2/+0, first strike and vigilance", "first strike", "vigilance"),
	})
}
