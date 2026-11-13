package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// target_player_draws.go — "Target player draws N cards." as a spell's
// whole OnResolve. Append-only.

// targetPlayerDrawsTwo is "Target player draws two cards." (Deep
// Analysis, Inspiration #5631). A target that is no longer legal never
// gets here (CR 608.2b).
func targetPlayerDrawsTwo(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	return DrawCards{Player: item.Targets[0].ID, N: 2}.Apply(ctx)
}

// theTargetPlayerDraws is "Target player draws N cards." as an
// activated row's Effect (Limestone Golem's "draws a card" is N = 1). A
// target that is no longer legal as the ability resolves draws nothing
// (CR 608.2b).
func theTargetPlayerDraws(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		p, ok := firstLegalPlayerTarget(ctx)
		if !ok {
			return nil
		}
		return DrawCards{Player: p, N: n}.Apply(ctx)
	}
}
