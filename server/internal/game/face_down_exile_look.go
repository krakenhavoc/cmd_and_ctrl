package game

import "github.com/google/uuid"

// LookAtFaceDownExileForEffect lets `looker` look at the cards of `ids` that
// are in exile face down under FaceDownPermitted, with no cast permission
// naming them (ExileTopFaceDownUnseenKeyedForEffect's cards): "Look at the
// cards in each pile" (Hostile Negotiations). The looker becomes a knower of
// each, and, as CR 406.3 says, keeps looking until the card leaves exile.
// Nobody else is told. A card that is not in exile face down, or is a
// different kind of face-down exile (Necropotence's, which nobody may read),
// is left alone. It returns the ids it let the looker see.
//
// Caller must hold g.mu (write).
func (g *Game) LookAtFaceDownExileForEffect(looker uuid.UUID, ids []uuid.UUID) []uuid.UUID {
	if g.Exile == nil || g.playerByIDLocked(looker) == nil {
		return nil
	}
	var seen []uuid.UUID
	for _, id := range ids {
		for i := range g.Exile.Cards {
			c := &g.Exile.Cards[i]
			if c.InstanceID != id {
				continue
			}
			if c.FaceDown && c.FaceDownKind == FaceDownPermitted {
				c.AddKnower(looker)
				seen = append(seen, id)
			}
			break
		}
	}
	return seen
}
