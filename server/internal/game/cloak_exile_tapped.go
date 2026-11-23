package game

import "github.com/google/uuid"

// cloak_exile_tapped.go — "cloak them. They enter tapped" (Become
// Anonymous, S82): the shuffled pile of cloak_exile.go, each cloaked card
// entering tapped (CR 701.58a, CR 614.1d).

// CloakShuffledTappedFromExileForEffect is CloakShuffledFromExileForEffect
// with every cloaked permanent entering tapped. It returns the permanents
// that entered.
//
// Caller must hold g.mu.
func (g *Game) CloakShuffledTappedFromExileForEffect(playerID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
	pile := append([]uuid.UUID(nil), ids...)
	if len(pile) > 1 {
		r := g.randForLocked(rngStream{kind: rngStreamShuffle, player: playerID})
		r.Shuffle(len(pile), func(i, j int) { pile[i], pile[j] = pile[j], pile[i] })
	}
	var out []uuid.UUID
	for _, id := range pile {
		if z := g.findCardZoneLocked(id); z == nil || z.Kind != ZoneExile {
			continue
		}
		entered, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{id}, ZoneExile, ZoneEntryOptions{
			Controller: playerID,
			FaceDown:   FaceDownCloaked,
			Tapped:     true,
		})
		if err != nil {
			return out, err
		}
		out = append(out, entered...)
	}
	return out, nil
}
