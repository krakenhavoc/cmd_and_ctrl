package game

import "github.com/google/uuid"

// ExileHandFaceDownKeyedForEffect exiles every card in `playerID`'s hand
// face down, the player as the one who exiles them, as one simultaneous
// event: "each player exiles all cards from their hand face down" (Memory
// Jar, #5841).
//
// The cards land FaceDownExiled, Necropotence's kind: CR 406.3 lets no
// player examine a card exiled face down "except when instructions allow
// it", and nothing here allows it, so no seat is a knower, the owner who
// held the card a moment ago included. The landing replaces the knowledge
// set rather than keeping the hand's, which is the point.
//
// `then` (a registered tail) is told the cards that reached exile, in hand
// order, as TailOutcome.IDs, once the whole batch is done: at once when
// nothing paused, and from the answer when a leg paused on a CR 616 prompt,
// so that prompt is a restore point. It also runs, told nothing, for an
// empty hand and for a player who is not found. A card a replacement sent
// somewhere else is not in the list (CR 400.7): it was not exiled this way.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileHandFaceDownKeyedForEffect(playerID uuid.UUID, then KeyedTail) error {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Hand == nil || g.Exile == nil || p.Hand.Size() == 0 {
		return then.Run(g, nil)
	}
	ids := make([]uuid.UUID, len(p.Hand.Cards))
	for i, c := range p.Hand.Cards {
		ids[i] = c.InstanceID
	}
	return g.routeAllKeyedLocked(zoneRoute{Dst: ZoneExile, Actor: playerID, FaceDown: FaceDownExiled}, ids, then)
}
