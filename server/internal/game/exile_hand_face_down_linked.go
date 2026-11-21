package game

import "github.com/google/uuid"

// ExileHandFaceDownLinkedKeyedForEffect is ExileHandFaceDownKeyedForEffect
// for "exile all cards from your hand face down" on a card whose other
// ability reads those cards back as "cards you own exiled with this
// artifact" (Bottled Cloister, CR 607.2a): the same one batch, with every
// card that reaches exile linked to the source OBJECT `source` as it is now
// (Card.ExiledWith, read by ExiledWithForEffect). A zero source exiles with
// no link.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileHandFaceDownLinkedKeyedForEffect(source ObjectRef, playerID uuid.UUID, then KeyedTail) error {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Hand == nil || g.Exile == nil || p.Hand.Size() == 0 {
		return then.Run(g, nil)
	}
	ids := make([]uuid.UUID, len(p.Hand.Cards))
	for i, c := range p.Hand.Cards {
		ids[i] = c.InstanceID
	}
	return g.routeAllKeyedLocked(zoneRoute{
		Dst:        ZoneExile,
		Actor:      playerID,
		FaceDown:   FaceDownExiled,
		ExiledWith: PermissionCardRef(source),
	}, ids, then)
}
