package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			if err := ctx.Game.DrawNForEffect(ctx.Controller(), x); err != nil {
				return err
			}
			return playerGetsRadCounters(ctx.Game, ctx.Controller(), ctx.Controller(), (x+1)/2)
		},
	})
}
