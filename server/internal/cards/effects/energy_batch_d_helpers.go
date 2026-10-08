package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// energy_batch_d_helpers.go — shared bodies for ADR 0129 PR 1's energy
// cards.

// youGetThatManyEnergy is "you get that many {E}" on a trigger whose
// event carries the amount: the combat damage dealt (Peema
// Trailblazer), the life gained (Sphinx of the Revelation).
func youGetThatManyEnergy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return GetEnergy{N: ctx.Trigger().Event.Amount}.Apply(ctx)
}
