package game

import "github.com/google/uuid"

// ClearGoadsForEffect ends every goad marker on the battlefield creature
// `cardID`, for a card effect that already holds the lock: "each creature
// you control is no longer goaded" (Serene Sleuth). A goad a continuous
// effect applies (an Aura that goads) is not a marker and stays. Reports
// false when there is no such permanent.
//
// Caller must hold g.mu.
func (g *Game) ClearGoadsForEffect(cardID uuid.UUID) bool {
	c, ok := g.battlefieldCardLocked(cardID)
	if !ok {
		return false
	}
	c.Goads = nil
	return true
}
