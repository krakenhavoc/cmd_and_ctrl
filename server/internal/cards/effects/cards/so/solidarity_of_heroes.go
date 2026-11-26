package cardsso

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Solidarity of Heroes — Instant {1}{G} (Commander 2016, S82 older sets,
// #5537):
//
//	"Strive — This spell costs {1}{G} more to cast for each target beyond the
//	 first.
//	 Choose any number of target creatures. Double the number of +1/+1
//	 counters on each of them."
//
// Any number of targets, including none; each still-legal target is doubled
// on resolution, a creature with no +1/+1 counters gets none (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3319b2df-7354-44b5-afa1-b9bc4a577254",
		Name:         "Solidarity of Heroes",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{1}{G}", "Strive — This spell costs {1}{G} more to cast for each target beyond the first."),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			chosen := effects.LegalTargetCardIDs(ctx)
			return effects.DoublePlusOneCountersOnEach(ctx.Game, ctx.Item, func(c game.Card) bool {
				return slices.Contains(chosen, c.InstanceID)
			})
		},
	})
}
