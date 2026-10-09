package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extended Absence — Instant {3}{B}:
//
//	"Exile target creature or planeswalker. Extended Absence deals 1
//	 damage to each opponent and you gain 1 life."
//
// Targeted exile, then the ping and the life. A spell whose only
// target is illegal on resolution does not resolve (CR 608.2b), so
// neither the exile nor the second sentence happens.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "973fb1e3-83d5-4763-bea8-799b426a8848",
		Name:         "Extended Absence",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := ExileFirstTarget(ctx.Game, item); err != nil {
				return err
			}
			if err := damageToEachOpponent(ctx.Game, item, 1); err != nil {
				return err
			}
			return GainLife{Amount: 1}.Apply(ctx)
		},
	})
}
