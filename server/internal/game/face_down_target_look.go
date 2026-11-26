package game

import "github.com/google/uuid"

// face_down_target_look.go — S82 OLD batch 350 (#5554): "Look at target
// face-down creature" (Smoke Teller).
//
// A one-shot look, unlike face_down_look.go's standing grant: the looker
// becomes a knower of the creature (Card.KnownBy, the path a controller's
// own look at a face-down creature takes, CR 708.5), so the card shows to
// them in their own view and to nobody else. Nothing is announced to the
// table. They keep what they saw until the creature is turned face up (it
// is public then) or leaves the battlefield (CR 400.7 makes it a new
// object); that is a player's memory, not a standing permission.

// LookAtFaceDownCreatureForEffect lets `looker` look at the face-down
// creature `id` on the battlefield. It reports whether the look was made:
// false for a seat that is not in the game, for an object that is not a
// face-down permanent, and for one that is not a creature (a face-down
// card that is only a land, for instance, which is not "a face-down
// creature").
//
// Caller must hold g.mu (write).
func (g *Game) LookAtFaceDownCreatureForEffect(looker, id uuid.UUID) bool {
	if p := g.playerByIDLocked(looker); p == nil || p.Eliminated || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != id {
			continue
		}
		if !c.faceDownPermanent() || !IsCreaturePtr(c) {
			return false
		}
		c.AddKnower(looker)
		return true
	}
	return false
}
