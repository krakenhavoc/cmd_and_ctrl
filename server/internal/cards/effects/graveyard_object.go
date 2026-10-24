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
	if !cardObjectStillInGraveyard(g, item) {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     cost,
		Question: question,
		Then:     graveyardReturnPaid,
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
