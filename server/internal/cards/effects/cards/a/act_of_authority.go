package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Act of Authority — Enchantment {1}{W}{W} (Commander 2013, S82 older sets):
//
//	"When this enchantment enters, you may exile target artifact or
//	 enchantment.
//	 At the beginning of your upkeep, you may exile target artifact or
//	 enchantment. If you do, its controller gains control of this
//	 enchantment."
//
// Both triggers choose their target as they go on the stack, only when the
// player takes the "may" (Optional over Targeting). The upkeep one reads the
// exiled permanent's controller before it leaves, and only when it was exiled
// does that player gain control of the enchantment, with no end (CR 611.2a);
// an Act that has left or come back as a new object is not handed over
// (CR 400.7). Exiling a permanent you control hands the enchantment to
// yourself, which changes nothing.
//
// No simplification.
func init() {
	const enters = "Act of Authority — you may exile target artifact or enchantment"
	const upkeep = "Act of Authority — you may exile target artifact or enchantment; if you do, its controller gains control of this enchantment"
	target := effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))
	effects.Register(effects.Spec{
		OracleID:     "c12b8e31-5702-4d3c-80c9-56c87e50679d",
		Name:         "Act of Authority",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters(enters, effects.ExileTheFirstLegalTarget), target),
				"Act of Authority — exile target artifact or enchantment?"),
			effects.Optional(effects.Targeting(effects.AtYourUpkeep(upkeep, actOfAuthorityUpkeep), target),
				"Act of Authority — exile target artifact or enchantment? If you do, its controller gains control of this enchantment."),
		},
	})
}

func actOfAuthorityUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		tail := ctx.ItemTail(actOfAuthorityHandOver)
		tail.Carry = []uuid.UUID{c.Controller}
		return g.ExileCardKeyedForEffect(t.ID, tail)
	}
	return nil
}

// actOfAuthorityHandOver is "If you do, its controller gains control of this
// enchantment": it runs once the exile has settled, and only when the
// permanent reached exile. The controller is the one it had as it left.
var actOfAuthorityHandOver = effects.CardTail("act-of-authority/hand-over", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	owner, g, item := o.Tail.Carry[0], ctx.Game, ctx.Item
	if owner == uuid.Nil || owner == item.Controller || effects.SourceIsNewObject(g, item) {
		return nil
	}
	g.GainControlForEffect(item.SourceCardID, item.SourceCardID, owner, game.IndefiniteDuration(),
		"Act of Authority — its controller gains control of this enchantment")
	return nil
})
