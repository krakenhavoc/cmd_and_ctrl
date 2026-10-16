package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shuffle_itself_from_graveyard.go — "When ~ is put into a graveyard
// from anywhere, shuffle it into its owner's library" (#2038, ADR 0113
// amendment 2026-10-16): the Judgment Incarnations (Hostility, Purity,
// Vigor, Dread), Serra Avatar and Worldspine Wurm. Append-only,
// mechanic-named.
//
// It is a TRIGGER, not the replacement Blightsteel Colossus prints
// (shuffle_into_library_instead.go). The card reaches the graveyard,
// so "dies" triggers see it die, and then this ability shuffles it
// away. The trigger is the titans' (WhenThisIsPutIntoAGraveyardFrom
// Anywhere, ADR 0113 §2 decision 7), so it fires from the graveyard
// however the card got there — killed, sacrificed, discarded, milled,
// countered — and never for a card a replacement sent elsewhere (Rest
// in Peace), because the harvest finds the card where it landed. CR
// 603.6c: an ability that triggers on a card being put into a zone
// "from anywhere" is never a leaves-the-battlefield ability, so it
// does not look back in time (CR 603.10a), and the card triggers with
// the abilities it has in the graveyard: it fires even if it had lost
// its abilities on the battlefield.
//
// The resolution, per CR 701.24c and its Guile example: the card is
// shuffled in only if it is still the object that was put into the
// graveyard (CR 400.7); if it has left — exiled in response — its
// owner shuffles their library anyway.

// WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn is the whole
// printed line. Its Build stamps the graveyard object the trigger is
// about (Params.Object) — the card as the harvest found it in the
// graveyard, not the permanent a dying card was, which is what the
// stack item's own SourceObject names for a death — so the resolution
// can tell it from a card that left the graveyard and came back.
func WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn(label string) game.TriggeredAbility {
	t := WhenThisIsPutIntoAGraveyardFromAnywhere(label, shuffleThisCardIntoItsOwnersLibrary)
	// A fill-in Build (ADR 0041 P9): it records the graveyard object and
	// leaves the Effect to the row.
	t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
		return item
	}
	return t
}

// shuffleThisCardIntoItsOwnersLibrary is "shuffle it into its owner's
// library": the card goes into its owner's library if it is still in
// the graveyard as the object the trigger saw, and that library is
// shuffled whether or not it did (CR 701.24c). The move is the ordinary
// tuck, so a commander's owner is offered the command zone (CR 903.9b),
// and the shuffle is its continuation, so it waits for that answer.
func shuffleThisCardIntoItsOwnersLibrary(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	owner := item.Controller
	if c, ok := g.LookupCardForEffect(id); ok && c.Owner != uuid.Nil {
		owner = c.Owner
	}
	if !thisCardStillInTheGraveyard(g, item) {
		return g.ShuffleLibraryForEffect(owner)
	}
	return g.TuckToLibraryThenForEffect(id, game.TuckOptions{}, func(g *game.Game, _ bool) error {
		return g.ShuffleLibraryForEffect(owner)
	})
}

// thisCardStillInTheGraveyard reports whether the trigger's card is in a
// graveyard as the object the trigger stamped (CR 400.7). An item with
// no stamp (built by hand) keeps the plain "is it in a graveyard"
// answer.
func thisCardStillInTheGraveyard(g *game.Game, item *game.StackItem) bool {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return false
	}
	if ref := item.Params.Object; ref.ID == id {
		return g.ObjectEpochForEffect(id) == ref.Epoch
	}
	return true
}
