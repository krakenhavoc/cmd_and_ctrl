package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "97435e03-2fff-4fe0-8fa9-69ab0a046a33",
		Name:         "Unbreakable Formation",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			yours := And(Creature(), YouControl())
			if err := (GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"indestructible"},
				Label:    "Unbreakable Formation — indestructible",
			}).Apply(ctx); err != nil {
				return err
			}
			if !b14InYourMainPhase(ctx.Game, ctx.Controller()) {
				return nil
			}
			if err := b11PutCounterOnEachCreatureYouControl(ctx.Game, item); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"vigilance"},
				Label:    "Unbreakable Formation — vigilance (addendum)",
			}.Apply(ctx)
		},
	})
}
