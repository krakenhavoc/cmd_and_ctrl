package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// destroyTargetThenGainLife is "Destroy target <permanent>. You gain N
// life." (Sanctify, Radiant Strike): the destruction, then the life in
// printed order (CR 608.2c). A target that is no longer legal makes the
// spell do nothing, the life included (CR 608.2b).
func DestroyTargetThenGainLife(life int) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		return GainLife{Amount: life}.Apply(ctx)
	}
}
