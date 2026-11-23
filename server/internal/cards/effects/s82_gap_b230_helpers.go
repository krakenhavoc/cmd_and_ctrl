package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b230_helpers.go — card-side helpers of set GAP, batch 230
// (Commander Legends: Battle for Baldur's Gate, S82, #5948).

// DestroyTargetThenScry is "Destroy target <thing>. Scry N." (Guiding Bolt):
// the first target is destroyed if it is still legal, and the controller
// scries whether or not anything was destroyed. DestroyTargetThenSurveil's
// twin.
//
//nolint:unparam // the printed scry count, kept at the call site
func DestroyTargetThenScry(n int) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard && ctx.IsTargetLegal(item.Targets[0]) {
			if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return Scry{Player: ctx.Controller(), N: n}.Apply(ctx)
	}
}
