package game

import "github.com/google/uuid"

// exile_until_hits.go — S82: "exile cards from the top of your library
// until you exile two nonland cards with mana value 4 or less" (Invasion of
// Alara). exileUntilLocked's walk, run for N hits instead of one.

// ExileUntilHitsForEffect exiles cards from the top of `player`'s library,
// face up, until `n` of them satisfy `hit` or the library runs out. It
// returns the cards passed over (in exile order, none of which is a hit)
// and the hits (in exile order). Each card is marked known as it lands and
// each move is announced, as cascade's walk does.
//
// `hit` is called during the walk only and never stored. Caller must hold
// g.mu (write).
func (g *Game) ExileUntilHitsForEffect(player uuid.UUID, n int, hit func(Card) bool) (pile, hits []uuid.UUID, err error) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Library == nil || g.Exile == nil || n <= 0 {
		return nil, nil, nil
	}
	for len(hits) < n && p.Library.Size() > 0 {
		// The library's top is the LAST element.
		top := p.Library.Cards[len(p.Library.Cards)-1]
		id := top.InstanceID
		if _, err := MoveCard(p.Library, g.Exile, id); err != nil {
			return pile, hits, err
		}
		g.markCardKnownInZoneLocked(g.Exile, id)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   player,
			CardID:  id,
			OldZone: ZoneLibrary,
			NewZone: ZoneExile,
		})
		if hit(top) {
			hits = append(hits, id)
		} else {
			pile = append(pile, id)
		}
	}
	return pile, hits, nil
}
