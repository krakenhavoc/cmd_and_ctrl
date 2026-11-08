package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// left_graveyard.go — "a card leaves a graveyard", every way it can
// (#4899). Three routes carry one, and only two event kinds:
//
//   - an EventZoneMove out of a graveyard: an exile, a regrowth, a
//     reanimation, a graveyard sweep, and a land PLAYED from the
//     graveyard (a Crucible of Worlds play is a Played zone move to the
//     battlefield);
//   - an EventCast whose OldZone is the graveyard: a flashback, an
//     escape, a cast permission. The cast path moves the card to the
//     stack without a zone-move event of its own and stamps the
//     zone-move fields on the cast instead (CR 601.2a), so a trigger
//     that watches only EventZoneMove misses every cast.
//
// A trigger reading these watches leftGraveyardEvents. "One or more
// cards leave …" is still OncePerBatch over the predicate: a cast is its
// own event, and cards leaving as one event are one batch.

// leftGraveyardEvents are the event kinds a "leaves a graveyard"
// trigger watches.
var leftGraveyardEvents = []game.EventKind{game.EventZoneMove, game.EventCast}

// cardThatLeftAGraveyard is the card `ev` took out of a graveyard, read
// where it went: any graveyard, any route above. A token is not a card
// (CR 108.2), so one ceasing to exist in a graveyard is not counted.
func cardThatLeftAGraveyard(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventZoneMove && ev.Kind != game.EventCast {
		return game.Card{}, false
	}
	if ev.OldZone != game.ZoneGraveyard || ev.NewZone == game.ZoneGraveyard || ev.CardID == uuid.Nil {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || IsToken(c) {
		return game.Card{}, false
	}
	return c, true
}

// cardThatLeftYourGraveyard is cardThatLeftAGraveyard for "your
// graveyard": the one the source's controller OWNS.
func cardThatLeftYourGraveyard(ev game.Event, source *game.Card, g *game.Game) (game.Card, bool) {
	if source == nil {
		return game.Card{}, false
	}
	c, ok := cardThatLeftAGraveyard(ev, g)
	if !ok || c.Owner != source.Controller {
		return game.Card{}, false
	}
	return c, true
}

// cardLeftYourGraveyard is "a card leaves your graveyard" (Tormod,
// Quintorius, Teval's Judgment).
func cardLeftYourGraveyard(ev game.Event, source *game.Card, g *game.Game) bool {
	_, ok := cardThatLeftYourGraveyard(ev, source, g)
	return ok
}

// creatureCardLeftYourGraveyard is "a creature card leaves your
// graveyard" (Desecrated Tomb, Syr Konrad), asked of the card as it was
// in the graveyard (wasCardInAGraveyard): a spell cast from it can be
// another face of the card.
func creatureCardLeftYourGraveyard(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := cardThatLeftYourGraveyard(ev, source, g)
	return ok && wasCardInAGraveyard(c, game.Card.IsCreature)
}

// wasCardInAGraveyard asks a type question of a card that has just LEFT
// a graveyard, as the card it was there (CR 603.10a: a
// leaves-the-graveyard trigger looks back in time).
//
// The card is looked up where it went, and for a single-faced card that
// is the same answer. A multi-faced one can have changed: a modal
// double-faced card cast from a graveyard as its back face, or an
// adventurer card cast as its Adventure, is on the stack as the half
// that was cast. In a graveyard it had only its front face's
// characteristics (CR 712.8a, 715.4), or, for a split card, both halves'
// together (CR 709.4), so those are what is asked here — never the half
// it left as.
func wasCardInAGraveyard(c game.Card, is func(game.Card) bool) bool {
	if !c.IsMultiFace() {
		return is(c)
	}
	faces := c.Faces[:1]
	if game.IsSplitCard(c) {
		faces = c.Faces
	}
	for _, f := range faces {
		if is(game.Card{TypeLine: f.TypeLine}) {
			return true
		}
	}
	return false
}

// whenCardsLeaveYourGraveyard is cardLeftYourGraveyard as a trigger's
// When: "whenever one or more cards leave your graveyard" under
// OncePerBatch (Quintorius, Spirit of Resilience).
func whenCardsLeaveYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return cardLeftYourGraveyard(ev, source, g)
}

// whenCardsLeaveYourGraveyardDuringYourTurn is whenCardsLeaveYourGraveyard
// with "during your turn" as part of the trigger condition, so cards that
// leave on an opponent's turn do nothing (Attuned Hunter, Kheru
// Goldkeeper, Kishla Skimmer).
func whenCardsLeaveYourGraveyardDuringYourTurn(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return IsYourTurn(g, source.Controller) && whenCardsLeaveYourGraveyard(ev, source, lki, g)
}
