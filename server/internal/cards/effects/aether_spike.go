package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aether Spike — Instant {1}{U}:
//
//	"Choose target spell. You get {E}{E} (two energy counters), then you
//	 may pay any amount of {E}. Counter that spell unless its controller
//	 pays {1} for each {E} paid this way."
//
// ADR 0129 §3 (#1995, owner decision 3): the caster pays through the
// pay_amount prompt; then the spell's controller is asked to pay {N},
// N the energy paid, through the counter-unless prompt that holds the
// table while the spell is on the stack. Paying no energy is a tax of
// {0}, which is always paid, so nothing is asked. The bot pays all its
// energy: every counter is another {1} to find.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4a85647-0b8c-40b2-a5d5-43c78877ce36",
		Name:         "Aether Spike",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (GetEnergy{N: 2}).Apply(ctx); err != nil {
				return err
			}
			return PayEnergyAmount{
				Question: "Aether Spike — pay any amount of {E}; its controller must pay {1} for each or it is countered",
				Unit:     game.PayAmountTax,
				Goal:     AsMuchAsYouCan,
				Then: func(ctx *Context, paid int) error {
					if paid <= 0 {
						return nil
					}
					return CounterUnlessPaid{
						StackID:  ctx.Item.Targets[0].ID,
						Cost:     fmt.Sprintf("{%d}", paid),
						Question: fmt.Sprintf("Aether Spike — pay {%d}, or your spell is countered?", paid),
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
