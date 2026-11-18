package game

import "github.com/google/uuid"

// ExileCardsFaceDownChosenKeyedForEffect exiles cards the effect has
// already CHOSEN face down, with no permission yet: "exile one of them face
// down" (Discover the Impossible, #5493). They land as FaceDownPermitted,
// so a grant made over them afterwards makes its holder their one viewer
// (stampPermittedViewersLocked), as ExileTopFaceDownUnseenKeyedForEffect's
// cards. `then` is told the cards that landed in exile once the batch is
// done, and runs told nothing when there was nothing to exile.
//
// Caller must hold g.mu (write).
func (g *Game) ExileCardsFaceDownChosenKeyedForEffect(actor uuid.UUID, ids []uuid.UUID, then KeyedTail) error {
	if g.Exile == nil || len(ids) == 0 {
		return then.Run(g, nil)
	}
	return g.routeAllKeyedLocked(zoneRoute{Dst: ZoneExile, Actor: actor, FaceDown: FaceDownPermitted}, append([]uuid.UUID(nil), ids...), then)
}
