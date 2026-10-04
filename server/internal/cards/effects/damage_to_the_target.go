package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// DealDamageToTheTarget is the Effect body for "<this> deals N damage to
// target X" with a printed N and one target (Fall of Cair Andros's
// {7}{R}). The source is the ability's own source. A target that left
// in response (CR 608.2b) leaves the ability to do nothing rather than
// error.
func DealDamageToTheTarget(amount int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}.Apply(ctx)
		}
		return nil
	}
}
