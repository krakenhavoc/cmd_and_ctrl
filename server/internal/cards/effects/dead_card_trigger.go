package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dead_card_trigger.go — a trigger that acts on "it": the card its event
// named, found again when the trigger resolves (#3143). Append-only, per
// the clone gate's rule. Ares' own copy of this shape predates it; The
// Master, Missy and Kolodin use these.

// eventCardTriggerItem is the fill-in Build for a trigger whose
// resolution acts on the card its event named (the creature that died or
// entered). It records the card and the object epoch it has after the
// event, which is the board read the resolution compares against to know
// "it" is still the same object (CR 400.7) — a card that left and came
// back is not.
func eventCardTriggerItem(label string) func(game.Event, *game.Card, game.Characteristic, *game.Game) *game.StackItem {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		epoch := -1
		if c, ok := g.LookupCardForEffect(ev.CardID); ok {
			epoch = c.ObjectEpoch
		}
		item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: epoch}
		return item
	}
}

// deadCardInGraveyard reports whether the card a eventCardTriggerItem
// named is still the object that died and is still in a graveyard, and
// returns its ID. A token (CR 111.7) and a card moved on in response
// answer false.
func deadCardInGraveyard(g *game.Game, item *game.StackItem) (uuid.UUID, bool) {
	dead := item.Params.Object
	c, ok := g.LookupCardForEffect(dead.ID)
	if !ok || c.ObjectEpoch != dead.Epoch {
		return uuid.Nil, false
	}
	z := g.FindCardZoneForEffect(dead.ID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return uuid.Nil, false
	}
	return dead.ID, true
}
