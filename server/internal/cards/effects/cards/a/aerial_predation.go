package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aerial Predation — Instant {2}{G} (Return to Ravnica, #5566):
//
//	"Destroy target creature with flying. You gain 2 life."
//
// "With flying" is read through the layered view when the target is chosen and
// again on resolution (CR 608.2b): a creature that has lost flying in response
// is an illegal target and the spell does nothing, the life included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bac1495-16cf-4e1e-9126-482f85322c78",
		Name:         "Aerial Predation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, LifeGain: 2},
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.Do(effects.DestroyTarget{Target: id}, effects.GainLife{Amount: 2})(ctx.Game, ctx.Item)
		},
	})
}
