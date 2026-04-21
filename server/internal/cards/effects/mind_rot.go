package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Rot — "Target player discards two cards." Sandbox: "the
// player chooses" which two is replaced by a uniform-random pick
// (S14's DiscardCards primitive is always random). Casual targets
// self-police if they want to choose; the target-picker UX lands
// in S22.
func init() {
	Register(Spec{
		OracleID:   "ad44cf74-b717-48fb-9fa2-77512024d76a",
		Name:       "Mind Rot",
		TargetMode: "player",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return DiscardCards{Player: item.Targets[0].ID, N: 2}.Apply(ctx)
		},
	})
}
