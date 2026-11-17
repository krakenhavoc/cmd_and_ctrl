package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// counter_unless_activated.go — activated-ability bodies for "Counter
// target spell unless its controller pays <cost>." First written for the
// Spiketail Drakes (Prophecy, #5617).
//
// Append-only: add a builder, never change what one means.

// counterTheTargetSpellUnlessPaid is an activated ability's "Counter
// target spell unless its controller pays <cost>." through the shared
// CounterUnlessPaid prompt: the spell's controller is asked as the
// ability resolves, the stack waits for the answer (#951), and a "pay"
// they can't fund is a decline. A spell that has left the stack by then
// makes the ability do nothing (CR 608.2b).
func counterTheTargetSpellUnlessPaid(name, cost string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 {
			return nil
		}
		return CounterUnlessPaid{
			StackID:  targets[0].ID,
			Cost:     cost,
			Question: name + " — pay " + cost + " or your spell is countered",
		}.Apply(ctx)
	}
}
