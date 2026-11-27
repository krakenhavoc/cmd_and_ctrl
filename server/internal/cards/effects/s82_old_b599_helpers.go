package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b599_helpers.go — shared bodies for Betrayers of Kamigawa,
// batch 599 (#5602).

// ReturnObjectToHandBody is the delayed trigger "return <the object named
// in Params.Object> to its owner's hand (return it only if it's on the
// battlefield)": Wall of Junk's own creature at end of combat, Kaijin of the
// Vanishing Touch's blocked one. It does nothing when the object has gone
// or come back as a new one (CR 400.7). Its key is on disk.
var ReturnObjectToHandBody = game.SimpleDelayedBody("s82-gap-b134/return-source-to-hand", func(g *game.Game, item *game.StackItem) error {
	obj := item.Params.Object
	if !OnBattlefield(g, obj.ID) || g.ObjectEpochForEffect(obj.ID) != obj.Epoch {
		return nil
	}
	return BounceToHand{Target: obj.ID}.Apply(NewContext(g, item))
})

// SacrificeThisIfStillThere is the "otherwise" of a sacrifice-unless
// (Charnel Troll, Chisei): the ability's source is sacrificed only if it is
// still on the battlefield and still the same object (CR 400.7).
func SacrificeThisIfStillThere(ctx *Context) error {
	if !OnBattlefield(ctx.Game, ctx.Source()) || SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
}
