package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tuckTheModesTargetOnTopOfItsOwnersLibrary is "put target <permanent or
// card> on top of its owner's (or your) library" as a modal bullet's body
// (Primal Command, The Unicomplex): the target goes on top of its owner's
// library, and a bullet whose target is no longer legal does nothing (CR
// 608.2b).
func TuckTheModesTargetOnTopOfItsOwnersLibrary(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ctx.Game.TuckToLibraryForEffect(t.ID, false)
}
