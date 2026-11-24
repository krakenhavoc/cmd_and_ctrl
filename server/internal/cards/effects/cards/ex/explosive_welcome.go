package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explosive Welcome — Instant {7}{R} (Strixhaven: School of Mages, #5502):
//
//	"Explosive Welcome deals 5 damage to any target and 3 damage to any other
//	 target. Add {R}{R}{R}."
//
// Cone of Flame's two clauses (the second Distinct, "any other target", CR
// 601.2c), dealt as one damage instance; a pick that has become illegal is
// skipped and the other still takes its damage (CR 608.2b). The mana is added
// as the spell resolves and empties at the end of the step like any mana
// (CR 106.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac27ccb8-81e2-42e2-a340-0b0ce9609688",
		Name:         "Explosive Welcome",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Clauses(effects.TargetAny(), effects.Distinct(effects.TargetAny())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5), effects.DamageToTarget(1, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageEachClauseTarget(5, 3)(item, ctx); err != nil {
				return err
			}
			return effects.AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
		},
	})
}
