package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiora's Dismissal — Instant {U} (Journey into Nyx):
//
//	"Strive — This spell costs {U} more to cast for each target beyond the first.
//	 Return any number of target enchantments to their owners' hands."
//
// Consign to Dust's shape: any number of targets (zero allowed), each one
// still legal as the spell resolves goes to its owner's hand (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f866b14-dcdd-4c7d-b935-6db431d83b86",
		Name:         "Kiora's Dismissal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("any number of target enchantments", effects.Enchantment()).WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{U}", "Strive — This spell costs {U} more to cast for each target beyond the first."),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceEachLegalTarget(ctx.Game, ctx.Item)
		},
	})
}
