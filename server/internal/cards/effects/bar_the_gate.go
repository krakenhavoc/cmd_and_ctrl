package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bar the Gate — Instant {2}{U}:
//
//	"Counter target creature or planeswalker spell. Venture into the
//	 dungeon."
//
// A spell whose only target has become illegal does not resolve at all
// (CR 608.2b), so it ventures only when it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11f55ec5-b9a3-4670-8afb-31d29478fb59",
		Name:         "Bar the Gate",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target creature or planeswalker spell", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 {
				if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return VentureIntoTheDungeon{}.Apply(ctx)
		},
	})
}
