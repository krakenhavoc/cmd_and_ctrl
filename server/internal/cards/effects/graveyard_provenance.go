package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_provenance.go — questions about HOW a card reached a
// graveyard this turn, read off this turn's events.
//
// The turn tally (#586) counts; it does not remember which card went
// where by what route, and a card in a graveyard carries no record of
// how it got there. So these are scans of g.EventsThisTurn(), which is
// bounded at the real turn boundary (#1009) — never anchored on the
// upkeep. A card keeps its instance ID across zone changes, so its
// arrivals are the events that name it with NewZone == graveyard.

// discardedOrMilledThisTurn reports whether `cardID`'s most recent
// arrival in a graveyard this turn was a DISCARD or a move out of a
// LIBRARY — Ghost of Ramirez DePietro's "a card in a graveyard that
// was discarded or put there from a library this turn".
//
// One arrival can be announced by more than one event (a discard is an
// EventDiscardCard beside the zone move; a mill an EventMill beside
// it), so every event of the same arrival is read together: the
// arrival counts if ANY of them says discard or library. A later
// departure from the graveyard ends that arrival, and only the last
// arrival matters — a card discarded, reanimated and then destroyed
// this turn was put there by the destruction.
//
// False for a card with no graveyard arrival this turn, including one
// that was already in the graveyard when the turn began.
//
// Caller holds g.mu.
func discardedOrMilledThisTurn(g *game.Game, cardID uuid.UUID) bool {
	inArrival, counts := false, false
	for _, ev := range g.EventsThisTurn() {
		if ev.CardID != cardID {
			continue
		}
		switch {
		case ev.NewZone == game.ZoneGraveyard && ev.OldZone != game.ZoneGraveyard:
			if !inArrival {
				inArrival, counts = true, false
			}
			if ev.Kind == game.EventDiscardCard || ev.OldZone == game.ZoneLibrary {
				counts = true
			}
		case ev.OldZone == game.ZoneGraveyard && ev.NewZone != "" && ev.NewZone != game.ZoneGraveyard:
			inArrival, counts = false, false
		}
	}
	return inArrival && counts
}

// descendedThisTurn is descend's "you descended this turn" (Brass's
// Tunnel-Grinder): a permanent card was put into `player`'s graveyard
// from anywhere this turn. A token is not a card (CR 108.2), so a token
// dying does not count. The card's types are read as it is now, which
// for a card that stayed in the graveyard is its printed front face;
// a card that has since left keeps its instance ID and its printed
// types, so it still counts.
//
// Caller holds g.mu.
func descendedThisTurn(g *game.Game, player uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard || ev.CardID == uuid.Nil {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if ok && c.Owner == player && isPermanentCard(c) {
			return true
		}
	}
	return false
}
