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
		OnPay: func(ctx *Context) error {
			if !cardObjectStillInGraveyard(ctx.Game, ctx.Item) {
				return nil
			}
			return returnThisCardFromYourGraveyard(ctx.Game, ctx.Item)
		},
	}.Apply(NewContext(g, item))
}
