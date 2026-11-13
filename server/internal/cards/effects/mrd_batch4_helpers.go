package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// mrd_batch4_helpers.go — pieces shared by Mirrodin, batch 4 (#5606).
// Append-only.

// upkeepMayPayToUntapThis is "At the beginning of your upkeep, you may
// pay <cost>. If you do, untap this <permanent>": Goblin Dirigible and
// Goblin War Wagon, beside their "doesn't untap during your untap step".
// It is Brass Man's trigger: the trigger goes on the stack whether or not
// the permanent is tapped, the "you may pay" is asked as it resolves
// (MayPay), and `paid` is the card's own keyed paid branch, which untaps
// the permanent only while it is still on the battlefield
// (untapThisIfStillOnBattlefield).
func upkeepMayPayToUntapThis(name, cost string, paid game.PayUnlessElse) game.TriggeredAbility {
	return AtYourUpkeep(name+" — you may pay "+cost+" to untap it",
		func(g *game.Game, item *game.StackItem) error {
			return MayPay{
				Chooser:  item.Controller,
				Cost:     cost,
				Question: name + " — pay " + cost + " to untap it?",
				Then:     paid,
			}.Apply(NewContext(g, item))
		})
}

// untapThisIfStillOnBattlefield is the paid branch of
// upkeepMayPayToUntapThis: the item's source is untapped if it is still
// on the battlefield, and nothing happens once it has left.
func untapThisIfStillOnBattlefield(ctx *Context, _ game.PayUnlessAnswer) error {
	if !onBattlefield(ctx.Game, ctx.Item.SourceCardID) {
		return nil
	}
	return UntapTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
}
