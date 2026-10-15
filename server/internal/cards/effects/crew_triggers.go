package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// crew_triggers.go — "whenever this creature crews a Vehicle" and
// "saddles a Mount" (ADR 0186 §2, #3545; CR 702.122b, CR 702.171c).
//
// The two events fire while the cost is paid, one per creature tapped
// (game/activated.go), so a trigger built here goes on the stack ABOVE
// the crew or saddle ability and resolves first (CR 603.3b, and CR
// 208.3a's example: the Vehicle is not yet a creature when the trigger
// resolves).

// WhenThisCrewsOrSaddles builds "Whenever this creature crews a Vehicle",
// "… saddles a Mount" or "… saddles a Mount or crews a Vehicle",
// according to which of crew and saddle are set. It watches the kinds
// asked for and matches only an event whose creature is the source.
//
// The built item carries the Vehicle or Mount the creature paid for on
// Params.Object, as the object it was at the time (instance ID and
// epoch), which is what "that Mount or Vehicle" reads: an effect that
// finds a different object under that ID — the Vehicle was flickered in
// response — does nothing (CR 400.7). Read it with CrewedOrSaddledObject.
//
// effect is the resolution; set Targets on the result for a targeted
// trigger, exactly as on any other. A "during your main phase" clause is
// DuringYourMainPhase, not part of this builder.
func WhenThisCrewsOrSaddles(label string, crew, saddle bool, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	var kinds []game.EventKind
	if crew {
		kinds = append(kinds, game.EventCrewedVehicle)
	}
	if saddle {
		kinds = append(kinds, game.EventSaddledMount)
	}
	return game.TriggeredAbility{
		Watches: kinds,
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return source != nil && ev.CardID == source.InstanceID
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			epoch := -1
			if c, ok := g.LookupCardForEffect(ev.Target); ok {
				epoch = c.ObjectEpoch
			}
			item.Params.Object = game.ObjectRef{ID: ev.Target, Epoch: epoch}
			return item
		},
		Effect: effect,
	}
}

// DuringYourMainPhase narrows a trigger to events that happen during a
// main phase of its controller's own turn ("… during your main phase").
// It is a condition on the EVENT, checked when the event happens, and not
// an intervening "if" rechecked at resolution: the phase cannot end while
// the trigger waits on the stack, so the two read the same here.
func DuringYourMainPhase(t game.TriggeredAbility) game.TriggeredAbility {
	inner := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		if source == nil || (inner != nil && !inner(ev, source, lki, g)) {
			return false
		}
		return b14InYourMainPhase(g, source.Controller)
	}
	return t
}

// CrewedOrSaddledObject is "that Mount or Vehicle" for an item built by
// WhenThisCrewsOrSaddles: the permanent's instance ID, and false when it
// has since left the battlefield or come back as a new object.
func CrewedOrSaddledObject(g *game.Game, item *game.StackItem) (uuid.UUID, bool) {
	ref := item.Params.Object
	if ref.ID == uuid.Nil {
		return uuid.Nil, false
	}
	c, ok := g.LookupCardForEffect(ref.ID)
	if !ok || c.ObjectEpoch != ref.Epoch || !onBattlefield(g, ref.ID) {
		return uuid.Nil, false
	}
	return ref.ID, true
}
