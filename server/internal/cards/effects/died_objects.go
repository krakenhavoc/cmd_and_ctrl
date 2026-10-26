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

// cardTheCostMovedToAPublicZone is the object a permanent a cost
// sacrificed became in the zone it went to (#2099; CR 400.7j: "if the
// cost of a spell or ability causes an object to move to a public zone,
// that spell or ability's effects can find that object"). `sacrificed`
// is the battlefield object, from PaidCost.SacrificedObjects. The object
// is the card one zone change later, in a graveyard, or in exile or the
// command zone when a replacement sent it there instead.
//
// ok is false when there is no such object any more: a token, which has
// ceased to exist (CR 111.7); a card put into a hidden zone, which no
// effect can follow; and a card that has moved again since, even back
// into the same zone, which is a new object (CR 400.7). A sacrificed
// commander its owner moved to the command zone under CR 903.9a is the
// last of these: the move to the graveyard was the first zone change and
// the state-based action is the second.
func cardTheCostMovedToAPublicZone(g *game.Game, sacrificed game.ObjectRef) (game.ObjectRef, bool) {
	next := game.ObjectRef{ID: sacrificed.ID, Epoch: sacrificed.Epoch + 1}
	_, ok := objectStillInAPublicZone(g, next)
	return next, ok
}

// objectStillInAPublicZone reports the zone `ref` is in, if its card is
// still that object (CR 400.7) in a graveyard, in exile or in the
// command zone, and is not a token. The zone is what a return has to
// name as the place the card is put from.
func objectStillInAPublicZone(g *game.Game, ref game.ObjectRef) (game.ZoneKind, bool) {
	if ref.ID == uuid.Nil {
		return "", false
	}
	z := g.FindCardZoneForEffect(ref.ID)
	if z == nil {
		return "", false
	}
	switch z.Kind {
	case game.ZoneGraveyard, game.ZoneExile, game.ZoneCommand:
	default:
		return "", false
	}
	c, ok := g.LookupCardForEffect(ref.ID)
	if !ok || c.IsToken() || c.ObjectEpoch != ref.Epoch {
		return "", false
	}
	return z.Kind, true
}

// returnTheObjectsTogetherUnderYourControl is the delayed body "return
// that card and the sacrificed card to the battlefield under your
// control" (Rescue from the Underworld, #2099): every object Objects
// names that is still that object in the public zone it was in when the
// trigger was created goes onto the battlefield under the trigger's
// controller, all of them as one event (CR 603.6a). One that has left
// that zone stays where it is (CR 603.7c); the rest still return.
func returnTheObjectsTogetherUnderYourControl(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	entries := make([]game.BatchEntry, 0, len(p.Objects))
	for _, ref := range p.Objects {
		from, ok := objectStillInAPublicZone(g, ref)
		if !ok {
			continue
		}
		if c, _ := g.LookupCardForEffect(ref.ID); !c.IsPermanent() {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: ref.ID, From: from})
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{Controller: item.Controller}, game.KeyedTail{})
}
