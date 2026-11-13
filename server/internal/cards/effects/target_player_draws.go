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
