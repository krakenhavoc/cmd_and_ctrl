package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unbreakable Formation — Instant {2}{W} (EDHREC top 1,000, #3654):
//
//	"Creatures you control gain indestructible until end of turn.
//	 Addendum — If you cast this spell during your main phase, put a
//	 +1/+1 counter on each of those creatures and they gain vigilance
//	 until end of turn."
//
// "Those creatures" are the ones that gained indestructible, fixed as
// the spell resolves (CR 611.2c), so the counters go on the same set.
// The addendum is read at resolution with b14InYourMainPhase, which is
// the same answer as at cast time: the stack has to be empty for a step
// to end, so the spell resolves in the step it was cast in.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97435e03-2fff-4fe0-8fa9-69ab0a046a33",
		Name:         "Unbreakable Formation",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			// The vigilance waits for a prompt the counters ask (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				if err := (effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"indestructible"},
					Label:    "Unbreakable Formation — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				if !effects.B14InYourMainPhase(ctx.Game, ctx.Controller()) {
					return nil
				}
				return effects.B11PutCounterOnEachCreatureYouControl(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !effects.B14InYourMainPhase(ctx.Game, ctx.Controller()) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"vigilance"},
					Label:    "Unbreakable Formation — vigilance (addendum)",
				}.Apply(ctx)
			}),
		),
	})
}
