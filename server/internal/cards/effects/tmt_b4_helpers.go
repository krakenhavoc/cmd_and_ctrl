package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tmt_b4_helpers.go — shared bodies for Teenage Mutant Ninja Turtles,
// batch 4 (#4457).

// plusOneCounterOnThisThenItGains is "Put a +1/+1 counter on this
// creature. It gains <keywords> until end of turn." (Metalhead). The
// grant is its own step, so it waits for any prompt the counter asks (a
// CR 616 order) and lands on the permanent only while it is still the
// object the ability came from (CR 400.7, CR 608.2b).
func plusOneCounterOnThisThenItGains(label string, keywords ...string) Effect {
	return Do(
		Instruction(func(ctx *Context) error {
			return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		}),
		Instruction(func(ctx *Context) error {
			if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
				return nil
			}
			return GrantKeywordUntilEOT{Target: ctx.Item.SourceCardID, Keywords: keywords, Label: label}.Apply(ctx)
		}),
	)
}
