package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cruel Feeding — Instant {B} (Journey into Nyx, #5558):
//
//	"Strive — This spell costs {2}{B} more to cast for each target
//	 beyond the first.
//	 Any number of target creatures each get +1/+0 and gain lifelink
//	 until end of turn. (Damage dealt by a creature with lifelink also
//	 causes its controller to gain that much life.)"
//
// Strive is Call the Coppercoats' coloured per-target surcharge (CR
// 207.2c). Zero targets is legal; each target still legal as the spell
// resolves gets the effect, and one that left does not (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef60bb5b-7cf5-4b4d-a725-4531fc0c852a",
		Name:         "Cruel Feeding",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "lifelink")),
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{2}{B}", "Strive — This spell costs {2}{B} more to cast for each target beyond the first."),
		},
		OnResolve: effects.JouStrivePump(1, 0, "Cruel Feeding — +1/+0 and lifelink", "lifelink"),
	})
}
