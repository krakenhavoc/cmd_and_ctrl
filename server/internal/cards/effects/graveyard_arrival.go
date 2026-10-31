package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_arrival.go — "put into your graveyard from anywhere" (#4721).
//
// library_to_graveyard.go reads the one route out of a library. A clause
// that says "from anywhere", or "from anywhere other than the battlefield",
// has to read every route a card can take into a graveyard:
//
//   - a discard is an EventDiscardCard (hand to graveyard) and nothing else;
//   - a mill is an EventMill (library to graveyard) and nothing else;
//   - everything else, a spell that resolves or is countered, a surveil, a
//     search, a card pulled out of exile or a death, is an EventZoneMove
//     with the zone pair stamped on it. A death is also an EventLTB, which
//     carries no OldZone and is not an arrival here.
//
// Each arrival is announced by exactly one of those three kinds, so a
// reader that watches all three counts a card once. graveyardArrivals is
// the description of the clause (which source zones, which cards) and the
// two reads of it: the per-event test a trigger's AppliesTo asks, and the
// per-batch list a "one or more" trigger's effect asks as it resolves.
//
// What counts as one occurrence is the engine's event batch (event_batch.go,
// ADR 0049's amendment of 2026-09-17 and its 2026-10-30 one for this seam,
// CR 603.2c): everything one resolution or one turn-based action puts into a
// graveyard, so two creature cards discarded by one spell, or a creature that
// dies beside a creature card that is milled in the same resolution, are one
// batch. Separate resolutions are separate batches.
//
// A token is not a card (CR 108.2b, 111.6) and so is never an arrival read
// here, even in the moment it spends in the graveyard before CR 704.5d removes
// it (CR 111.7). A clause that does count tokens, "an artifact is put into
// your graveyard from the battlefield", reads the leaves-the-battlefield event
// instead (WhenAnotherArtifactIsPutIntoYourGraveyard).

// graveyardArrivalKinds are the three event kinds an arrival in a graveyard,
// from any zone, is announced by.
var graveyardArrivalKinds = []game.EventKind{game.EventMill, game.EventZoneMove, game.EventDiscardCard}

// graveyardArrivals describes the cards a "put into your graveyard from ..."
// clause counts. The zero value is every card, from anywhere.
type graveyardArrivals struct {
	// From lists the zones the card came from; empty means from anywhere.
	From []game.ZoneKind
	// Except lists zones it must not have come from: "from anywhere other
	// than the battlefield".
	Except []game.ZoneKind
	// Match filters the card as it sits in the graveyard, so a card is judged
	// by its printed card types (CR 603.6c: "from anywhere" looks at the new
	// object, never at what it was on the battlefield). Nil accepts any.
	Match func(game.Card) bool
	// Present, for the batch read, keeps only an arrival whose card is still in
	// that graveyard as the same object: the card's most recent zone change
	// this turn is the arrival (CR 400.7). The default counts every arrival,
	// which is what "that many" counts.
	Present bool
}

// arrival reports whether ev is a card owned by `owner` arriving in its
// graveyard that the clause counts.
func (a graveyardArrivals) arrival(ev game.Event, g *game.Game, owner uuid.UUID) bool {
	switch ev.Kind {
	case game.EventMill, game.EventZoneMove, game.EventDiscardCard:
	default:
		return false
	}
	if ev.NewZone != game.ZoneGraveyard || ev.OldZone == "" || ev.OldZone == game.ZoneGraveyard {
		return false
	}
	if len(a.From) > 0 && !slices.Contains(a.From, ev.OldZone) {
		return false
	}
	if slices.Contains(a.Except, ev.OldZone) {
		return false
	}
	// A graveyard is always its owner's, so ownership reads the card.
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.IsToken() || c.Owner != owner {
		return false
	}
	return a.Match == nil || a.Match(c)
}

// When is the clause as a trigger's AppliesTo: ev is a card the source's
// controller owns arriving in their graveyard. Watch graveyardArrivalKinds.
func (a graveyardArrivals) When() When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return a.arrival(ev, g, source.Controller)
	}
}

// InBatch is the cards `owner` had put into their graveyard in event batch
// `batch` that the clause counts, one entry per arrival, in the order they
// landed. A trigger's effect reads it as it resolves, from the triggering
// event's Batch, so the first arrival of a batch fires the ability and every
// later arrival of that batch is in the list.
func (a graveyardArrivals) InBatch(g *game.Game, batch uint64, owner uuid.UUID) []uuid.UUID {
	events := g.EventsThisTurn()
	var lastMove map[uuid.UUID]int
	if a.Present {
		lastMove = map[uuid.UUID]int{}
		for i, ev := range events {
			if ev.CardID != uuid.Nil && ev.OldZone != "" && ev.NewZone != "" {
				lastMove[ev.CardID] = i
			}
		}
	}
	var out []uuid.UUID
	for i, ev := range events {
		if ev.Batch != batch {
			continue
		}
		if !a.arrival(ev, g, owner) {
			continue
		}
		if a.Present {
			if lastMove[ev.CardID] != i {
				continue
			}
			if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
		}
		out = append(out, ev.CardID)
	}
	return out
}

// creatureCardsToYourGraveyard is "creature cards are put into your
// graveyard from anywhere" (Voracious Brood).
var creatureCardsToYourGraveyard = graveyardArrivals{Match: isCreatureCard}

// artifactCardsToYourGraveyardNotFromTheBattlefield is the second half of
// "an artifact card is put into your graveyard from anywhere other than the
// battlefield" (Ultron's Auxiliary, Ultron the Annihilator).
var artifactCardsToYourGraveyardNotFromTheBattlefield = graveyardArrivals{
	Except: []game.ZoneKind{game.ZoneBattlefield},
	Match:  isArtifactCard,
}

// isArtifactCard is the artifact filter the arrival readers are asked with.
func isArtifactCard(c game.Card) bool { return c.IsArtifact() }

// WhenAnotherArtifactIsPutIntoYourGraveyard is "Whenever another artifact is
// put into your graveyard from the battlefield or an artifact card is put
// into your graveyard from anywhere other than the battlefield" (Ultron's
// Auxiliary, Ultron the Annihilator): one trigger per object, never batched
// (CR 603.2c lets an ability trigger repeatedly when one event holds several
// occurrences).
//
// The first clause is a death of a permanent that was an artifact as it last
// existed on the battlefield (CR 603.10a, leftAsType), whoever controlled it,
// a token included: a token is put into the graveyard before it ceases to
// exist (CR 111.7, 704.5d). "Another" is the source itself, which the
// ability's own clause does not name. The second is a card arrival read by
// graveyardArrivals, which a token can never be (CR 108.2b), the permanent's
// own battlefield death excluded because the first clause already took it.
func WhenAnotherArtifactIsPutIntoYourGraveyard(label string, effect Effect) game.TriggeredAbility {
	kinds := append([]game.EventKind{game.EventLTB}, graveyardArrivalKinds...)
	cards := artifactCardsToYourGraveyardNotFromTheBattlefield.When()
	return OnAny(kinds, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventLTB {
			return cards(ev, source, lki, g)
		}
		if ev.NewZone != game.ZoneGraveyard || ev.CardID == source.InstanceID {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.Owner == source.Controller && leftAsType(ev, c, "artifact")
	}, label, effect)
}
