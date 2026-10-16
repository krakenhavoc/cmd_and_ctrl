package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counters_put_this_turn.go — "if you put a counter on a creature this
// turn" (Lord Jyscal Guado) and "if a counter was put on ~ this turn"
// (Wakka, Devoted Guardian), read back off this turn's event log
// (Game.EventsThisTurn, #3246).
//
// The engine emits one EventCounterPlaced per kind per permanent with
// the count AFTER the change, so a placement is an event whose count
// went up (b12CounterTotalBefore). A removal, a cancel to zero and a
// placement of nothing are not placements.

// counterPlacementDelta is how many counters ev put on its permanent:
// zero for anything that was not a placement.
func counterPlacementDelta(ev game.Event, g *game.Game) int {
	if ev.Kind != game.EventCounterPlaced || ev.Target == uuid.Nil || ev.Label == "" {
		return 0
	}
	before := b12CounterTotalBefore(ev, g)
	if d := ev.Amount - before; d > 0 {
		return d
	}
	return 0
}

// counterPlacedBy reports whether `me` put the counters ev placed: the
// placer the event names (ADR 0056 Decision 5; every placement names
// one since #3942).
func counterPlacedBy(ev game.Event, me uuid.UUID, g *game.Game) bool {
	_, ok := b12CountersPlacedBy(ev, me, g, false)
	return ok
}

// wasACreatureWhenCountered reports whether the permanent ev names is a
// creature: as it is on the battlefield now, or as it last existed there
// if it has left since.
func wasACreatureWhenCountered(ev game.Event, g *game.Game) bool {
	if z := g.FindCardZoneForEffect(ev.Target); z != nil && z.Kind == game.ZoneBattlefield {
		c, ok := g.LookupCardForEffect(ev.Target)
		return ok && c.IsCreature()
	}
	if info, ok := g.LastKnownPermanentForEffect(ev.Target); ok {
		for _, t := range info.Characteristic.Types {
			if strings.EqualFold(t, "creature") {
				return true
			}
		}
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.IsCreature()
}

// YouPutACounterOnACreatureThisTurn is "if you put a counter on a
// creature this turn".
//
// Caller must hold g.mu.
func YouPutACounterOnACreatureThisTurn(g *game.Game, you uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventCounterPlaced && counterPlacedBy(ev, you, g) && wasACreatureWhenCountered(ev, g) {
			return true
		}
	}
	return false
}

// ACounterWasPutOnThisObjectThisTurn is "if a counter was put on ~ this
// turn", by anyone, on the permanent `id`. Counters put on an
// earlier object of the same card (before it left and came back) do not
// count (CR 400.7): the scan starts after the card's last departure from
// the battlefield this turn.
//
// Caller must hold g.mu.
func ACounterWasPutOnThisObjectThisTurn(g *game.Game, id uuid.UUID) bool {
	events := g.EventsThisTurn()
	for i := len(events) - 1; i >= 0; i-- {
		ev := events[i]
		if ev.Kind == game.EventLTB && ev.CardID == id {
			return false
		}
		if ev.Kind == game.EventCounterPlaced && ev.Target == id && counterPlacementDelta(ev, g) > 0 {
			return true
		}
	}
	return false
}
