package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// FlickerTheModesTarget is a "choose one or both" bullet "Exile target
// creature you control, then return it to the battlefield under its owner's
// control" as a ModeDoing body (Go Ninja Go, Settle Beyond Reality): the
// bullet's own target, read per occurrence (ModeTarget), flickered with no
// Controller so it returns under its owner's control. A target that has
// become illegal is skipped (CR 608.2b).
func FlickerTheModesTarget(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return Flicker{Target: t.ID}.Apply(ctx)
}
