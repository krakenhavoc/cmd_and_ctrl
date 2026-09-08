package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exsanguinate — Sorcery for {X}{B}{B}:
//
//	"Each opponent loses X life. You gain life equal to the life
//	lost this way."
//
// S20 sub-PR 3: an untargeted X spell. "Life lost this way" is the
// sum actually lost — a player at 2 facing X=5 loses 2, and
// ChangePlayerLifeForEffect reports the delta it applied, so the
// gain matches paper even at low life totals.
func init() {
	Register(Spec{
		OracleID: "8164b1e8-3350-465e-8a17-75f57d326344",
		Name:     "Exsanguinate",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			lost := 0
			for _, opp := range ctx.Opponents() {
				p := ctx.PlayerByID(opp)
				if p == nil {
					continue
				}
				before := p.Life
				if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), opp, -x); err != nil {
					return err
				}
				lost += before - p.Life
			}
			if lost > 0 {
				return GainLife{Player: ctx.Controller(), Amount: lost}.Apply(ctx)
			}
			return nil
		},
	})
}
