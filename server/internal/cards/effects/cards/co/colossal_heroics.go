package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Colossal Heroics — Instant {2}{G} (Journey into Nyx):
//
//	"Strive — This spell costs {1}{G} more to cast for each target beyond the first.
//	 Any number of target creatures each get +2/+2 until end of turn. Untap those creatures."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b01acaf-bcf4-4ac1-ac62-2f25a45c2589",
		Name:         "Colossal Heroics",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{1}{G}", "Strive — This spell costs {1}{G} more to cast for each target beyond the first."),
		},
		OnResolve: effects.S82OldB366PumpAndUntapEachTarget(2, 2, "Colossal Heroics — +2/+2 until end of turn"),
	})
}
