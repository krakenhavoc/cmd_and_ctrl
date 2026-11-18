package game

import "github.com/google/uuid"

// ExileCardsFaceDownUnseenKeyedForEffect exiles the cards `ids` an effect
// has already chosen (from a library, say) face down, so that nobody may
// look at them (CR 406.3) until a cast permission later names them: they
// land as FaceDownPermitted with no permission yet, exactly as
// ExileTopFaceDownUnseenKeyedForEffect's do. The Dragon-Kami Reborn's
// "exile one of them face down with a hatching counter on it" is the shape
// (S82): the counter is the keyed tail's to place once the card has landed.
//
// `then` is told the cards that landed in exile once the batch is done; it
// also runs, told nothing, when `ids` is empty or nothing can be exiled.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileCardsFaceDownUnseenKeyedForEffect(actor uuid.UUID, ids []uuid.UUID, then KeyedTail) error {
	if g.Exile == nil || len(ids) == 0 {
		return then.Run(g, nil)
	}
	return g.routeAllKeyedLocked(zoneRoute{Dst: ZoneExile, Actor: actor, FaceDown: FaceDownPermitted}, append([]uuid.UUID(nil), ids...), then)
}
