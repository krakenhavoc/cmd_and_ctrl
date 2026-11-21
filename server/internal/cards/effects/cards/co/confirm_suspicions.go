package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Confirm Suspicions — Instant {3}{U}{U} (Murders at Karlov Manor
// Commander, #5368):
//
//	"Counter target spell.
//	 Investigate three times."
//
// Cancel's counter, then three Clues (CR 701.16a). The Clues are made
// even when the target was already gone: the spell is only countered if
// its target is still legal, but with an illegal target the whole spell
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a30d076b-c942-45e4-b943-3749ce955291",
		Name:         "Confirm Suspicions",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (effects.CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.Investigate{N: 3}.Apply(ctx)
		},
	})
}
