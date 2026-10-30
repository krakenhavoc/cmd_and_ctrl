package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_to_graveyard.go — "put into your graveyard from your library",
// which is wider than "milled" (#2809).
//
// A mill (CR 701.17a) emits EventMill per card that lands in a
// graveyard. Every other library-to-graveyard move — surveil's
// graveyard leg (CR 701.25a), a search that puts the card there
// (Entomb), dredge's milled cards are mills — emits an ordinary
// EventZoneMove with the same zone pair. A card that says "milled"
// watches EventMill alone; a card that says "put into your graveyard
// from your library" watches both kinds and reads the zone pair, so a
// surveil or an Entomb counts as printed. One event per arrival, never
// two (zone_route.go), so the two kinds never count a card twice.

// libraryToGraveyardKinds are the two event kinds a library-to-graveyard
// arrival is announced by.
var libraryToGraveyardKinds = []game.EventKind{game.EventMill, game.EventZoneMove}

// cardPutIntoYourGraveyardFromYourLibrary reports whether ev is a card
// the source's controller owns arriving in their graveyard from their
// library, by any route, that `match` accepts. A graveyard is always
// its owner's, so ownership reads the destination.
func cardPutIntoYourGraveyardFromYourLibrary(ev game.Event, source *game.Card, g *game.Game, match func(game.Card) bool) bool {
	return libraryArrivals(match).When()(ev, source, game.Characteristic{}, g)
}

// cardsPutIntoGraveyardFromLibraryInBatch is the cards `owner` put into
// their graveyard from their library in event batch `batch` that
// `match` accepts, in the order they landed, and only those whose most
// recent zone change this turn is still that move — still the same
// object in the graveyard (CR 400.7). It is graveyardArrivals
// (graveyard_arrival.go) narrowed to the library and to the present.
func cardsPutIntoGraveyardFromLibraryInBatch(g *game.Game, batch uint64, owner uuid.UUID, match func(game.Card) bool) []uuid.UUID {
	return libraryArrivals(match).InBatch(g, batch, owner)
}

// libraryArrivals is "put into your graveyard from your library" for the
// cards `match` accepts, counting only those still there.
func libraryArrivals(match func(game.Card) bool) graveyardArrivals {
	return graveyardArrivals{From: []game.ZoneKind{game.ZoneLibrary}, Match: match, Present: true}
}

// isCreatureCard is the creature filter the library batch is read with
// (isLandCard, enters_only_if.go, is the land one).
func isCreatureCard(c game.Card) bool { return c.IsCreature() }
