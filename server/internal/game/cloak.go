package game

import "github.com/google/uuid"

// cloak.go — cloak (CR 701.58), the keyword action: manifest's sibling,
// a card put onto the battlefield face down as a 2/2 creature with ward
// {2} (FaceDownCloaked carries the ward, CR 702.168a via CR 701.58a).
// Cryptic Coat cloaks the top card of a library; Expose the Culprit cloaks
// cards it exiled.

// CloakForEffect is "cloak the top card of <player>'s library" (CR
// 701.58a): the top card enters under the player's control face down as a
// 2/2 with ward {2}, through the CR 614 entry pipeline, and the new
// permanent's ID is returned. An empty library cloaks nothing and is
// neither an error nor a draw (uuid.Nil, nil).
//
// Caller must hold g.mu.
func (g *Game) CloakForEffect(playerID uuid.UUID) (uuid.UUID, error) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return uuid.Nil, ErrPlayerNotFound
	}
	if p.Library.Size() == 0 {
		return uuid.Nil, nil
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	entered, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{top}, ZoneLibrary, ZoneEntryOptions{
		Controller: playerID,
		FaceDown:   FaceDownCloaked,
	})
	if err != nil || len(entered) == 0 {
		return uuid.Nil, err
	}
	return entered[0], nil
}

// CloakCardsKeyedForEffect is "cloak <cards>" for cards already in a
// public zone, such as the exile (Expose the Culprit): every entry enters
// face down as a cloaked 2/2 under `controller` as ONE simultaneous entry
// (CR 603.6a), and `then`, a registered tail, is told the permanents that
// entered as TailOutcome.IDs.
//
// Caller must hold g.mu.
func (g *Game) CloakCardsKeyedForEffect(entries []BatchEntry, controller uuid.UUID, then KeyedTail) error {
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries,
		ZoneEntryOptions{Controller: controller, FaceDown: FaceDownCloaked}, then)
}
