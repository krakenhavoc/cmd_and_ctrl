package game

import "github.com/google/uuid"

// LookAtFaceDownCombatantsForEffect lets `looker` look at each face-down
// creature that is attacking or blocking as the effect resolves (Revealing
// Wind: "You may look at each face-down creature that's attacking or
// blocking"). It is the one-shot look of CR 708.5, not the standing grant of
// face_down_look.go: only the creatures in combat now are read, and the looker
// becomes a knower of each (nobody else is told). It returns the ids it showed.
//
// Caller must hold g.mu (write).
func (g *Game) LookAtFaceDownCombatantsForEffect(looker uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil || g.playerByIDLocked(looker) == nil {
		return nil
	}
	var seen []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.faceDownPermanent() || !IsCreaturePtr(c) || (c.AttackingTarget == uuid.Nil && c.BlockingTarget == uuid.Nil) {
			continue
		}
		if c.Controller == looker {
			continue // CR 708.5: its controller may already look
		}
		c.AddKnower(looker)
		seen = append(seen, c.InstanceID)
	}
	return seen
}
