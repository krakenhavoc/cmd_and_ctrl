package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serene Offering — Instant {1}{W} (Tempest, S82 OLD, #5628):
//
//	"Destroy target enchantment. You gain life equal to its mana value."
//
// Two instructions in printed order (CR 608.2c). The mana value is read
// BEFORE the destroy, as Expel the Unworthy reads it, and the life is gained
// whether or not the enchantment was destroyed: an indestructible one still
// pays out, since "its" names the target, not the result.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6cef0548-6010-4781-8e83-431d014dd1da",
		Name:         "Serene Offering",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			mv, _ := ctx.Game.ManaValueForEffect(card)
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: mv}.Apply(ctx)
		},
	})
}
