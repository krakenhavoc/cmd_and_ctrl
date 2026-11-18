package game

import "github.com/google/uuid"

// SetNotedNumberForEffect stores n on the battlefield permanent `id` in
// Card.ChosenNumber WITHOUT the clamp at zero that SetChosenNumberForEffect
// applies, for "note your life total" (Sigarda's Splendor): a life total
// may be negative, and "greater than or equal to the last noted life total"
// must compare against the number that was noted. The field is already
// per-instance, snapshotted, cleared when the permanent leaves the
// battlefield (CR 400.7) and not a copiable value; a negative value
// round-trips as it is a plain JSON integer. A permanent that has left
// stores nothing.
//
// Caller must hold g.mu.
func (g *Game) SetNotedNumberForEffect(id uuid.UUID, n int) {
	if c := findBattlefieldCard(g, id); c != nil {
		c.ChosenNumber = n
	}
}
