package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nature's Panoply — Instant {G} (Journey into Nyx):
//
//	"Strive — This spell costs {2}{G} more to cast for each target beyond the first.
//	 Choose any number of target creatures. Put a +1/+1 counter on each of them."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3126751-27fe-42d3-a001-33ecee72b528",
		Name:         "Nature's Panoply",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{2}{G}", "Strive — This spell costs {2}{G} more to cast for each target beyond the first."),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutPlusOneCounterOnEachLegalTarget(ctx.Game, ctx.Item)
		},
	})
}
