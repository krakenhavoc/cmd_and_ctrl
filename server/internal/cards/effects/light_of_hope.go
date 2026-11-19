package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5b6f655b-1518-4ab1-8765-3f549819ef3b",
		Name:         "Light of Hope",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("You gain 4 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
			}),
			ModeWithPurpose(ModeDoing("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					c, ok := ModeTarget(ctx, occ)
					if !ok || c.Kind != game.TargetCard {
						return nil
					}
					return DestroyTarget{Target: c.ID}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerRemove}),
			ModeDoing("Put a +1/+1 counter on target creature.", TargetCreature("target creature"), lightOfHopeCounter),
		),
	})
}

// lightOfHopeCounter is the third bullet: one +1/+1 counter on its target.
func lightOfHopeCounter(_ *game.StackItem, ctx *Context, occ int) error {
	c, ok := ModeTarget(ctx, occ)
	if !ok || c.Kind != game.TargetCard {
		return nil
	}
	return AddCounter{Target: c.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
