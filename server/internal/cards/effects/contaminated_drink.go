package effects

// Contaminated Drink — Instant {X}{U}{B}:
//
//	"Draw X cards, then you get half X rad counters, rounded up."
//
// #2042. Cast for X=0 it draws nothing and gives nothing.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ee327c5e-ed29-4de3-bf50-92488f51a3a5",
		Name:         "Contaminated Drink",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: ResolveSteps(
			// The rad counters wait for a prompt the draws ask (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				return ctx.Game.DrawNForEffect(ctx.Controller(), x)
			}),
			Instruction(func(ctx *Context) error {
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				return playerGetsRadCounters(ctx.Game, ctx.Controller(), ctx.Controller(), (x+1)/2)
			}),
		),
	})
}
