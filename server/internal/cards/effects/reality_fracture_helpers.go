package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_helpers.go — small helpers shared by the Reality
// Fracture (FRA / FRC) cards. Append-only.

// rfCardsMilledThisTurn is the number of cards that were put into
// `owner`'s graveyard from their library this turn (Cruel
// Calculations). It reads this turn's events, which are bounded at the
// real turn boundary, and counts one per move: a card that left the
// graveyard and was milled again counts twice.
//
// A mill announces itself as an EventMill; other library-to-graveyard
// routes announce a plain EventZoneMove. A route that emitted both for
// one card would count it twice, so an event naming the same card as
// the one counted immediately before it is the same move.
//
// Caller holds g.mu.
func rfCardsMilledThisTurn(g *game.Game, owner uuid.UUID) int {
	n := 0
	var lastCounted uuid.UUID
	var lastIdx = -2
	for i, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventMill && ev.Kind != game.EventZoneMove {
			continue
		}
		if ev.OldZone != game.ZoneLibrary || ev.NewZone != game.ZoneGraveyard {
			continue
		}
		if ev.CardID == lastCounted && i-lastIdx <= 2 {
			lastIdx = i
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if ok && c.Owner == owner {
			n++
			lastCounted, lastIdx = ev.CardID, i
		}
	}
	return n
}
