package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tdm_omen_helpers.go — a body two Tarkir: Dragonstorm Omen halves
// share (#5175).

// effectThenDrawACard is the OnResolve "<sentence>. Draw a card." whose
// first sentence asks nothing: `effect` on the spell's still-legal
// targets (CR 608.2b), then the draw. Skimming Strike taps up to one
// target creature, Dusk Sight puts a +1/+1 counter on up to one.
func effectThenDrawACard(effect Effect) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if err := effect(ctx.Game, item); err != nil {
			return err
		}
		return DrawCards{N: 1}.Apply(ctx)
	}
}
