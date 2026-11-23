package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tapTheModesTargets is a modal bullet's "tap <N> target creatures": every
// target of the chosen occurrence that is still legal as the spell resolves
// is tapped (CR 608.2b), the others are skipped. "Up to two" targets with
// none chosen taps nothing (Urza's Rebuff).
func TapTheModesTargets(_ *game.StackItem, ctx *Context, occ int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
