package game

import "github.com/google/uuid"

// ExileIfItWouldDieForEffect is "if this creature would die, exile it
// instead" granted to ONE permanent with no duration (CR 614.1a, 700.4,
// 611.2): Can't Stay Away's "It gains 'If this creature would die, exile
// it instead.'". It is ExileIfItWouldDieThisTurnForEffect's record with an
// indefinite duration pinned to the object (CR 400.7), so it outlasts the
// turn and ends when the permanent leaves the battlefield. `controller`
// controls the replacement effect. Registers nothing for a card that is not
// on the battlefield.
//
// Caller must hold g.mu (write).
func (g *Game) ExileIfItWouldDieForEffect(sourceID, cardID, controller uuid.UUID, label string) bool {
	affected := g.PinnedObjectsLocked(cardID)
	if len(affected) == 0 {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, controller,
		[]Mod{{Kind: ModExileIfWouldDie}}, g.PinnedTo(IndefiniteDuration(), cardID),
		label, timeNowUnixNano())
}
