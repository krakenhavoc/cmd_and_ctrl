package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6052d2d5-6310-42fb-bdec-7b114d74697d",
		Name:         "Lifecrafter's Gift",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
			Instruction(func(ctx *Context) error {
				for _, id := range jmpB5CreaturesYouControlWithACounter(ctx.Game, ctx.Controller()) {
					if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
