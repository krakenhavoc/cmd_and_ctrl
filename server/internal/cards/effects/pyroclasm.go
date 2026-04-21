package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyroclasm — "Pyroclasm deals 2 damage to each creature." Iterated
// DealDamage over every battlefield creature. Each DealDamage emits
// its own EventDealDamage; the lethal-damage SBA runs at the next
// priority-grant boundary (already bookended by the resolution
// path) to reap any creatures the damage killed.
func init() {
	Register(Spec{
		OracleID: "e4bcd4ea-e7cd-4471-8f3b-18bb51d3d70c",
		Name:     "Pyroclasm",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, id := range ctx.CreatureIDs() {
				if err := (DealDamage{
					Source: ctx.Source(),
					Target: id,
					Amount: 2,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
