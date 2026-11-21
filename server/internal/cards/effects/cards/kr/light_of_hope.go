package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Light of Hope — Instant {W}:
//
//	"Choose one —
//	 • You gain 4 life.
//	 • Destroy target enchantment.
//	 • Put a +1/+1 counter on target creature."
//
// Three independent bullets, each with its own target group (or none).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b6f655b-1518-4ab1-8765-3f549819ef3b",
		Name:         "Light of Hope",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("You gain 4 life.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
			}),
			effects.ModeWithPurpose(effects.ModeDoing("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					c, ok := effects.ModeTarget(ctx, occ)
					if !ok || c.Kind != game.TargetCard {
						return nil
					}
					return effects.DestroyTarget{Target: c.ID}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeDoing("Put a +1/+1 counter on target creature.", effects.TargetCreature("target creature"), lightOfHopeCounter),
		),
	})
}

// lightOfHopeCounter is the third bullet: one +1/+1 counter on its target.
func lightOfHopeCounter(_ *game.StackItem, ctx *effects.Context, occ int) error {
	c, ok := effects.ModeTarget(ctx, occ)
	if !ok || c.Kind != game.TargetCard {
		return nil
	}
	return effects.AddCounter{Target: c.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
