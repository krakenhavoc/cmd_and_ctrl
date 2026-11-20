package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// target_player_gains_life.go — "Target player gains N life." as a
// spell's whole OnResolve. Append-only.

// targetPlayerGainsLife is "Target player gains N life." (Sylvan Bounty,
// Natural Spring #5630). A target player who is no longer legal as the
// spell resolves gains nothing (CR 608.2b).
func targetPlayerGainsLife(n int) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		if p, ok := firstLegalPlayerTarget(ctx); ok {
			return GainLife{Player: p, Amount: n}.Apply(ctx)
		}
		return nil
	}
}
