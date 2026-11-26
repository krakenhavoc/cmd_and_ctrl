package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lost in the Mist — Instant {3}{U}{U} (Innistrad, S82 OLD batch 450,
// #5572):
//
//	"Counter target spell. Return target permanent to its owner's hand."
//
// Trap Essence's two target clauses: each half acts only while its own
// target is still legal (CR 608.2b), so the spell is countered and the
// permanent returned independently.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98014ad2-7a2b-45f5-808c-edd4d0b2c597",
		Name:         "Lost in the Mist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetSpell("target spell").Then(
			effects.TargetPermanent("target permanent")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) > 0 && ctx.IsTargetLegal(item.Targets[0]) {
				if err := effects.CounterTheTargetSpell(item, ctx); err != nil {
					return err
				}
			}
			if len(item.Targets) > 1 && ctx.IsTargetLegal(item.Targets[1]) {
				return effects.BounceToHand{Target: item.Targets[1].ID}.Apply(ctx)
			}
			return nil
		},
	})
}
