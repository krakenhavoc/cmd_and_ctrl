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
				Then:     aetherSpikeCounterUnless,
			}.Apply(ctx)
		},
	})
}

// aetherSpikeCounterUnless is the paid half (#4294): the spell it chose is
// countered unless its controller pays {1} for each {E} paid, a tax of the
// amount, read off the asking item's target.
var aetherSpikeCounterUnless = NumberThen("number/aether-spike-counter-unless-paid",
	func(ctx *Context, r game.NumberAnswer) error {
		if r.Number <= 0 || ctx.Item == nil || len(ctx.Item.Targets) == 0 {
			return nil
		}
		return CounterUnlessPaid{
			StackID:  ctx.Item.Targets[0].ID,
			Cost:     fmt.Sprintf("{%d}", r.Number),
			Question: fmt.Sprintf("Aether Spike — pay {%d}, or your spell is countered?", r.Number),
		}.Apply(ctx)
	})
