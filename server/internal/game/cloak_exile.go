package game

import "github.com/google/uuid"

// cloak_exile.go — cloaking cards from exile one at a time, and the shuffled
// pile (Expose the Culprit, #4469).

// CloakFromExileForEffect cloaks the cards `ids`, all in exile, in the
// order given, one at a time (CR 701.58e), each entering face down under
// playerID's control. It returns the permanents that entered; a card a
// replacement kept out stays in exile, face up (CR 701.58f).
//
// Caller must hold g.mu.
func (g *Game) CloakFromExileForEffect(playerID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
	var out []uuid.UUID
	for _, id := range ids {
		if z := g.findCardZoneLocked(id); z == nil || z.Kind != ZoneExile {
			continue
		}
		entered, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{id}, ZoneExile, ZoneEntryOptions{
			Controller: playerID,
			FaceDown:   FaceDownCloaked,
		})
		if err != nil {
			return out, err
		}
		out = append(out, entered...)
	}
	return out, nil
}

// CloakShuffledFromExileForEffect is "exile … in a face-down pile, shuffle
// that pile, then cloak them" (Expose the Culprit): the cards in `ids`, all
// in exile, are put in a random order from the player's shuffle stream
// before CloakFromExileForEffect cloaks them one at a time, so nobody can
// tell which cloaked creature is which card.
//
// Caller must hold g.mu.
func (g *Game) CloakShuffledFromExileForEffect(playerID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
	pile := append([]uuid.UUID(nil), ids...)
	if len(pile) > 1 {
		r := g.randForLocked(rngStream{kind: rngStreamShuffle, player: playerID})
		r.Shuffle(len(pile), func(i, j int) { pile[i], pile[j] = pile[j], pile[i] })
	}
	return g.CloakFromExileForEffect(playerID, pile)
}
