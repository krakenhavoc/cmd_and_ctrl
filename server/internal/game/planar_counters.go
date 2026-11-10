package game

import "github.com/google/uuid"

// planar_counters.go — counters on a face-up plane (ADR 0216 slice 4,
// #4870). Aplan Mortarium's "At the beginning of your upkeep, put an
// exposure counter on Aplan Mortarium. Then you lose life equal to the
// number of exposure counters on it." puts counters on a card in the
// command zone (CR 901.4). The planar containers are outside the card
// index (ADR 0216 §3), so AddCounterForEffect, which finds a card by
// its zone, never reaches one; this is the door for the one plane that
// counts on itself.
//
// The counters are plain Card.Counters on the face-up card, carried by
// the snapshot with the planar zones like any card field. A plane
// turned face down by a planeswalk is a new object (CR 901.7a), so
// planeswalkAwayFromLocked clears them. No counter event is emitted:
// every "counters are put on" reader of the catalog watches a permanent
// or a player, and a plane is neither (CR 311.2).

// PutCountersOnFaceUpPlanarCardForEffect puts `n` counters of `kind` on
// the face-up planar card `id` and reports how many of that kind it has
// now. ok is false, and nothing is put, when no face-up planar card has
// that ID (it was planeswalked away from: a new object, CR 901.7a).
//
// Caller must hold g.mu (write).
func (g *Game) PutCountersOnFaceUpPlanarCardForEffect(id uuid.UUID, kind string, n int) (int, bool) {
	up := g.planarFaceUpLocked()
	if up == nil || kind == "" {
		return 0, false
	}
	for i := range up.Cards {
		c := &up.Cards[i]
		if c.InstanceID != id {
			continue
		}
		if n > 0 {
			if c.Counters == nil {
				c.Counters = map[string]int{}
			}
			c.Counters[kind] += n
		}
		return c.Counters[kind], true
	}
	return 0, false
}
