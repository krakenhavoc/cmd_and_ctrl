package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bro_prototype_helpers.go — bodies shared by the Brothers' War
// prototype cards (#5656). Each reads "this creature's power" as the
// ability resolves (CR 608.2h), which is what makes a prototyped
// permanent's smaller size count: the permanent as it is now, or as it
// last existed if it has left.

// broScryEqualToSourcePower is "scry X, where X is its power" (Spotter
// Thopter). A power of zero or less scries nothing.
func broScryEqualToSourcePower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	return Scry{Player: ctx.Controller(), N: info.Power}.Apply(ctx)
}

// broPumpTargetBySourcePower is "target … creature gets +X/+0 until end
// of turn, where X is this creature's power" (Iron-Craw Crusher). A
// power of zero or less, or an illegal target, does nothing.
func broPumpTargetBySourcePower(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		info, ok := ctx.SourcePermanent()
		if !ok || info.Power <= 0 {
			return nil
		}
		return vsPumpTheTarget(ctx, info.Power, 0, label)
	}
}
