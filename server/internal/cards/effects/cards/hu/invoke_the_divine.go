package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invoke the Divine — Instant {2}{W}:
//
//	"Destroy target artifact or enchantment. You gain 4 life."
//
// A target that is illegal on resolution fizzles the whole spell, so no
// life is gained (the ruling, CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87bd0fdb-c5b9-46ea-9858-1a870960351f",
		Name:         "Invoke the Divine",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 4}.Apply(ctx)
		},
	})
}
