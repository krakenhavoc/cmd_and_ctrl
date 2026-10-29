package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cardObjectStillInGraveyard is "this card" in the graveyard: the very
// object that triggered, not a card of the same name that left and came
// back (CR 400.7).
func cardObjectStillInGraveyard(g *game.Game, item *game.StackItem) bool {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return false
	}
	ref := item.SourceObject
	if ref.ID != item.SourceCardID || ref.ID == uuid.Nil {
		return true
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == ref.ID {
			return z.Cards[i].ObjectEpoch == ref.Epoch
		}
	}
	return false
}

// mayPayForGraveyardReturn offers a resolution-time payment only for the
// same graveyard object, and checks it again when the payment finishes.
func mayPayForGraveyardReturn(g *game.Game, item *game.StackItem, cost, question string) error {
	return mayPayForGraveyardReturnThen(g, item, cost, question, graveyardReturnPaid)
}

// mayPayForGraveyardReturnToHand is the same offer for a card that comes back
// to its owner's hand (Silvan Reveler's "you may pay {1}{G}{U}. If you do,
// return this card from your graveyard to your hand").
func mayPayForGraveyardReturnToHand(g *game.Game, item *game.StackItem, cost, question string) error {
	return mayPayForGraveyardReturnThen(g, item, cost, question, graveyardReturnToHandPaid)
}

func mayPayForGraveyardReturnThen(g *game.Game, item *game.StackItem, cost, question string, then game.PayUnlessElse) error {
	if !cardObjectStillInGraveyard(g, item) {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     cost,
		Question: question,
		Then:     then,
	}.Apply(NewContext(g, item))
}

// graveyardReturnPaid is the paid half of mayPayForGraveyardReturn, keyed (#4264). The asking item carries the card's object, so a card that left the graveyard and came back is a new object (CR 400.7).
var graveyardReturnPaid = PayUnlessBranches("pay-unless/graveyard-return-this-card-to-battlefield",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if !cardObjectStillInGraveyard(ctx.Game, ctx.Item) {
			return nil
		}
		return returnThisCardFromYourGraveyard(ctx.Game, ctx.Item)
	}, nil)

// graveyardReturnToHandPaid is graveyardReturnPaid's hand twin, keyed the same
// way (#4264): the object check is repeated once the mana is paid, so a card
// that left the graveyard and came back is a new object (CR 400.7).
var graveyardReturnToHandPaid = PayUnlessBranches("pay-unless/graveyard-return-this-card-to-hand",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if !cardObjectStillInGraveyard(ctx.Game, ctx.Item) {
			return nil
		}
		return returnThisCardFromYourGraveyardToHand(ctx.Game, ctx.Item)
	}, nil)
