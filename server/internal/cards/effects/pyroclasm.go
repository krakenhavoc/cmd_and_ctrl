package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyroclasm — "Pyroclasm deals 2 damage to each creature."
//
// Damage, not destruction, and the distinction has consequences: an
// X/3 lives, damage prevention applies, and nothing dies until the
// lethal-damage state-based action runs at the next boundary. Since
// S23 that SBA sweep destroys its whole doomed set as one
// simultaneous event, so a Blood Artist that Pyroclasm killed sees
// every other creature it killed (see game/simultaneous.go).
func init() {
	Register(Spec{
		OracleID: "e4bcd4ea-e7cd-4471-8f3b-18bb51d3d70c",
		Name:     "Pyroclasm",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Creature(), 2)
		},
	})
}
