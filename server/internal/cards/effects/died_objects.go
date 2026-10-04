package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// died_objects.go — "exile it" / "return this card" read about a card
// that went from the battlefield to a graveyard while its ability
// waited (S58, Illicit Masquerade and Avatar Destiny). Append-only.
//
// CR 400.7: the card in the graveyard is a new object, and it stays the
// object the ability is about only until it moves again. MoveCard bumps
// Card.ObjectEpoch once per zone change, so the graveyard object that
// arrived straight from the battlefield has exactly the battlefield
// epoch plus one. A card reanimated, exiled or bounced and then put back
// in response has moved more than once and is a different object.

// inGraveyardOneMoveAfter reports whether `id` sits in a graveyard as
// the object that arrived there straight from the battlefield object
// whose epoch was `battlefieldEpoch`.
func inGraveyardOneMoveAfter(g *game.Game, id uuid.UUID, battlefieldEpoch int) bool {
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return false
	}
	card, ok := g.LookupCardForEffect(id)
	return ok && card.ObjectEpoch == battlefieldEpoch+1
}

// diedCardStillInGraveyard is "it" in a dies trigger's "exile it": the
// creature the trigger is about, still in a graveyard as the card that
// died. With no CR 603.10 snapshot on the trigger (a hand-built item),
// being in a graveyard is all that can be asked.
func diedCardStillInGraveyard(ctx *Context, dead uuid.UUID) bool {
	obj := ctx.Trigger().Object
	if obj == nil {
		z := ctx.Game.FindCardZoneForEffect(dead)
		return z != nil && z.Kind == game.ZoneGraveyard
	}
	return inGraveyardOneMoveAfter(ctx.Game, dead, obj.Epoch)
}

// sourceCardInGraveyardAfterLeaving is "this card" in an ability that
// triggered while its source was on the battlefield and resolves after
// the source was put into a graveyard — an Aura that fell off the
// creature whose death it watched (Avatar Destiny).
func sourceCardInGraveyardAfterLeaving(ctx *Context) bool {
	ref, ok := ctx.SourceRef()
	if !ok {
		return false
	}
	return inGraveyardOneMoveAfter(ctx.Game, ref.ID, ref.Epoch)
}
