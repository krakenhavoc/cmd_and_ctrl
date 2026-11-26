package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawnglow Infusion — Sorcery {X}{G/W} (Shadowmoor, S82 OLD b529, #5591):
//
//	"You gain X life if {G} was spent to cast this spell and X life if {W} was
//	 spent to cast this spell. (Do both if {G}{W} was spent.)"
//
// The spell reads its own payment as it resolves (Ribbons of Night's
// ReadsOwnPayment): green spent gains X, white spent gains X, and a payment
// with both does both. The hybrid symbol is paid with either colour, so the
// cast declares both as wanted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "4db813f4-3bd1-4177-9cf7-6c9799a343c5",
		Name:           "Dawnglow Infusion",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"G", "W"},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			for _, color := range []string{"G", "W"} {
				if ctx.ManaSpent().Count(color) > 0 {
					if err := (effects.GainLife{Player: ctx.Controller(), Amount: x}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
