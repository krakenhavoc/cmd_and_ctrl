package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB485TapModeTargets is a mode's "Tap <the chosen targets>": every
// target of the occurrence that is still legal as the spell resolves is
// tapped (CR 608.2b), in announce order. Reality Spasm's first bullet.
func S82OldB485TapModeTargets(_ *game.StackItem, ctx *Context, occ int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB485UntapModeTargets is S82OldB485TapModeTargets' twin for "Untap
// <the chosen targets>": Reality Spasm's second bullet.
func S82OldB485UntapModeTargets(_ *game.StackItem, ctx *Context, occ int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
