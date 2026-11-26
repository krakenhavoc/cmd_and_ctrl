package game

// starting_player_conspiracy.go — "You are the starting player. If multiple
// players would be the starting player, one of those players is chosen at
// random." (Power Play, #5557, CR 103.1, 315.5).
//
// A conspiracy is in the command zone, face up, by the time the deal begins
// (placeConspiraciesLocked), and a face-up conspiracy's abilities function
// (CR 315.5). The deal reads them once, before the companions are revealed
// and the hands are dealt: the owners of such conspiracies are the candidates,
// and a table with several picks one at random. A table with none keeps the
// seat the opening roll chose. A hidden agenda conspiracy is face down here and
// has no abilities (CR 315.5b), so only a face-up one counts.

// conspiracyStartingSeatLocked is the seat that takes the first turn: the
// chosen seat of a table with no such conspiracy, else the seat of one of the
// players whose face-up conspiracy says they are the starting player.
//
// Caller must hold g.mu (write).
func (g *Game) conspiracyStartingSeatLocked(chosen int) int {
	var seats []int
	for i, p := range g.Seats {
		if p == nil || p.Eliminated || p.Emblems == nil {
			continue
		}
		for _, c := range p.Emblems.Cards {
			if !IsConspiracy(c) || c.FaceDown {
				continue
			}
			if d := catalogDef(CatalogKey(c)); d != nil && d.StartingPlayer {
				seats = append(seats, i)
				break
			}
		}
	}
	switch len(seats) {
	case 0:
		return chosen
	case 1:
		return seats[0]
	}
	return seats[g.randForLocked(rngStream{kind: rngStreamPick}).IntN(len(seats))]
}
