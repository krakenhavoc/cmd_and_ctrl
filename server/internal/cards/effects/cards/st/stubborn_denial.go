package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stubborn Denial — Instant {U}:
//
//	"Counter target noncreature spell unless its controller pays {1}.
//	 Ferocious — If you control a creature with power 4 or greater, counter that
//	 spell instead."
//
// Spell Pierce's pay-unless with {1}, and the ferocious clause read as the
// spell resolves (CR 608.2c): with a creature of power 4 or greater (layered
// power) the spell is simply countered, with no payment offered; without one the
// spell's controller is asked for {1}. Ferocious is an ability word with no rules
// meaning of its own (CR 207.2c). The counter is "instead", so a spell that can't
// be countered is not countered by it, while the pay-unless branch never
// counters at all unless the payment is refused.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3adc3af3-5086-4d61-b38c-c1dddeaeebe6",
		Name:         "Stubborn Denial",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target noncreature spell", effects.Noncreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if effects.FerociousYouControlPowerFourOrGreater(ctx) {
				return effects.CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
			}
			return effects.CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{1}",
				Question: "Stubborn Denial — pay {1} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
