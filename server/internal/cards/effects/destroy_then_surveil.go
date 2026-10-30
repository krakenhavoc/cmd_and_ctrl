package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// destroyTargetThenSurveil is "Destroy target <thing>. Surveil N." (Hour of
// Defeat, Your Fate Ends Here): the first target is destroyed if it is still
// legal, and the controller surveils whether or not anything was destroyed
// (an indestructible target survives and the surveil still happens).
func destroyTargetThenSurveil(n int) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard && ctx.IsTargetLegal(item.Targets[0]) {
			if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return Surveil{Player: ctx.Controller(), N: n}.Apply(ctx)
	}
}
