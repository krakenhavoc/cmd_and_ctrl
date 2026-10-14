package game

import "github.com/google/uuid"

// Event.PublicNames: the names a log line needs after the object is gone
// (#3609).
//
// The public log resolves a line's card against the viewer's CURRENT
// view, so a token that has since ceased to exist (CR 111.7, CR 704.5d)
// is in no zone and its lines read "a card". Nothing about a token is
// hidden — mintTokenLocked makes every seat a knower — so there is
// nothing to protect by forgetting it. The name rides the event that
// mentions the token, captured while the token can still be found, and
// the log falls back to it only when the view has no such card.
//
// It is carried on the event rather than in a per-game map for one
// reason: events already survive Clone, undo (RestoreFrom truncates the
// list), persistence and the log fold's cache, and a side table would
// have to be taught each of them.
//
// Only a token that is face up and known to every seat is captured. A
// hidden-zone or face-down card is never named here, which is what keeps
// the fallback fail-closed.

// publicNameOfLocked is the name a card may be printed under in the log
// after it has gone, or "" when it may not be. Caller must hold g.mu.
func (g *Game) publicNameOfLocked(c *Card) string {
	if c == nil || !IsTokenPtr(c) || c.FaceDown {
		return ""
	}
	for _, s := range g.Seats {
		if !c.IsKnownTo(s.ID) {
			return ""
		}
	}
	return EffectiveOf(c).Name
}

// stampPublicNamesLocked records, on ev, the public name of every card
// it references that is still findable. A name already on the event
// (a caller that captured it before removing the object) is kept.
func (g *Game) stampPublicNamesLocked(ev *Event) {
	for _, id := range [...]uuid.UUID{ev.CardID, ev.Target, ev.Source} {
		if id == uuid.Nil {
			continue
		}
		if _, ok := ev.PublicNames[id]; ok {
			continue
		}
		z, pos := g.locateCardLocked(id)
		if z == nil || pos < 0 || pos >= len(z.Cards) {
			continue
		}
		name := g.publicNameOfLocked(&z.Cards[pos])
		if name == "" {
			continue
		}
		if ev.PublicNames == nil {
			ev.PublicNames = make(map[uuid.UUID]string, 2)
		}
		ev.PublicNames[id] = name
	}
}
