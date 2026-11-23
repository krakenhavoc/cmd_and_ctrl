package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lifecrafter's Gift — Instant {3}{G} (Aether Revolt; Jumpstart):
//
//	"Put a +1/+1 counter on target creature, then put a +1/+1 counter on
//	 each creature you control with a +1/+1 counter on it."
//
// Two steps in order. The first counter goes on the target (any
// creature, an opponent's included); the second sweep is read AFTER it
// lands, so a creature of yours that had none and was the target gets
// its second counter, and an opponent's target gets only the one. A
// target that left in response leaves the spell with no legal target,
// and it does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6052d2d5-6310-42fb-bdec-7b114d74697d",
		Name:         "Lifecrafter's Gift",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
			effects.Instruction(func(ctx *effects.Context) error {
				for _, id := range effects.JmpB5CreaturesYouControlWithACounter(ctx.Game, ctx.Controller()) {
					if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
