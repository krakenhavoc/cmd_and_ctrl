package game

// cast_this_turn.go is the turn-bound half of CastProvenance.CastThisTurn
// (#6098): "permanents that were cast this turn" stops naming a permanent
// as the next turn begins.

// clearCastThisTurnLocked forgets CastProvenance.CastThisTurn on every
// permanent, phased out or not (CR 702.26d keeps a phased-out permanent's
// state), and on every last-known record, as a turn begins. Called beside
// clearGraveyardArrivalsLocked. Caller must hold g.mu.
func (g *Game) clearCastThisTurnLocked() {
	for _, zone := range []*Zone{g.Battlefield, g.PhasedOut} {
		if zone == nil {
			continue
		}
		for i := range zone.Cards {
			zone.Cards[i].Provenance.CastThisTurn = false
		}
	}
	for _, recs := range g.lastKnownPermanents {
		for i := range recs {
			recs[i].CastThisTurn = false
		}
	}
}
