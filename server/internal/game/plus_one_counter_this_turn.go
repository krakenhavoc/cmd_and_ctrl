package game

import "github.com/google/uuid"

// plus_one_counter_this_turn.go — "if a +1/+1 counter was put on a
// permanent under your control this turn" (Fairgrounds Trumpeter, #5908).
//
// Read off the turn's event log: a placement (not a removal) of a +1/+1
// counter whose permanent is on the battlefield under `player`'s control
// now. Who put it is not asked, and the counters a permanent entered with
// count (CR 122.6). A permanent that has since left, or changed hands, no
// longer counts: weaker than "as it was" and never stronger. The log only
// grows within a turn.

// PlusOneCounterPutOnYourPermanentThisTurnForEffect reports whether a +1/+1
// counter was put this turn on a permanent `player` controls. Caller must
// hold g.mu.
func (g *Game) PlusOneCounterPutOnYourPermanentThisTurnForEffect(player uuid.UUID) bool {
	if player == uuid.Nil {
		return false
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != EventCounterPlaced || ev.Label != CounterPlusOne || ev.Target == uuid.Nil {
			continue
		}
		if ev.Amount-g.CounterCountBeforeForEffect(ev) <= 0 {
			continue
		}
		if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != ZoneBattlefield {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.Target); ok && c.Controller == player {
			return true
		}
	}
	return false
}
